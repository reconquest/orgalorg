package main

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

type lineFlushWriter struct {
	mutex  *sync.Mutex
	writer io.Writer
	buffer *bytes.Buffer
}

func newLineFlushWriter(writer io.Writer) lineFlushWriter {
	return lineFlushWriter{
		writer: writer,
		mutex:  &sync.Mutex{},
		buffer: &bytes.Buffer{},
	}
}

func (writer lineFlushWriter) Write(data []byte) (int, error) {
	_, err := writer.buffer.Write(data)
	if err != nil {
		return 0, err
	}

	err = writer.Flush()
	if err != nil {
		return 0, err
	}

	return len(data), nil
}

func (writer lineFlushWriter) Flush() error {
	writer.mutex.Lock()
	defer writer.mutex.Unlock()

	if !bytes.Contains(writer.buffer.Bytes(), []byte("\n")) {
		return nil
	}

	scanner := bufio.NewScanner(writer.buffer)
	for scanner.Scan() {
		line := scanner.Text()

		_, err := writer.writer.Write([]byte(line + "\n"))
		if err != nil {
			return err
		}

		if !bytes.Contains(writer.buffer.Bytes(), []byte("\n")) {
			return nil
		}

	}

	return nil
}

func (writer lineFlushWriter) Close() error {
	return writer.Flush()
}
