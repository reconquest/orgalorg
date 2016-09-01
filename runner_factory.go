package main

import (
	"fmt"

	"github.com/theairkit/runcmd"
)

type (
	runnerFactory func(address address) (runcmd.Runner, error)
)

func createRemoteRunnerFactoryWithKey(
	key string,
	timeouts *runcmd.Timeouts,
) runnerFactory {
	return func(address address) (runcmd.Runner, error) {
		return createRunner(
			runcmd.NewRemoteRawKeyAuthRunnerWithTimeouts,
			key,
			address,
			*timeouts,
		)
	}
}

func createRemoteRunnerFactoryWithPassword(
	password string,
	timeouts *runcmd.Timeouts,
) runnerFactory {
	return func(address address) (runcmd.Runner, error) {
		return createRunner(
			runcmd.NewRemotePassAuthRunnerWithTimeouts,
			password,
			address,
			*timeouts,
		)
	}
}

func createRemoteRunnerFactoryWithAgent(
	sock string,
	timeouts *runcmd.Timeouts,
) runnerFactory {
	return func(address address) (runcmd.Runner, error) {
		return createRunner(
			runcmd.NewRemoteAgentAuthRunnerWithTimeouts,
			sock,
			address,
			*timeouts,
		)
	}
}

func createRunner(
	factory func(string, string, string, runcmd.Timeouts) (
		*runcmd.Remote,
		error,
	),

	key string,
	address address,
	timeouts runcmd.Timeouts,
) (runcmd.Runner, error) {
	return factory(
		address.user,
		fmt.Sprintf("%s:%d", address.domain, address.port),
		key,
		timeouts,
	)
}
