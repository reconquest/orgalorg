package main

import (
	"regexp"
	"strconv"

	"github.com/seletskiy/hierr"
)

func parseAddress(
	address string, defaultUser string, defaultPort int,
) (user, host string, port int, err error) {
	addressRegexp := regexp.MustCompile(`^(?:([^@]+)@)?(.*?)(?::(\d+))?$`)

	matches := addressRegexp.FindStringSubmatch(address)

	user = matches[1]
	if user == "" {
		user = defaultUser
	}

	host = matches[2]

	rawPort := matches[3]
	if rawPort == "" {
		port = defaultPort
	} else {
		port, err = strconv.Atoi(rawPort)
		if err != nil {
			return "", "", 0, hierr.Errorf(
				err,
				`can't parse port number: '%s'`, rawPort,
			)
		}
	}

	return user, host, port, nil
}
