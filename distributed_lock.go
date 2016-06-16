package main

import (
	"strings"
	"sync"
	"time"

	"github.com/seletskiy/hierr"
	"github.com/theairkit/runcmd"
)

type distributedLock struct {
	nodes []*distributedLockNode

	failOnError bool
}

func (lock *distributedLock) addNodeRunner(
	runner runcmd.Runner,
	address address,
) {
	lock.nodes = append(lock.nodes, &distributedLockNode{
		address: address,
		runner:  runner,
	})
}

func (lock *distributedLock) acquire(filename string) error {
	for nodeIndex, node := range lock.nodes {
		tracef(
			"%4d/%d locking node: '%s'",
			nodeIndex+1,
			len(lock.nodes),
			node.String(),
		)

		err := node.lock(filename)
		if err != nil {
			if !lock.failOnError {
				warningf(
					"%s",
					hierr.Errorf(
						err,
						`failed to acquire lock, `+
							`but proceeding with execution`,
					),
				)

				continue
			}

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

	return nil
}

func (lock *distributedLock) runHeartbeats(
	period time.Duration,
	canceler *sync.Cond,
) {
	for _, node := range lock.nodes {
		if node.connection != nil {
			go heartbeat(period, node, canceler)
		}
	}
}
