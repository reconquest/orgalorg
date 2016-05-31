package main

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/reconquest/go-lineflushwriter"
	"github.com/reconquest/go-prefixwriter"
	"github.com/seletskiy/hierr"
)

func runRemoteExecution(
	lockedNodes *distributedLock,
	command []string,
	callback func(*remoteExecutionNode),
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

			if callback != nil {
				callback(remoteNode)
			}

			remoteNode.command.SetStdout(remoteNode.stdout)
			remoteNode.command.SetStderr(remoteNode.stderr)

			err = remoteNode.command.Start()
			if err != nil {
				errors <- hierr.Errorf(
					err,
					`can't start remote command`,
				)

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
		stdin: &multiWriteCloser{stdins},

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
	case verbosityQuiet:
		stdout = lineflushwriter.New(nopCloser{os.Stdout}, logMutex, false)
		stderr = lineflushwriter.New(nopCloser{os.Stderr}, logMutex, false)

	case verbosityNormal:
		stdout = lineflushwriter.New(
			prefixwriter.New(
				nopCloser{os.Stdout},
				node.address.domain+" ",
			),
			logMutex,
			true,
		)

		stderr = lineflushwriter.New(
			prefixwriter.New(
				nopCloser{os.Stderr},
				node.address.domain+" ",
			),
			logMutex,
			true,
		)

	default:
		fallthrough
	case verbosityDebug:
		stdout = lineflushwriter.New(
			prefixwriter.New(
				newDebugWriter(logger),
				node.String()+" {cmd} <stdout> ",
			),
			logMutex,
			false,
		)

		stderr = lineflushwriter.New(
			prefixwriter.New(
				newDebugWriter(logger),
				node.String()+" {cmd} <stderr> ",
			),
			logMutex,
			false,
		)
	}

	stdin, err := remoteCommand.StdinPipe()
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't get stdin from archive receiver command`,
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
