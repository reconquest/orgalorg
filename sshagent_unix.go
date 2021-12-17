//go:build !windows

package main

import (
	"net"
	"os"

	"github.com/reconquest/hierr-go"
	"golang.org/x/crypto/ssh/agent"
)

func getSshAgent() (agent.Agent, error) {
	debugf(`trying unix ssh-agent pipe`)
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		return nil, hierr.Errorf(
			nil,
			"SSH_AUTH_SOCK is not set",
		)
	}
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"unable to dial to ssh agent socket: %s",
			os.Getenv("SSH_AUTH_SOCK"),
		)
	}

	return agent.NewClient(sock), nil
}
