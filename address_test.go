package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAddress_ParseValidDomainAddressWithDefaults(t *testing.T) {
	assertions := assert.New(t)

	tests := []struct {
		Host        string
		DefaultUser string
		DefaultPort int

		ExpectedUser   string
		ExpectedDomain string
		ExpectedPort   int
	}{
		{"example.com", "_", 23,
			"_", "example.com", 23},

		{"user@example.com", "_", 23,
			"user", "example.com", 23},

		{"example.com:1234", "_", 23,
			"_", "example.com", 1234},

		{"user@example.com:1234", "_", 23,
			"user", "example.com", 1234},

		{"example.com:1234", "user", 23,
			"user", "example.com", 1234},

		{"user2@example.com:1234", "user", 23,
			"user2", "example.com", 1234},
	}

	for _, test := range tests {
		parsed, err := parseAddress(
			test.Host, test.DefaultUser, test.DefaultPort,
		)

		assertions.Nil(err)
		assertions.Equal(test.ExpectedUser, parsed.user)
		assertions.Equal(test.ExpectedPort, parsed.port)
		assertions.Equal(test.ExpectedDomain, parsed.domain)
	}
}
