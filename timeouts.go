package main

import (
	"strconv"
	"time"

	"github.com/seletskiy/hierr"
	"github.com/theairkit/runcmd"
)

func makeTimeouts(args map[string]interface{}) (*runcmd.Timeouts, error) {
	var (
		connectionTimeoutRaw = args["--conn-timeout"].(string)
		sendTimeoutRaw       = args["--send-timeout"].(string)
		receiveTimeoutRaw    = args["--recv-timeout"].(string)
		keepAliveRaw         = args["--keep-alive"].(string)
	)

	connectionTimeout, err := strconv.Atoi(connectionTimeoutRaw)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't convert specified connection timeout to number: '%s'`,
			connectionTimeoutRaw,
		)
	}

	sendTimeout, err := strconv.Atoi(sendTimeoutRaw)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't convert specified send timeout to number: '%s'`,
			sendTimeoutRaw,
		)
	}

	receiveTimeout, err := strconv.Atoi(receiveTimeoutRaw)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't convert specified receive timeout to number: '%s'`,
			receiveTimeoutRaw,
		)
	}

	keepAlive, err := strconv.Atoi(keepAliveRaw)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't convert specified keep alive time to number: '%s'`,
			keepAliveRaw,
		)
	}

	return &runcmd.Timeouts{
		ConnectionTimeout: time.Millisecond * time.Duration(connectionTimeout),
		SendTimeout:       time.Millisecond * time.Duration(sendTimeout),
		ReceiveTimeout:    time.Millisecond * time.Duration(receiveTimeout),
		KeepAlive:         time.Millisecond * time.Duration(keepAlive),
	}, nil
}
