package main

import "io"

type statusBarUpdateWriter struct {
	writer io.WriteCloser
}

func (writer *statusBarUpdateWriter) Write(data []byte) (int, error) {
	clearStatus()

	written, err := writer.writer.Write(data)

	drawStatus()

	return written, err
}

func (writer *statusBarUpdateWriter) Close() error {
	return writer.writer.Close()
}
