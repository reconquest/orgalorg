package main

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/reconquest/hierr-go"
	"github.com/reconquest/lineflushwriter-go"
	"github.com/reconquest/prefixwriter-go"
)

func startArchiveReceivers(
	cluster *distributedLock,
	rootDir string,
	sudo bool,
	serial bool,
) (*remoteExecution, error) {
	command := []string{
		"mkdir", "-p", rootDir, "&&", "tar", "--directory", rootDir, "-x",
	}

	if verbose >= verbosityDebug {
		command = append(command, `--verbose`)
	}

	logMutex := &sync.Mutex{}

	runner := &remoteExecutionRunner{
		command: command,
		serial:  serial,
		shell:   defaultRemoteExecutionShell,
		sudo:    sudo,
	}

	execution, err := runner.run(
		cluster,
		func(node *remoteExecutionNode) {
			node.stdout = lineflushwriter.New(
				prefixwriter.New(node.stdout, "{tar} "),
				logMutex,
				true,
			)

			node.stderr = lineflushwriter.New(
				prefixwriter.New(node.stderr, "{tar} "),
				logMutex,
				true,
			)
		},
	)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't start tar extraction command: '%v'`,
			command,
		)
	}

	return execution, nil
}

func archiveFilesToWriter(
	target io.WriteCloser,
	files []file,
	preserveUID, preserveGID bool,
) error {
	workDir, err := os.Getwd()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't get current working directory`,
		)
	}

	status := &struct {
		Phase   string
		Total   int
		Fails   int
		Success int
		Written bytesStringer
		Bytes   bytesStringer
	}{
		Phase: "upload",
		Total: len(files),
	}

	setStatus(status)

	for _, file := range files {
		status.Bytes.Amount += file.size
	}

	archive := tar.NewWriter(target)
	stream := io.MultiWriter(archive, callbackWriter(
		func(data []byte) (int, error) {
			status.Written.Amount += len(data)

			drawStatus()

			return len(data), nil
		},
	))

	for fileIndex, file := range files {
		infof(
			"%5d/%d sending file: '%s'",
			fileIndex+1,
			len(files),
			file.path,
		)

		err = writeFileToArchive(
			file.path,
			stream,
			archive,
			workDir,
			preserveUID,
			preserveGID,
		)
		if err != nil {
			return hierr.Errorf(
				err,
				`can't write file to archive: '%s'`,
				file.path,
			)
		}

		status.Success++
	}

	tracef("closing archive stream, %d files sent", len(files))

	err = archive.Close()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't close tar stream`,
		)
	}

	err = target.Close()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't close target stdin`,
		)
	}

	return nil
}

func getFilesList(relative bool, sources ...string) ([]file, error) {
	files := []file{}

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

				files = append(files, file{
					path: path,
					size: int(info.Size()),
				})

				return nil
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}
