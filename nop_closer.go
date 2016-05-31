package main

import (
	"io"
)

type nopCloser struct {
	io.Writer
}

func (closer nopCloser) Close() error {
	return nil
}
