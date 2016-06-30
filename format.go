package main

import (
	"github.com/reconquest/loreley"
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

func parseColorMode(args map[string]interface{}) loreley.ColorizeMode {
	switch args["--color"].(string) {
	case "always":
		return loreley.ColorizeAlways

	case "auto":
		return loreley.ColorizeOnTTY

	case "never":
		return loreley.ColorizeNever
	}

	return loreley.ColorizeNever
}
