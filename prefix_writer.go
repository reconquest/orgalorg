package main

import (
	"bytes"
	"io"
)

type prefixWriter struct {
	writer io.Writer
	prefix string

	streamStarted  bool
	lineIncomplete bool
}

func newPrefixWriter(writer io.Writer, prefix string) *prefixWriter {
	return &prefixWriter{
		writer: writer,
		prefix: prefix,
	}
}

func (writer *prefixWriter) Write(data []byte) (int, error) {
	reader := bytes.NewBuffer(data)
	eof := false
	for !eof {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return 0, err
			} else {
				eof = true
			}
		}

		if line == "" {
			continue
		}

		if !writer.streamStarted {
			line = writer.prefix + line

			writer.streamStarted = true
		} else {
			if !writer.lineIncomplete {
				line = writer.prefix + line
			}

			if eof {
				writer.lineIncomplete = true
			}
		}

		_, err = writer.writer.Write([]byte(line))
		if err != nil {
			return 0, err
		}
	}

	return len(data), nil
}
