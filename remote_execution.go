package main

import (
	"fmt"
	"io"
	"reflect"

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
	tracef(`waiting %d nodes to finish`, len(execution.nodes))

	results := make(chan *remoteExecutionResult, 0)
	for _, node := range execution.nodes {
		go func(node *remoteExecutionNode) {
			results <- &remoteExecutionResult{node, node.wait()}
		}(node)
	}

	executionErrors := fmt.Errorf(
		`commands are exited with non-zero code`,
	)

	var (
		failures  = 0
		exitCodes = map[int]int{}
	)

	for range execution.nodes {
		result := <-results
		if result.err != nil {
			exitCodes[result.node.exitCode]++

			executionErrors = hierr.Push(
				executionErrors,
				hierr.Errorf(
					result.err,
					`%s has finished with error`,
					result.node.node.String(),
				),
			)

			failures++

			continue
		}

		tracef(
			`%s has successfully finished execution`,
			result.node.node.String(),
		)
	}

	if failures > 0 {
		if failures == len(execution.nodes) {
			exitCodesValue := reflect.ValueOf(exitCodes)

			topError := fmt.Errorf(
				`commands are exited with non-zero exit code on all %d nodes`,
				len(execution.nodes),
			)

			for _, key := range exitCodesValue.MapKeys() {
				topError = hierr.Push(
					topError,
					fmt.Sprintf(
						`code %d (%d nodes)`,
						key.Int(),
						exitCodesValue.MapIndex(key).Int(),
					),
				)
			}

			return topError
		}

		return hierr.Errorf(
			executionErrors,
			`commands are exited with non-zero exit code on %d of %d nodes`,
			failures,
			len(execution.nodes),
		)
	}

	return nil
}
