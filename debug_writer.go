package main

import (
	"strings"

	"github.com/kovetskiy/lorg"
)

type debugWriter struct {
	log *lorg.Log
}

func newDebugWriter(log *lorg.Log) debugWriter {
	return debugWriter{
		log: log,
	}
}

func (writer debugWriter) Write(data []byte) (int, error) {
	writer.log.Debug(strings.TrimSuffix(string(data), "\n"))

	return len(data), nil
}

func (writer debugWriter) Close() error {
	return nil
}
