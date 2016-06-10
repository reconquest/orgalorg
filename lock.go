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
	noLockFail bool,
) (*distributedLock, error) {
	var (
		lock = &distributedLock{
			noFail: noLockFail,
		}

		nodeIndex = int64(0)
		errors    = make(chan error, 0)

		mutex = &sync.Mutex{}
	)

	for _, nodeAddress := range addresses {
		go func(nodeAddress address) {
			tracef(`connecting to address: '%s'`,
				nodeAddress,
			)

			runner, err := runnerFactory(nodeAddress)
			if err != nil {
				errors <- hierr.Errorf(
					err,
					`can't create runner for address: %s`,
					nodeAddress,
				)

				return
			}

			debugf(`%4d/%d connection established: %s`,
				atomic.AddInt64(&nodeIndex, 1),
				len(addresses),
				nodeAddress,
			)

			mutex.Lock()
			{
				err = lock.addNodeRunner(runner, nodeAddress)
			}
			mutex.Unlock()
			if err != nil {
				errors <- hierr.Errorf(
					err,
					`can't add host to the global cluster lock: %s`,
					nodeAddress,
				)

				return
			}

			errors <- nil
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

	err := lock.acquire(lockFile)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't acquire global cluster lock on %d hosts`,
			len(addresses),
		)
	}

	return lock, nil
}
