package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/reconquest/prefixwriter-go"
)

var (
	syncProtocolPrefix      = "ORGALORG"
	syncProtocolHello       = "HELLO"
	syncProtocolNode        = "NODE"
	syncProtocolNodeCurrent = "CURRENT"
	syncProtocolStart       = "START"
	syncProtocolSync        = "SYNC"
)

// syncProtocol handles SYNC protocol described in the main.go.
//
// It will handle protocol over all connected nodes.
type syncProtocol struct {
	// output represents writer, that should be connected to stdins of
	// all connected nodes.
	output io.WriteCloser

	// prefix is a unique string which prefixes every protocol message.
	prefix string
}

// newSyncProtocol returns syncProtocol instantiated with unique prefix.
func newSyncProtocol() *syncProtocol {
	return &syncProtocol{
		prefix: fmt.Sprintf(
			"%s:%d",
			syncProtocolPrefix,
			time.Now().UnixNano(),
		),
	}
}

// Init starts protocol and sends HELLO message to the writer. Specified writer
// will be used in all further communications.
func (protocol *syncProtocol) Init(output io.WriteCloser) error {
	protocol.output = prefixwriter.New(output, protocol.prefix+" ")

	_, err := io.WriteString(
		protocol.output,
		syncProtocolHello+"\n",
	)
	if err != nil {
		return protocolSuspendEOF(err)
	}

	return nil
}

// SendNode sends to the writer serialized representation of specified node as
// NODE message.
func (protocol *syncProtocol) SendNode(
	current *remoteExecutionNode,
	neighbor *remoteExecutionNode,
) error {
	var line = syncProtocolNode + " " + neighbor.String()

	if current == neighbor {
		line += " " + syncProtocolNodeCurrent
	}

	_, err := io.WriteString(current.stdin, line+"\n")
	if err != nil {
		return protocolSuspendEOF(err)
	}

	return nil
}

// SendStart sends START message to the writer.
func (protocol *syncProtocol) SendStart() error {
	_, err := io.WriteString(
		protocol.output,
		syncProtocolStart+"\n",
	)
	if err != nil {
		return protocolSuspendEOF(err)
	}

	return nil
}

// IsSyncCommand will return true,  if specified line looks like incoming
// SYNC message from the remote node.
func (protocol *syncProtocol) IsSyncCommand(line string) bool {
	return strings.HasPrefix(line, protocol.prefix+" "+syncProtocolSync)
}

// SendSync sends SYNC message to the writer, tagging it as sent from node,
// described by given source and adding optional description for the given
// SYNC phase taken by extraction it from the original SYNC message, sent
// by node.
func (protocol *syncProtocol) SendSync(
	source fmt.Stringer,
	sync string,
) error {
	data := strings.TrimSpace(
		strings.TrimPrefix(sync, protocol.prefix+" "+syncProtocolSync),
	)

	_, err := io.WriteString(
		protocol.output,
		syncProtocolSync+" "+source.String()+" "+data+"\n",
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
