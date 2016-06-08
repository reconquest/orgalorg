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

		nodeIndex = int64(0)
		errors    = make(chan error, 0)

		mutex = &sync.Mutex{}
	)

	for _, nodeAddress := range addresses {
		go func(nodeAddress address) {
			err := connectToNode(cluster, runnerFactory, nodeAddress, mutex)

			if err != nil {
				if noConnFail {
					warningf("%s", err.Error())
					errors <- nil
				} else {
					errors <- err
				}

				return
			}

			debugf(`%4d/%d connection established: %s`,
				atomic.AddInt64(&nodeIndex, 1),
				len(addresses),
				nodeAddress,
			)

			errors <- err
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
			`can't create runner for address: %s`,
			address,
		)
	}

	nodeAddMutex.Lock()
	defer nodeAddMutex.Unlock()

	cluster.addNodeRunner(runner, address)

	return nil
}
