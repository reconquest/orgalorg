package main

import (
	"io"

	"github.com/seletskiy/hierr"
)

type remoteExecution struct {
	stdin io.WriteCloser
	nodes map[*distributedLockNode]*remoteExecutionNode
}

type remoteExecutionResult struct {
	node *remoteExecutionNode

	err error
}

func (execution *remoteExecution) wait() error {
	tracef("waiting %d nodes to finish", len(execution.nodes))

	results := make(chan *remoteExecutionResult, 0)
	for _, node := range execution.nodes {
		go func(node *remoteExecutionNode) {
			results <- &remoteExecutionResult{node, node.wait()}
		}(node)
	}

	for range execution.nodes {
		result := <-results
		if result.err != nil {
			return hierr.Errorf(
				result.err,
				`%s has finished with error`,
				result.node.node.String(),
			)
		}

		tracef(
			`%s has successfully finished execution`,
			result.node.node.String(),
		)
	}

	return nil
}
