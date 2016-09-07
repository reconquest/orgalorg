package main

type distributedLock struct {
	nodes []*distributedLockNode
}
