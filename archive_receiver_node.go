package main

import "github.com/theairkit/runcmd"

type archiveReceiverNode struct {
	node    distributedLockNode
	command runcmd.CmdWorker
}
