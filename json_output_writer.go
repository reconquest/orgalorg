package main

import (
	"encoding/json"
	"io"
)

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
