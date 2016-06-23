package main

import (
	"strconv"

	"github.com/seletskiy/hierr"
)

type threadPool struct {
	available chan struct{}

	size int
}

func newThreadPool(size int) *threadPool {
	available := make(chan struct{}, size)
	for i := 0; i < size; i++ {
		available <- struct{}{}
	}

	return &threadPool{
		available,
		size,
	}
}

func (pool *threadPool) run(task func()) {
	<-pool.available
	defer func() {
		pool.available <- struct{}{}
	}()

	task()
}

func parseThreadPoolSize(args map[string]interface{}) (int, error) {
	var (
		poolSizeRaw = args["--threads"].(string)
	)

	poolSize, err := strconv.Atoi(poolSizeRaw)
	if err != nil {
		return 0, hierr.Errorf(
			err,
			`can't parse threads count`,
		)
	}

	return poolSize, nil
}
