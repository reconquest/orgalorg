//go:build windows

package main

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/reconquest/hierr-go"
)

func writeFileToArchive(
	fileName string,
	stream io.Writer,
	archive *tar.Writer,
	workDir string,
	preserveUID, preserveGID bool,
) error {
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
		Size: fileInfo.Size(),

		ModTime: fileInfo.ModTime(),
	}

	tracef(
		hierr.Errorf(
			fmt.Sprintf(
				"size: %d bytes; modtime: %s",
				header.Size,
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

	_, err = io.Copy(stream, fileToArchive)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't copy file to the archive: '%s'`,
			fileName,
		)
	}

	return nil
}
