package main

import (
	"io"
)

type writerWrapper struct {
	writer io.WriteCloser
}

func (wrapper *writerWrapper) Write(data []byte) (int, error) {
	return wrapper.writer.Write(data)
}

func (wrapper *writerWrapper) Close() error {
	return wrapper.writer.Close()
}
