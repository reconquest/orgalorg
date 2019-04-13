package main

import (
	"fmt"
	"io"
	"reflect"
	"syscall"

	"github.com/reconquest/hierr-go"
	"github.com/reconquest/karma-go"
)

type remoteExecution struct {
	stdin io.WriteCloser
	nodes map[*distributedLockNode]*remoteExecutionNode
}

type remoteExecutionResult struct {
	node *remoteExecutionNode

	err error
}

func (execution *remoteExecution) signal(signal syscall.Signal) error {
	errors := karma.Format(nil, "unable to send signal to nodes")
	fails := 0

	for _, node := range execution.nodes {
		tracef("sending signal %v to %v", signal, node)

		err := node.command.Signal(signal)
		if err != nil {
			errors = karma.Push(errors, fmt.Errorf("%s: %s", node.String(), err))
			fails++
		}
	}

	if fails > 0 {
		return errors
	}

	return nil
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
		status = &struct {
			Phase   string
			Total   int
			Fails   int
			Success int
		}{
			Phase: `wait`,
			Total: len(execution.nodes),
		}

		exitCodes = map[int]int{}
	)

	setStatus(status)

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

			status.Fails++
			status.Total--

			tracef(
				`%s finished with exit code: '%d'`,
				result.node.node.String(),
				result.node.exitCode,
			)

			continue
		}

		status.Success++

		tracef(
			`%s has successfully finished execution`,
			result.node.node.String(),
		)
	}

	if status.Fails > 0 {
		if status.Fails == len(execution.nodes) {
			exitCodesValue := reflect.ValueOf(exitCodes)

			topError := fmt.Errorf(
				`commands are failed on all %d nodes`,
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
			`commands are failed on %d out of %d nodes`,
			status.Fails,
			len(execution.nodes),
		)
	}

	return nil
}
