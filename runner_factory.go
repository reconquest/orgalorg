package main

import (
	"fmt"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/reconquest/runcmd"
)

type (
	runnerFactory func(address address) (runcmd.Runner, error)
)

func createRemoteRunnerFactoryWithAgent(
	keyring agent.Agent,
	timeout *runcmd.Timeout,
) runnerFactory {
	return func(address address) (runcmd.Runner, error) {
		return runcmd.NewRemoteRunner(
			address.user,
			fmt.Sprintf("%s:%d", address.domain, address.port),
			[]ssh.AuthMethod{
				ssh.PublicKeysCallback(keyring.Signers),
			},
			*timeout,
		)
	}
}

func createRemoteRunnerFactoryWithPassword(
	password string,
	timeout *runcmd.Timeout,
) runnerFactory {
	return func(address address) (runcmd.Runner, error) {
		return runcmd.NewRemotePasswordAuthRunner(
			address.user,
			fmt.Sprintf("%s:%d", address.domain, address.port),
			password,
			*timeout,
		)
	}
}
