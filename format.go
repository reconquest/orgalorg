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
) outputFormat {

	formatType := outputFormatText
	if args["--json"].(bool) {
		formatType = outputFormatJSON
	}

	return formatType
}

func isOutputOnTTY() bool {
	return terminal.IsTerminal(int(os.Stderr.Fd()))
}

func isColorEnabled(args map[string]interface{}, hasTTY bool) bool {
	isColorEnabled := hasTTY

	if format != outputFormatText {
		isColorEnabled = false
	}

	if args["--no-colors"].(bool) {
		isColorEnabled = false
	}

	return isColorEnabled
}
