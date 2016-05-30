package main

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/seletskiy/hierr"
)

func runRemoteExecution(
	lockedNodes *distributedLock,
	command []string,
) (*remoteExecution, error) {
	var (
		stdins      = []io.WriteCloser{}
		remoteNodes = map[*distributedLockNode]*remoteExecutionNode{}

		commandString = joinCommand(command)

		logMutex      = &sync.Mutex{}
		nodesMapMutex = &sync.Mutex{}
	)

	errors := make(chan error, 0)
	for _, node := range lockedNodes.nodes {
		go func(node *distributedLockNode) {
			tracef(
				"%s",
				hierr.Errorf(
					commandString,
					"%s starting command",
					node.String(),
				).Error(),
			)

			remoteNode, err := runRemoteExecutionNode(
				node,
				commandString,
				logMutex,
			)
			if err != nil {
				errors <- err
				return
			}

			nodesMapMutex.Lock()
			{
				stdins = append(stdins, remoteNode.stdin)
				remoteNodes[node] = remoteNode
			}
			nodesMapMutex.Unlock()

			errors <- nil
		}(node)
	}

	for range lockedNodes.nodes {
		err := <-errors
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't run remote command on node`,
			)
		}
	}

	return &remoteExecution{
		stdin: multiWriteCloser{stdins},

		nodes: remoteNodes,
	}, nil
}

func runRemoteExecutionNode(
	node *distributedLockNode,
	command string,
	logMutex *sync.Mutex,
) (*remoteExecutionNode, error) {
	remoteCommand, err := node.runner.Command(command)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't create remote command`,
		)
	}

	var stdout io.WriteCloser
	var stderr io.WriteCloser
	switch verbose {
	default:
		stdout = newLineFlushWriter(
			logMutex,
			newPrefixWriter(
				os.Stdout,
				node.address.domain+" ",
			),
			true,
		)

		stderr = newLineFlushWriter(
			logMutex,
			newPrefixWriter(
				os.Stderr,
				node.address.domain+" ",
			),
			true,
		)

	case verbosityQuiet:
		stdout = newLineFlushWriter(logMutex, os.Stdout, false)
		stderr = newLineFlushWriter(logMutex, os.Stderr, false)

	case verbosityDebug:
		stdout = newLineFlushWriter(
			logMutex,
			newPrefixWriter(
				newDebugWriter(logger),
				node.String()+" {cmd} <stdout> ",
			),
			false,
		)

		stderr = newLineFlushWriter(
			logMutex,
			newPrefixWriter(
				newDebugWriter(logger),
				node.String()+" {cmd} <stderr> ",
			),
			false,
		)
	}

	remoteCommand.SetStdout(stdout)
	remoteCommand.SetStderr(stderr)

	stdin, err := remoteCommand.StdinPipe()
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't get stdin from archive receiver command`,
		)
	}

	err = remoteCommand.Start()
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't start remote command`,
		)
	}

	return &remoteExecutionNode{
		node:    node,
		command: remoteCommand,

		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}, nil
}

func joinCommand(command []string) string {
	escapedParts := []string{}

	for _, part := range command {
		part = strings.Replace(part, `\`, `\\`, -1)
		part = strings.Replace(part, ` `, `\ `, -1)

		escapedParts = append(escapedParts, part)
	}

	return strings.Join(escapedParts, " ")
}
