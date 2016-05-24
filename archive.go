package main

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/crypto/ssh"

	"github.com/seletskiy/hierr"
	"github.com/theairkit/runcmd"
)

type archiveReceiverNode struct {
	node    distributedLockNode
	command runcmd.CmdWorker
}

type archiveReceivers struct {
	stdin io.WriteCloser
	nodes []archiveReceiverNode
}

func (receivers *archiveReceivers) wait() error {
	err := receivers.stdin.Close()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't close archive stream`,
		)
	}

	for _, receiver := range receivers.nodes {
		err := receiver.command.Wait()
		if err != nil {
			if sshErr, ok := err.(*ssh.ExitError); ok {
				return fmt.Errorf(
					`%s failed to receive archive, `+
						`remote command exited with non-zero code: %d`,
					receiver.node.String(),
					sshErr.Waitmsg.ExitStatus(),
				)
			} else {
				return hierr.Errorf(
					err,
					`%s failed to receive archive, unexpected error`,
					receiver.node.String(),
				)
			}
		}
	}

	return nil
}

func startArchiveReceivers(
	lockedNodes *distributedLock,
	args map[string]interface{},
) (*archiveReceivers, error) {
	var (
		rootDir = args["--root"].(string)
	)

	archiveReceiverCommandString := fmt.Sprintf(
		`tar -x --verbose --directory="%s"`,
		rootDir,
	)

	unpackers := []io.WriteCloser{}

	nodes := []archiveReceiverNode{}

	for _, node := range lockedNodes.nodes {
		debugf(hierr.Errorf(
			archiveReceiverCommandString,
			"%s starting archive receiver command",
			node.String(),
		).Error())

		archiveReceiverCommand, err := node.runner.Command(
			archiveReceiverCommandString,
		)

		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't create archive receiver command`,
			)
		}

		stdin, err := archiveReceiverCommand.StdinPipe()
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't get stdin from archive receiver command`,
			)
		}

		unpackers = append(unpackers, stdin)

		stdout := newLineFlushWriter(
			newPrefixWriter(
				newDebugWriter(logger),
				fmt.Sprintf("%s {tar} <stdout> ", node.String()),
			),
		)

		archiveReceiverCommand.SetStdout(stdout)

		stderr := newLineFlushWriter(
			newPrefixWriter(
				newDebugWriter(logger),
				fmt.Sprintf("%s {tar} <stderr> ", node.String()),
			),
		)

		archiveReceiverCommand.SetStderr(stderr)

		err = archiveReceiverCommand.Start()
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't start archive receiver command`,
			)
		}

		nodes = append(nodes, archiveReceiverNode{
			node:    node,
			command: archiveReceiverCommand,
		})
	}

	return &archiveReceivers{
		stdin: multiWriteCloser{unpackers},
		nodes: nodes,
	}, nil
}

func archiveFilesToWriter(target io.Writer, files []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't get current working directory`,
		)
	}

	archive := tar.NewWriter(target)
	for fileIndex, fileName := range files {
		fileInfo, err := os.Stat(fileName)

		if err != nil {
			return hierr.Errorf(
				err,
				`can't stat file for archiving: '%s`, fileName,
			)
		}

		// avoid tar warnings about leading slash
		tarFileName := fileName
		if tarFileName[0] == '/' {
			tarFileName = tarFileName[1:]

			fileName, err = filepath.Rel(workDir, fileName)
			if err != nil {
				return hierr.Errorf(
					err,
					`can't make relative path from: '%s'`,
					fileName,
				)
			}
		}

		header := &tar.Header{
			Name: tarFileName,
			Mode: int64(fileInfo.Sys().(*syscall.Stat_t).Mode),
			Size: fileInfo.Size(),

			Uid: int(fileInfo.Sys().(*syscall.Stat_t).Uid),
			Gid: int(fileInfo.Sys().(*syscall.Stat_t).Gid),

			ModTime: fileInfo.ModTime(),
		}

		logger.Infof(
			"(%d/%d) sending file: '%s'",
			fileIndex+1,
			len(files),
			fileName,
		)

		debugf(
			hierr.Errorf(
				fmt.Sprintf(
					"size: %d bytes; mode: %o; uid/gid: %d/%d; modtime: %s",
					header.Size,
					header.Mode,
					header.Uid,
					header.Gid,
					header.ModTime,
				),
				`local file: %s; remote file: %s`,
				fileName,
				tarFileName,
			).Error(),
		)

		err = archive.WriteHeader(header)

		if err != nil {
			return hierr.Errorf(
				err,
				`can't write tar header for fileName: '%s'`, fileName,
			)
		}

		fileToArchive, err := os.Open(fileName)
		if err != nil {
			return hierr.Errorf(
				err,
				`can't open fileName for reading: '%s'`,
				fileName,
			)
		}

		_, err = io.Copy(archive, fileToArchive)
		if err != nil {
			return hierr.Errorf(
				err,
				`can't copy file to the archive: '%s'`,
				fileName,
			)
		}
	}

	debugf("closing archive stream, %d files sent", len(files))

	err = archive.Close()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't close tar stream`,
		)
	}

	return nil
}

func getFilesList(relative bool, sources ...string) ([]string, error) {
	files := []string{}

	for _, source := range sources {
		err := filepath.Walk(
			source,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				if !relative {
					path, err = filepath.Abs(path)
					if err != nil {
						return hierr.Errorf(
							err,
							`can't get absolute path for local file: '%s'`,
							path,
						)
					}
				}

				files = append(files, path)

				return nil
			},
		)

		if err != nil {
			return nil, err
		}
	}

	return files, nil
}
