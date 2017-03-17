package main

import (
	"io"
	"os"
	"sync"

	"github.com/reconquest/hierr-go"
	"github.com/reconquest/lineflushwriter-go"
	"github.com/reconquest/prefixwriter-go"
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
	command []string,
	setupCallback func(*remoteExecutionNode),
	serial bool,
) (*remoteExecution, error) {
	var (
		stdins = []io.WriteCloser{}

		logLock    = &sync.Mutex{}
		stdinsLock = &sync.Mutex{}
		outputLock = &sync.Mutex{}

		nodes = &remoteNodes{&sync.Mutex{}, remoteNodesMap{}}
	)

	if !serial {
		outputLock = nil
	}

	var (
		status = &struct {
			sync.Mutex

			Phase   string
			Total   int
			Fails   int
			Success int
		}{
			Phase: `exec`,
			Total: len(lockedNodes.nodes),
		}
	)

	setStatus(status)

	type nodeErr struct {
		err  error
		node *distributedLockNode
	}

	errors := make(chan *nodeErr, 0)
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
					outputLock,
				)
				if err != nil {
					errors <- &nodeErr{err, node}

					status.Lock()
					defer status.Unlock()

					status.Total--
					status.Fails++

					return
				}

				if setupCallback != nil {
					setupCallback(remoteNode)
				}

				remoteNode.command.SetStdout(remoteNode.stdout)
				remoteNode.command.SetStderr(remoteNode.stderr)

				err = remoteNode.command.Start()
				if err != nil {
					errors <- &nodeErr{
						hierr.Errorf(
							err,
							`can't start remote command`,
						),
						node,
					}

					status.Lock()
					defer status.Unlock()

					status.Total--
					status.Fails++

					return
				}

				nodes.Set(node, remoteNode)

				stdinsLock.Lock()
				defer stdinsLock.Unlock()

				stdins = append(stdins, remoteNode.stdin)

				status.Lock()
				defer status.Unlock()

				status.Success++

				errors <- nil
			})
		}(node)
	}

	for range lockedNodes.nodes {
		err := <-errors
		if err != nil {
			return nil, hierr.Errorf(
				err.err,
				`%s remote execution failed`,
				err.node,
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
	command []string,
	logLock sync.Locker,
	outputLock sync.Locker,
) (*remoteExecutionNode, error) {
	remoteCommand := node.runner.Command(command[0], command[1:]...)

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
				"{cmd} <stdout> "+node.String()+" ",
			),
			logLock,
			false,
		)

		stderr = lineflushwriter.New(
			prefixwriter.New(
				newDebugWriter(logger),
				"{cmd} <stderr> "+node.String()+" ",
			),
			logLock,
			false,
		)
	}

	stdout = &statusBarUpdateWriter{stdout}
	stderr = &statusBarUpdateWriter{stderr}

	if outputLock != (*sync.Mutex)(nil) {
		sharedLock := newSharedLock(outputLock, 2)

		stdout = newLockedWriter(stdout, sharedLock)
		stderr = newLockedWriter(stderr, sharedLock)
	}

	stdin, err := remoteCommand.StdinPipe()
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't get stdin from remote command`,
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
