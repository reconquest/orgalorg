package main

import (
	"archive/tar"
	"io"
	"os"
	"syscall"

	"github.com/seletskiy/hierr"
)

func archiveFilesToStream(target io.Writer, files []string) error {
	archive := tar.NewWriter(target)
	for _, file := range files {
		fileInfo, err := os.Stat(file)

		if err != nil {
			return hierr.Errorf(
				err,
				`can't stat file for archieving: '%s`, file,
			)
		}

		err = archive.WriteHeader(&tar.Header{
			Name: file,
			Mode: int64(fileInfo.Sys().(*syscall.Stat_t).Mode),
			Size: fileInfo.Size(),

			Uid: int(fileInfo.Sys().(*syscall.Stat_t).Uid),
			Gid: int(fileInfo.Sys().(*syscall.Stat_t).Gid),

			ModTime: fileInfo.ModTime(),
		})

		if err != nil {
			return hierr.Errorf(
				err,
				`can't write tar header for file: '%s'`, file,
			)
		}
	}

	err := archive.Close()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't close tar stream`,
		)
	}

	return nil
}
