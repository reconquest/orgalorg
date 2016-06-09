package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/seletskiy/hierr"
)

type (
	outputFormat int
)

const (
	outputFormatText outputFormat = iota
	outputFormatJSON
)

func parseOutputFormat(args map[string]interface{}) outputFormat {
	if args["--json"].(bool) {
		return outputFormatJSON
	}

	return outputFormatText
}

type jsonOutputWriter struct {
	stream string
	node   string

	output io.Writer
}

func (writer *jsonOutputWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	message := map[string]interface{}{
		"stream": writer.stream,
	}

	if writer.node == "" {
		message["node"] = nil
	} else {
		message["node"] = writer.node
	}

	message["body"] = string(data)

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return 0, err
	}

	_, err = writer.output.Write(append(jsonMessage, '\n'))
	if err != nil {
		return 0, err
	}

	return len(data), nil
}

func serializeError(err error) string {
	if hierarchicalError, ok := err.(hierr.Error); ok {
		serializedError := fmt.Sprint(hierarchicalError.Nested)
		if nested, ok := hierarchicalError.Nested.(error); ok {
			serializedError = serializeError(nested)
		}

		return hierarchicalError.Message + ": " + serializedError
	}

	return err.Error()
}
