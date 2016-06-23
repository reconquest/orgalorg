package main

import (
	"sync"
	"time"
)

type distributedLock struct {
	nodes []*distributedLockNode
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
