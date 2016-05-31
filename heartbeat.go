package main

import (
	"bufio"
	"io"
	"strings"
	"time"

	"github.com/seletskiy/hierr"
)

const (
	heartbeatPing = "PING"
)

func heartbeat(period time.Duration, node *distributedLockNode) {
	ticker := time.Tick(period)

	for {
		_, err := io.WriteString(node.connection.stdin, heartbeatPing+"\n")
		if err != nil {
			logger.Fatal(hierr.Errorf(err, `can't send heartbeat`))
		}

		<-ticker

		ping, err := bufio.NewReader(node.connection.stdout).ReadString('\n')
		if err != nil {
			logger.Fatal(hierr.Errorf(err, `can't receive heartbeat`))
		}

		if strings.TrimSpace(ping) != heartbeatPing {
			logger.Fatalf(
				`received unexpected heartbeat ping: '%s'`,
				ping,
			)
		}

		tracef(`%s heartbeat`, node.String())
	}
}
