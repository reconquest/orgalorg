package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/seletskiy/hierr"
	"github.com/theairkit/runcmd"
)

const (
	lockAcquiredString = `acquired`
	lockLockedString   = `locked`
)

type distributedLock struct {
	nodes []distributedLockNode
}

func (lock *distributedLock) addNodeRunner(
	runner runcmd.Runner,
	user string, domain string, port int,
) error {
	lock.nodes = append(lock.nodes, distributedLockNode{
		user:   user,
		domain: domain,
		port:   port,
		runner: runner,
	})

	return nil
}

func (lock *distributedLock) acquire(filename string) error {
	for _, node := range lock.nodes {
		_, err := node.lock(filename)
		if err != nil {
			nodes := []string{}
			for _, node := range lock.nodes {
				nodes = append(nodes, node.String())
			}

			return hierr.Errorf(
				err,
				"failed to lock %d nodes: %s",
				len(lock.nodes),
				strings.Join(nodes, ", "),
			)
		}
	}

	log.Infof(`global lock acquired on %d nodes`, len(lock.nodes))

	return nil
}

type distributedLockNode struct {
	user   string
	domain string
	port   int
	runner runcmd.Runner
}

func (node *distributedLockNode) String() string {
	return fmt.Sprintf("[%s@%s:%d]", node.user, node.domain, node.port)
}

type distributedLockReleaser struct {
	lockReadStdin io.WriteCloser
}

func (node *distributedLockNode) lock(
	filename string,
) (*distributedLockReleaser, error) {
	lockCommandString := fmt.Sprintf(
		`sh -c "`+
			`flock -nx %s -c '`+
			`printf \"%s\\n\" && read unlock' || `+
			`printf \"%s\\n\""`,
		filename,
		lockAcquiredString,
		lockLockedString,
	)

	log.Debug(hierr.Errorf(
		lockCommandString,
		`%s running lock command`,
		node,
	))

	lockCommand, err := node.runner.Command(lockCommandString)
	if err != nil {
		return nil, err
	}

	stdout, err := lockCommand.StdoutPipe()
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't get control stdout pipe from lock process`,
		)
	}

	stdin, err := lockCommand.StdinPipe()
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't get control stdin pipe to lock process`,
		)
	}

	err = lockCommand.Start()
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`%s can't start lock command: '%s`,
			node, lockCommandString,
		)
	}

	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't read line from lock process`,
		)
	}

	switch strings.TrimSpace(line) {
	case lockAcquiredString:
		// pass

	case lockLockedString:
		return nil, fmt.Errorf(
			`%s can't acquire lock, `+
				`lock already obtained by another process`,
			node,
		)

	default:
		return nil, fmt.Errorf(
			`%s unexpected reply string encountered `+
				`instead of '%s' or '%s': '%s'`,
			node, lockAcquiredString, lockLockedString,
			line,
		)
	}

	log.Debugf(`%s lock acquired`, node)

	return &distributedLockReleaser{
		lockReadStdin: stdin,
	}, nil
}
