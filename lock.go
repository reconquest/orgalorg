package main

import (
	"sync"
	"sync/atomic"

	"github.com/seletskiy/hierr"
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
	failOnError bool,
	noConnFail bool,
) (*distributedLock, error) {
	var (
		cluster = &distributedLock{
			failOnError: failOnError,
		}

		connectedCount = int64(0)
		failedCount    = int64(0)

		errors = make(chan error, 0)

		mutex = &sync.Mutex{}
	)

	for _, nodeAddress := range addresses {
		go func(nodeAddress address) {
			pool.run(func() {
				err := connectToNode(cluster, runnerFactory, nodeAddress, mutex)

				if err != nil {
					atomic.AddInt64(&failedCount, 1)

					if noConnFail {
						warningf("%s", err)
						errors <- nil
					} else {
						errors <- err
					}

					return
				}

				debugf(`%4d/%d (failed: %d) connection established: %s`,
					atomic.AddInt64(&connectedCount, 1),
					int64(len(addresses))-failedCount,
					failedCount,
					nodeAddress,
				)

				errors <- err
			})
		}(nodeAddress)
	}

	for range addresses {
		err := <-errors
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't acquire lock`,
			)
		}
	}

	err := cluster.acquire(lockFile)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't acquire global cluster lock on %d hosts`,
			len(addresses),
		)
	}

	return cluster, nil
}

func connectToNode(
	cluster *distributedLock,
	runnerFactory runnerFactory,
	address address,
	nodeAddMutex sync.Locker,
) error {
	tracef(`connecting to address: '%s'`, address)

	runner, err := runnerFactory(address)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't connect to address: %s`,
			address,
		)
	}

	nodeAddMutex.Lock()
	defer nodeAddMutex.Unlock()

	cluster.addNodeRunner(runner, address)

	return nil
}
