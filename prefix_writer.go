package main

import (
	"bytes"
	"io"
	"regexp"
)

type prefixWriter struct {
	writer io.Writer
	prefix string
}

func newPrefixWriter(writer io.Writer, prefix string) prefixWriter {
	return prefixWriter{
		writer: writer,
		prefix: prefix,
	}
}

func (writer prefixWriter) Write(data []byte) (int, error) {
	prefixedData := regexp.MustCompile(`(?m)^`).ReplaceAllLiteral(
		bytes.TrimRight(data, "\n"),
		[]byte(writer.prefix),
	)

	_, err := writer.writer.Write(prefixedData)
	if err != nil {
		return 0, err
	}

	return len(data), nil
}
