package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"

	"github.com/kovetskiy/lorg"
	"github.com/seletskiy/hierr"
)

func setLoggerOutputFormat(logger *lorg.Log, format outputFormat) {
	if format == outputFormatJSON {
		logger.SetOutput(&jsonOutputWriter{
			stream: `stderr`,
			node:   ``,
			output: os.Stderr,
		})
	}
}

func setLoggerVerbosity(level verbosity, logger *lorg.Log) {
	logger.SetLevel(lorg.LevelWarning)

	switch {
	case level >= verbosityTrace:
		logger.SetLevel(lorg.LevelTrace)

	case level >= verbosityDebug:
		logger.SetLevel(lorg.LevelDebug)

	case level >= verbosityNormal:
		logger.SetLevel(lorg.LevelInfo)
	}
}

func setLoggerStyle(logger *lorg.Log, style *lorg.Format) {
	logger.SetFormat(style)
}

func colorize(
	attributes ...color.Attribute,
) string {
	if !isColorEnabled {
		return ""
	}

	sequence := []string{}
	for _, attribute := range attributes {
		sequence = append(sequence, fmt.Sprint(attribute))
	}

	return fmt.Sprintf("\x1b[%sm", strings.Join(sequence, ";"))
}

func tracef(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Tracef(format, args...)

	drawStatus()
}

func debugf(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Debugf(format, args...)

	drawStatus()
}

func infof(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Infof(format, args...)

	drawStatus()
}

func warningf(format string, args ...interface{}) {
	args = serializeErrors(args)

	if verbose <= verbosityQuiet {
		return
	}

	logger.Warningf(format, args...)

	drawStatus()
}

func errorf(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Errorf(format, args...)
}

func serializeErrors(args []interface{}) []interface{} {
	for i, arg := range args {
		if err, ok := arg.(error); ok {
			args[i] = serializeError(err)
		}
	}

	return args
}

func shouldDrawStatus() bool {
	if !isOutputOnTTY {
		return false
	}

	if format != outputFormatText {
		return false
	}

	if verbose <= verbosityQuiet {
		return false
	}

	if status == nil {
		return false
	}

	return true
}

func drawStatus() {
	if !shouldDrawStatus() {
		return
	}

	status.Draw(os.Stderr)
}

func clearStatus() {
	if !shouldDrawStatus() {
		return
	}

	status.Clear(os.Stderr)
}

func serializeError(err error) string {
	if format == outputFormatText {
		return fmt.Sprint(err)
	}

	if hierarchicalError, ok := err.(hierr.Error); ok {
		serializedError := fmt.Sprint(hierarchicalError.Nested)
		switch nested := hierarchicalError.Nested.(type) {
		case error:
			serializedError = serializeError(nested)

		case []hierr.NestedError:
			serializeErrorParts := []string{}

			for _, nestedPart := range nested {
				serializedPart := fmt.Sprint(nestedPart)
				switch part := nestedPart.(type) {
				case error:
					serializedPart = serializeError(part)

				case string:
					serializedPart = part
				}

				serializeErrorParts = append(
					serializeErrorParts,
					serializedPart,
				)
			}

			serializedError = strings.Join(serializeErrorParts, "; ")
		}

		return hierarchicalError.Message + ": " + serializedError
	}

	return err.Error()
}
