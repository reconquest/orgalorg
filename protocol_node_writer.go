package main

import (
	"bufio"
	"bytes"
	"io"
)

type protocolNodeWriter struct {
	node     *remoteExecutionNode
	protocol *syncProtocol

	stdout io.Writer

	buffer *bytes.Buffer
}

func newProtocolNodeWriter(
	node *remoteExecutionNode,
	protocol *syncProtocol,
) *protocolNodeWriter {
	return &protocolNodeWriter{
		node:     node,
		stdout:   node.stdout,
		protocol: protocol,
		buffer:   &bytes.Buffer{},
	}
}

func (writer *protocolNodeWriter) Write(data []byte) (int, error) {
	written, err := writer.buffer.Write(data)
	if err != nil {
		return written, err
	}

	reader := bufio.NewReader(writer.buffer)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				_, err := io.WriteString(writer.buffer, line)
				if err != nil {
					return 0, err
				}

				break
			}
		}

		switch {
		case writer.protocol.IsSyncCommand(line):
			tracef(
				"%s sent sync command: '%s'",
				writer.node.String(),
				line,
			)

			err := writer.protocol.SendSync(writer.node, line)

			if err != nil {
				return 0, err
			}
		default:
			_, err := io.WriteString(writer.stdout, line)
			if err != nil {
				return 0, err
			}
		}
	}

	return written, nil
}

func (writer *protocolNodeWriter) Close() error {
	return nil
}
