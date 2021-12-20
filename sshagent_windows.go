//go:build windows

package main

import (
	"github.com/davidmz/go-pageant"
	"github.com/reconquest/hierr-go"

	"github.com/Microsoft/go-winio"
	"golang.org/x/crypto/ssh/agent"
)

const (
	openSshAgentPipe = `\\.\pipe\openssh-ssh-agent`
)

func getSshAgent() (agent.Agent, error) {
	debugf(`trying windows pageant`)
	if pageant.Available() {
		return pageant.New(), nil
	} else {
		debugf("    pageant is not found")
	}
	debugf(`trying windows openssh-agent`)
	sock, err := winio.DialPipe(openSshAgentPipe, nil)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"unable to dial openssh-agent socket: %s",
			openSshAgentPipe,
		)
	}
	return agent.NewClient(sock), nil
}
