package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/reconquest/hierr-go"
)

var (
	hostRegexp = regexp.MustCompile(`^(?:([^@]+)@)?(.*?)(?::(\d+))?$`)
)

type address struct {
	user   string
	domain string
	port   int
}

func (address address) String() string {
	return fmt.Sprintf(
		"[%s@%s:%d]",
		address.user,
		address.domain,
		address.port,
	)
}

func parseAddress(
	host string, defaultUser string, defaultPort int,
) (address, error) {
	matches := hostRegexp.FindStringSubmatch(host)

	var (
		user    = defaultUser
		domain  = matches[2]
		rawPort = matches[3]
		port    = defaultPort
	)

	if matches[1] != "" {
		user = matches[1]
	}

	if rawPort != "" {
		var err error
		port, err = strconv.Atoi(rawPort)
		if err != nil {
			return address{}, hierr.Errorf(
				err,
				`can't parse port number: '%s'`, rawPort,
			)
		}
	}

	return address{
		user:   user,
		domain: domain,
		port:   port,
	}, nil
}

func getUniqueAddresses(addresses []address) []address {
	result := []address{}

	for _, origin := range addresses {
		keep := true

		for _, another := range result {
			if origin.user != another.user {
				continue
			}

			if origin.domain != another.domain {
				continue
			}

			if origin.port != another.port {
				continue
			}

			keep = false
		}

		if keep {
			result = append(result, origin)
		}
	}

	return result
}
