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
		return runcmd.NewRemoteKeyAuthRunnerWithTimeouts(
			address.user,
			fmt.Sprintf("%s:%d", address.domain, address.port),
			key,
			*timeouts,
		)
	}
}

func createRemoteRunnerFactoryWithPassword(
	password string,
	timeouts *runcmd.Timeouts,
) runnerFactory {
	return func(address address) (runcmd.Runner, error) {
		return runcmd.NewRemotePassAuthRunnerWithTimeouts(
			address.user,
			fmt.Sprintf("%s:%d", address.domain, address.port),
			password,
			*timeouts,
		)
	}
}
