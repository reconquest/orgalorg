package main

import (
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

type (
	outputFormat int
)

const (
	outputFormatText outputFormat = iota
	outputFormatJSON
)

func parseOutputFormat(
	args map[string]interface{},
) (outputFormat, bool, bool) {

	format := outputFormatText
	if args["--json"].(bool) {
		format = outputFormatJSON
	}

	isOutputOnTTY := terminal.IsTerminal(int(os.Stderr.Fd()))

	isColorEnabled := isOutputOnTTY

	if format != outputFormatText {
		isColorEnabled = false
	}

	if args["--no-colors"].(bool) {
		isColorEnabled = false
	}

	return format, isOutputOnTTY, isColorEnabled
}
