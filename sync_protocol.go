package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"
)

var (
	syncProtocolPrefix = "ORGALORG"
	syncProtocolHello  = "HELLO"
	syncProtocolNode   = "NODE"
	syncProtocolStart  = "START"
	syncProtocolSync   = "SYNC"
)

type syncProtocol struct {
	node *remoteExecutionNode

	input  *bytes.Buffer
	output io.WriteCloser

	prefix string
}

func newSyncProtocol() *syncProtocol {
	return &syncProtocol{
		input: &bytes.Buffer{},
		prefix: fmt.Sprintf(
			"%s:%d",
			syncProtocolPrefix,
			time.Now().UnixNano(),
		),
	}
}

func (protocol *syncProtocol) Close() error {
	return nil
}

func (protocol *syncProtocol) Init(output io.WriteCloser) error {
	protocol.output = output

	_, err := io.WriteString(
		protocol.output,
		protocol.prefix+" "+syncProtocolHello+"\n",
	)
	if err != nil {
		return protocolSuspendEOF(err)
	}

	return nil
}

func (protocol *syncProtocol) SendNode(node *remoteExecutionNode) error {
	_, err := io.WriteString(
		protocol.output,
		protocol.prefix+" "+syncProtocolNode+" "+node.String()+"\n",
	)
	if err != nil {
		return protocolSuspendEOF(err)
	}

	return nil
}

func (protocol *syncProtocol) SendStart() error {
	_, err := io.WriteString(
		protocol.output,
		protocol.prefix+" "+syncProtocolStart+"\n",
	)
	if err != nil {
		return protocolSuspendEOF(err)
	}

	return nil
}

func (protocol *syncProtocol) IsSyncCommand(line string) bool {
	return strings.HasPrefix(line, protocol.prefix+" "+syncProtocolSync)
}

func (protocol *syncProtocol) SendSync(
	source *remoteExecutionNode,
	sync string,
) error {
	data := strings.TrimSpace(
		strings.TrimPrefix(sync, protocol.prefix+" "+syncProtocolSync),
	)

	_, err := io.WriteString(
		protocol.output,
		protocol.prefix+" "+syncProtocolSync+" "+source.String()+" "+data+"\n",
	)

	if err != nil {
		return protocolSuspendEOF(err)
	}

	return nil
}

// Suspend EOF for be compatible with simple commands, that are not support
// protocol, and therefore can close exit earlier, than protocol is initiated.
func protocolSuspendEOF(err error) error {
	if err == io.EOF {
		return nil
	}

	return err
}
