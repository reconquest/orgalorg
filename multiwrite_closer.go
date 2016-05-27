package main

import (
	"fmt"
	"io"
	"strings"
)

type multiWriteCloser struct {
	writers []io.WriteCloser
}

func (closer multiWriteCloser) Write(data []byte) (int, error) {
	writers := []io.Writer{}
	for _, writer := range closer.writers {
		writers = append(writers, writer)
	}

	return io.MultiWriter(writers...).Write(data)
}

func (closer multiWriteCloser) Close() error {
	errs := []string{}

	for _, closer := range closer.writers {
		err := closer.Close()
		if err != nil && err != io.EOF {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"%d errors: %s",
			len(errs),
			strings.Join(errs, ";"),
		)
	}

	return nil
}
