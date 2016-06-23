package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/seletskiy/hierr"
)

const (
	longConnectionWarningTimeout = 2 * time.Second
)

// acquireDistributedLock tries to acquire atomic file lock on each of
// specified remote nodes. lockFile is used to specify target lock file, it
// must exist on every node. runnerFactory will be used to make connection
// to remote node. If noLockFail is given, then only warning will be printed
// if lock process has been failed.
func acquireDistributedLock(
	lockFile string,
	runnerFactory runnerFactory,
	addresses []address,
	noLockFail bool,
	noConnFail bool,
) (*distributedLock, error) {
	var (
		cluster = &distributedLock{}

		connections = int64(0)
		failures    = int64(0)

		errors = make(chan error, 0)

		nodeAddMutex = &sync.Mutex{}
	)

	for _, nodeAddress := range addresses {
		go func(nodeAddress address) {
			pool.run(func() {
				failed := false

				node, err := connectToNode(cluster, runnerFactory, nodeAddress)
				if err != nil {
					atomic.AddInt64(&failures, 1)

					if noConnFail {
						failed = true
						warningf("%s", err)
					} else {
						errors <- err
						return
					}
				} else {
					err = node.lock(lockFile)
					if err != nil {
						if noLockFail {
							failed = true
							warningf("%s", err)
						} else {
							errors <- err
							return
						}
					}
				}

				status := "established"
				if failed {
					status = "failed"
				} else {
					atomic.AddInt64(&connections, 1)

					nodeAddMutex.Lock()
					defer nodeAddMutex.Unlock()

					cluster.nodes = append(cluster.nodes, node)
				}

				debugf(
					`%4d/%d (%d failed) connection %s: %s`,
					connections,
					int64(len(addresses))-failures,
					failures,
					status,
					nodeAddress,
				)

				errors <- nil
			})
		}(nodeAddress)
	}

	erronous := 0
	topError := hierr.Push(`can't connect to nodes`)
	for range addresses {
		err := <-errors
		if err != nil {
			erronous += 1

			topError = hierr.Push(topError, err)
		}
	}

	if erronous > 0 {
		return nil, hierr.Push(
			fmt.Errorf(
				`connection to %d of %d nodes failed`,
				erronous,
				len(addresses),
			),
			topError,
		)
	}

	return cluster, nil
}

func connectToNode(
	cluster *distributedLock,
	runnerFactory runnerFactory,
	address address,
) (*distributedLockNode, error) {
	tracef(`connecting to address: '%s'`, address)

	done := make(chan struct{}, 0)

	go func() {
		select {
		case <-time.After(longConnectionWarningTimeout):
			warningf(
				"still connecting to address after %s: %s",
				longConnectionWarningTimeout,
				address,
			)

		case <-done:
		}

		return
	}()

	runner, err := runnerFactory(address)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't connect to address: %s`,
			address,
		)
	}

	done <- struct{}{}

	return &distributedLockNode{
		address: address,
		runner:  runner,
	}, nil
}
