package main

import (
	"fmt"
	"io"

	"github.com/seletskiy/hierr"
	"golang.org/x/crypto/ssh"
)

type archiveReceivers struct {
	stdin io.WriteCloser
	nodes []archiveReceiverNode
}

func (receivers *archiveReceivers) wait() error {
	err := receivers.stdin.Close()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't close archive stream`,
		)
	}

	for _, receiver := range receivers.nodes {
		err := receiver.command.Wait()
		if err != nil {
			if sshErr, ok := err.(*ssh.ExitError); ok {
				return fmt.Errorf(
					`%s failed to receive archive, `+
						`remote command exited with non-zero code: %d`,
					receiver.node.String(),
					sshErr.Waitmsg.ExitStatus(),
				)
			}

			return hierr.Errorf(
				err,
				`%s failed to receive archive, unexpected error`,
				receiver.node.String(),
			)
		}
	}

	return nil
}
