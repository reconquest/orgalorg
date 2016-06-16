package main

import (
	"fmt"
	"io"
	"strings"
)

type multiWriteCloser struct {
	writers []io.WriteCloser
}

func (closer *multiWriteCloser) Write(data []byte) (int, error) {
	errs := []string{}

	for _, writer := range closer.writers {
		_, err := writer.Write(data)
		if err != nil && err != io.EOF {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return 0, fmt.Errorf(
			"%d errors: %s",
			len(errs),
			strings.Join(errs, "; "),
		)
	}

	return len(data), nil
}

func (closer *multiWriteCloser) Close() error {
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
			strings.Join(errs, "; "),
		)
	}

	return nil
}
