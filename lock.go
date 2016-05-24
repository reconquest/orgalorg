package main

import "github.com/seletskiy/hierr"

func acquireDistributedLock(
	args map[string]interface{},
	runnerFactory runnerFactory,
	addresses []address,
) (*distributedLock, error) {
	var (
		lockFile = args["--lock-file"].(string)
	)

	lock := &distributedLock{}

	for _, address := range addresses {
		runner, err := runnerFactory(address)
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't create runner for address '%s'`,
				address,
			)
		}

		err = lock.addNodeRunner(runner, address)
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't add host to the global cluster lock: '%s'`,
				address,
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
