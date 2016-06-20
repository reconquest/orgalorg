package main

import (
	"io"
	"os"
	"sync"

	"github.com/reconquest/go-lineflushwriter"
	"github.com/reconquest/go-prefixwriter"
	"github.com/seletskiy/hierr"
)

type remoteNodesMap map[*distributedLockNode]*remoteExecutionNode

type remoteNodes struct {
	*sync.Mutex

	nodes remoteNodesMap
}

func (nodes *remoteNodes) Set(
	node *distributedLockNode,
	remote *remoteExecutionNode,
) {
	nodes.Lock()
	defer nodes.Unlock()

	nodes.nodes[node] = remote
}

func runRemoteExecution(
	lockedNodes *distributedLock,
	command string,
	setupCallback func(*remoteExecutionNode),
) (*remoteExecution, error) {
	var (
		stdins = []io.WriteCloser{}

		logLock    = &sync.Mutex{}
		stdinsLock = &sync.Mutex{}

		nodes = &remoteNodes{&sync.Mutex{}, remoteNodesMap{}}
	)

	errors := make(chan error, 0)
	for _, node := range lockedNodes.nodes {
		go func(node *distributedLockNode) {
			pool.run(func() {
				tracef(
					"%s",
					hierr.Errorf(
						command,
						"%s starting command",
						node.String(),
					).Error(),
				)

				remoteNode, err := runRemoteExecutionNode(
					node,
					command,
					logLock,
				)
				if err != nil {
					errors <- err
					return
				}

				if setupCallback != nil {
					setupCallback(remoteNode)
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

				nodes.Set(node, remoteNode)

				stdinsLock.Lock()
				defer stdinsLock.Unlock()

				stdins = append(stdins, remoteNode.stdin)

				errors <- nil
			})
		}(node)
	}

	for _, node := range lockedNodes.nodes {
		err := <-errors
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`remote execution failed on node: '%s'`,
				node,
			)
		}
	}

	return &remoteExecution{
		stdin: &multiWriteCloser{stdins},

		nodes: nodes.nodes,
	}, nil
}

func runRemoteExecutionNode(
	node *distributedLockNode,
	command string,
	logLock *sync.Mutex,
) (*remoteExecutionNode, error) {
	remoteCommand, err := node.runner.Command(command)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't establish remote session`,
		)
	}

	stdoutBackend := io.Writer(os.Stdout)
	stderrBackend := io.Writer(os.Stderr)

	if format == outputFormatJSON {
		stdoutBackend = &jsonOutputWriter{
			stream: `stdout`,
			node:   node.String(),

			output: os.Stdout,
		}

		stderrBackend = &jsonOutputWriter{
			stream: `stderr`,
			node:   node.String(),

			output: os.Stderr,
		}
	}

	var stdout io.WriteCloser
	var stderr io.WriteCloser
	switch {
	case verbose == verbosityQuiet || format == outputFormatJSON:
		stdout = lineflushwriter.New(nopCloser{stdoutBackend}, logLock, false)
		stderr = lineflushwriter.New(nopCloser{stderrBackend}, logLock, false)

	case verbose == verbosityNormal:
		stdout = lineflushwriter.New(
			prefixwriter.New(
				nopCloser{stdoutBackend},
				node.address.domain+" ",
			),
			logLock,
			true,
		)

		stderr = lineflushwriter.New(
			prefixwriter.New(
				nopCloser{stderrBackend},
				node.address.domain+" ",
			),
			logLock,
			true,
		)

	default:
		stdout = lineflushwriter.New(
			prefixwriter.New(
				newDebugWriter(logger),
				node.String()+" {cmd} <stdout> ",
			),
			logLock,
			false,
		)

		stderr = lineflushwriter.New(
			prefixwriter.New(
				newDebugWriter(logger),
				node.String()+" {cmd} <stderr> ",
			),
			logLock,
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
