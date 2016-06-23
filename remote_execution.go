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
			pool.run(func() {
				results <- &remoteExecutionResult{node, node.wait()}
			})
		}(node)
	}

	executionErrors := hierr.Errorf(
		nil,
		`can't run remote commands on %d nodes`,
		len(execution.nodes),
	)

	erroneous := false

	for range execution.nodes {
		result := <-results
		if result.err != nil {
			executionErrors = hierr.Push(
				executionErrors,
				hierr.Errorf(
					result.err,
					`%s has finished with error`,
					result.node.node.String(),
				),
			)

			erroneous = true

			continue
		}

		tracef(
			`%s has successfully finished execution`,
			result.node.node.String(),
		)
	}

	if erroneous {
		return executionErrors
	}

	return nil
}
