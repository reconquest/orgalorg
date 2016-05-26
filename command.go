package main

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/seletskiy/hierr"
)

func runCommand(
	lockedNodes *distributedLock,
	command []string,
	verbosityLevel verbosity,
) error {
	commandString := joinCommand(command)

	remoteCommands := map[*distributedLockNode]*remoteExecutionNode{}

	logMutex := &sync.Mutex{}

	for _, node := range lockedNodes.nodes {
		tracef(
			"%s",
			hierr.Errorf(
				commandString,
				"%s starting command",
				node.String(),
			).Error(),
		)

		remoteCommand, err := node.runner.Command(
			commandString,
		)
		if err != nil {
			return hierr.Errorf(
				err,
				`can't create remote command`,
			)
		}

		var stdout io.WriteCloser
		var stderr io.WriteCloser
		switch verbosityLevel {
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

		err = remoteCommand.Start()
		if err != nil {
			return hierr.Errorf(
				err,
				`can't start remote command`,
			)
		}

		remoteCommands[node] = &remoteExecutionNode{
			node:    node,
			command: remoteCommand,

			stdout: stdout,
			stderr: stderr,
		}
	}

	for node, remoteCommand := range remoteCommands {
		err := remoteCommand.command.Wait()
		_ = remoteCommand.stdout.Close()
		_ = remoteCommand.stderr.Close()
		if err != nil {
			return hierr.Errorf(
				err,
				`%s can't wait for remote command to finish: '%s'`,
				node.String(),
				commandString,
			)
		}
	}

	return nil
}

func joinCommand(command []string) string {
	escapedParts := []string{}

	for _, part := range command {
		part = strings.Replace(part, `"`, `\"`, -1)
		part = strings.Replace(part, `\`, `\\`, -1)

		escapedParts = append(escapedParts, part)
	}

	return strings.Join(escapedParts, " ")
}
