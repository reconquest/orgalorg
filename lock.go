package main

import "github.com/seletskiy/hierr"
import "sync/atomic"

func acquireDistributedLock(
	lockFile string,
	runnerFactory runnerFactory,
	addresses []address,
) (*distributedLock, error) {
	lock := &distributedLock{}

	nodeIndex := int64(0)
	errors := make(chan error, 0)
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

			err = lock.addNodeRunner(runner, nodeAddress)
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
