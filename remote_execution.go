package main

import (
	"io"

	"github.com/seletskiy/hierr"
)

type remoteExecution struct {
	stdin io.WriteCloser
	nodes []remoteExecutionNode
}

func (execution *remoteExecution) wait() error {
	err := execution.stdin.Close()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't close stdin stream`,
		)
	}

	for _, node := range execution.nodes {
		err := node.wait()
		if err != nil {
			return hierr.Errorf(
				err,
				`wait finished with error`,
			)
		}
	}

	return nil
}
