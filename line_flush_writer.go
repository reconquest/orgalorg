package main

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"sync"
)

type lineFlushWriter struct {
	mutex  *sync.Mutex
	writer io.Writer
	buffer *bytes.Buffer

	newlineAtEnd bool
}

func newLineFlushWriter(
	mutex *sync.Mutex,
	writer io.Writer,
	newlineAtEnd bool,
) lineFlushWriter {
	return lineFlushWriter{
		writer: writer,
		mutex:  mutex,
		buffer: &bytes.Buffer{},

		newlineAtEnd: newlineAtEnd,
	}
}

func (writer lineFlushWriter) Write(data []byte) (int, error) {
	written, err := writer.buffer.Write(data)
	if err != nil {
		return written, err
	}

	var (
		reader = bufio.NewReader(writer.buffer)
		eof    = false
	)

	for !eof {
		line, err := reader.ReadString('\n')

		writer.mutex.Lock()

		if err != nil {
			if err != io.EOF {
				writer.mutex.Unlock()
				return 0, err
			} else {
				eof = true
			}
		}

		if eof {
			writer.buffer.Reset()
			written, err := writer.buffer.WriteString(line)
			writer.mutex.Unlock()
			if err != nil {
				return written, err
			}
		} else {
			written, err := writer.writer.Write([]byte(line))
			writer.mutex.Unlock()
			if err != nil {
				return written, err
			}
		}
	}

	return written, nil
}

func (writer lineFlushWriter) Close() error {
	if writer.newlineAtEnd && writer.buffer.Len() > 0 {
		if !strings.HasSuffix(writer.buffer.String(), "\n") {
			_, err := writer.buffer.WriteString("\n")
			if err != nil {
				return err
			}
		}
	}

	_, err := writer.writer.Write(writer.buffer.Bytes())
	return err
}
