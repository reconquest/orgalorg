package main

import (
	"bufio"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/seletskiy/hierr"
	"github.com/theairkit/runcmd"
)

const (
	heartbeatPing = "PING"
)

// heartbeat runs infinite process of sending test messages to the connected
// node. All heartbeats to all nodes are connected to each other, so if one
// heartbeat routine exits, all heartbeat routines will exit, because in that
// case orgalorg can't guarantee global lock.
func heartbeat(
	period time.Duration,
	node *distributedLockNode,
	canceler *sync.Cond,
) {
	abort := make(chan struct{}, 0)

	// Internal go-routine for listening abort broadcast and finishing current
	// heartbeat process.
	go func() {
		canceler.L.Lock()
		canceler.Wait()
		canceler.L.Unlock()

		abort <- struct{}{}
	}()

	// Finish finishes current go-routine and send abort broadcast to all
	// connected go-routines.
	finish := func(code int) {
		canceler.L.Lock()
		canceler.Broadcast()
		canceler.L.Unlock()

		<-abort

		if remote, ok := node.runner.(*runcmd.Remote); ok {
			tracef("%s closing connection", node.String())
			err := remote.CloseConnection()
			if err != nil {
				warningf(
					"%s",
					hierr.Errorf(
						err,
						"%s error while closing connection",
						node.String(),
					),
				)
			}
		}

		exit(code)
	}

	ticker := time.Tick(period)

	// Infinite loop of heartbeating. It will send heartbeat message, wait
	// fraction of send timeout time and try to receive heartbeat response.
	// If no response received, heartbeat process aborts.
	for {
		_, err := io.WriteString(node.connection.stdin, heartbeatPing+"\n")
		if err != nil {
			errorf(
				"%s",
				hierr.Errorf(
					err,
					`%s can't send heartbeat`,
					node.String(),
				),
			)

			finish(2)
		}

		select {
		case <-abort:
			return

		case <-ticker:
			// pass
		}

		ping, err := bufio.NewReader(node.connection.stdout).ReadString('\n')
		if err != nil {
			errorf(
				"%s",
				hierr.Errorf(
					err,
					`%s can't receive heartbeat`,
					node.String(),
				),
			)

			finish(2)
		}

		if strings.TrimSpace(ping) != heartbeatPing {
			errorf(
				`%s received unexpected heartbeat ping: '%s'`,
				node.String(),
				ping,
			)

			finish(2)
		}

		tracef(`%s heartbeat`, node.String())
	}
}
