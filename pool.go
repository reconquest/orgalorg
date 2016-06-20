package main

import (
	"strconv"

	"github.com/seletskiy/hierr"
)

type threadPool struct {
	available chan struct{}
}

func newThreadPool(size int) *threadPool {
	available := make(chan struct{}, size)
	for i := 0; i < size; i++ {
		available <- struct{}{}
	}

	return &threadPool{
		available,
	}
}

func (pool *threadPool) run(task func()) {
	<-pool.available
	defer func() {
		pool.available <- struct{}{}
	}()

	task()
}

func createThreadPool(args map[string]interface{}) (*threadPool, error) {
	var (
		poolSizeRaw = args["--threads"].(string)
	)

	poolSize, err := strconv.Atoi(poolSizeRaw)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't parse threads count`,
		)
	}

	debugf(`using %d threads`, poolSize)

	return newThreadPool(poolSize), nil
}
