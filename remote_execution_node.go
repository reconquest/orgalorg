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

	stdout io.WriteCloser
	stderr io.WriteCloser
}

func (node *remoteExecutionNode) wait() error {
	err := node.command.Wait()
	if err != nil {
		if sshErr, ok := err.(*ssh.ExitError); ok {
			return fmt.Errorf(
				`%s failed to evaluate command, `+
					`remote command exited with non-zero code: %d`,
				node.node.String(),
				sshErr.Waitmsg.ExitStatus(),
			)
		}

		return hierr.Errorf(
			err,
			`%s failed to receive archive, unexpected error`,
			node.node.String(),
		)
	}

	return nil
}
