package main

import (
	"strings"

	"github.com/seletskiy/hierr"
	"github.com/theairkit/runcmd"
)

type distributedLock struct {
	nodes []distributedLockNode
}

func (lock *distributedLock) addNodeRunner(
	runner runcmd.Runner,
	address address,
) error {
	lock.nodes = append(lock.nodes, distributedLockNode{
		address: address,
		runner:  runner,
	})

	return nil
}

func (lock *distributedLock) acquire(filename string) error {
	for _, node := range lock.nodes {
		_, err := node.lock(filename)
		if err != nil {
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
