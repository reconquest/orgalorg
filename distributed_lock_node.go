package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/reconquest/go-lineflushwriter"
	"github.com/reconquest/go-prefixwriter"
	"github.com/seletskiy/hierr"
	"github.com/theairkit/runcmd"
)

const (
	lockAcquiredString = `acquired`
	lockLockedString   = `locked`
)

type distributedLockNode struct {
	address address
	runner  runcmd.Runner

	connection *distributedLockConnection
}

func (node *distributedLockNode) String() string {
	return node.address.String()
}

type distributedLockConnection struct {
	stdin  io.WriteCloser
	stdout io.Reader
}

func (node *distributedLockNode) lock(
	filename string,
) error {
	lockCommandString := fmt.Sprintf(
		`sh -c $'`+
			`flock -nx %s -c \'`+
			`printf "%s\\n" && cat\' || `+
			`printf "%s\\n"'`,
		filename,
		lockAcquiredString,
		lockLockedString,
	)

	logMutex := &sync.Mutex{}

	tracef("%s", hierr.Errorf(
		lockCommandString,
		`%s running lock command`,
		node,
	))

	lockCommand, err := node.runner.Command(lockCommandString)
	if err != nil {
		return err
	}

	stdout, err := lockCommand.StdoutPipe()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't get control stdout pipe from lock process`,
		)
	}

	stderr := lineflushwriter.New(
		prefixwriter.New(
			newDebugWriter(logger),
			fmt.Sprintf("%s {flock} <stderr> ", node.String()),
		),
		logMutex,
		true,
	)

	lockCommand.SetStderr(stderr)

	stdin, err := lockCommand.StdinPipe()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't get control stdin pipe to lock process`,
		)
	}

	err = lockCommand.Start()
	if err != nil {
		return hierr.Errorf(
			err,
			`%s can't start lock command: '%s`,
			node, lockCommandString,
		)
	}

	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		return hierr.Errorf(
			err,
			`can't read line from lock process`,
		)
	}

	switch strings.TrimSpace(line) {
	case lockAcquiredString:
		// pass

	case lockLockedString:
		return fmt.Errorf(
			`%s can't acquire lock, `+
				`lock already obtained by another process `+
				`or unavailable`,
			node,
		)

	default:
		return fmt.Errorf(
			`%s unexpected reply string encountered `+
				`instead of '%s' or '%s': '%s'`,
			node, lockAcquiredString, lockLockedString,
			line,
		)
	}

	tracef(`lock acquired: '%s' on '%s'`, node, filename)

	node.connection = &distributedLockConnection{
		stdin:  stdin,
		stdout: stdout,
	}

	return nil
}
