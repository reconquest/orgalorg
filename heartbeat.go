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

func heartbeat(
	period time.Duration,
	node *distributedLockNode,
	canceler *sync.Cond,
) {
	abort := make(chan struct{}, 0)

	go func() {
		canceler.L.Lock()
		canceler.Wait()
		canceler.L.Unlock()

		abort <- struct{}{}
	}()

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

	for {
		_, err := io.WriteString(node.connection.stdin, heartbeatPing+"\n")
		if err != nil {
			logger.Error(
				hierr.Errorf(
					err,
					`%s can't send heartbeat`,
					node.String(),
				).Error(),
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
			logger.Error(
				hierr.Errorf(
					err,
					`%s can't receive heartbeat`,
					node.String(),
				),
			)

			finish(2)
		}

		if strings.TrimSpace(ping) != heartbeatPing {
			logger.Errorf(
				`%s received unexpected heartbeat ping: '%s'`,
				node.String(),
				ping,
			)

			finish(2)
		}

		tracef(`%s heartbeat`, node.String())
	}
}
