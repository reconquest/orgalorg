package main

import (
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
	writer.log.Debug(string(data))

	return len(data), nil
}
