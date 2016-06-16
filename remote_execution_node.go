package main

import (
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"

	"github.com/seletskiy/hierr"
	"github.com/theairkit/runcmd"
)

type remoteExecutionNode struct {
	node    *distributedLockNode
	command runcmd.CmdWorker

	stdin  io.WriteCloser
	stdout io.WriteCloser
	stderr io.WriteCloser
}

func (node *remoteExecutionNode) wait() error {
	err := node.command.Wait()
	if err != nil {
		_ = node.stdout.Close()
		_ = node.stderr.Close()
		if sshErr, ok := err.(*ssh.ExitError); ok {
			return fmt.Errorf(
				`%s had failed to evaluate command, `+
					`remote command exited with non-zero code: %d`,
				node.node.String(),
				sshErr.Waitmsg.ExitStatus(),
			)
		}

		return hierr.Errorf(
			err,
			`%s failed to finish execution, unexpected error`,
			node.node.String(),
		)
	}

	err = node.stdout.Close()
	if err != nil {
		return hierr.Errorf(
			err,
			`%s can't close stdout`,
			node.node.String(),
		)
	}

	err = node.stderr.Close()
	if err != nil {
		return hierr.Errorf(
			err,
			`%s can't close stderr`,
			node.node.String(),
		)
	}

	return nil
}

func (node *remoteExecutionNode) String() string {
	return node.node.String()
}
