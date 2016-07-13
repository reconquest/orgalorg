package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/kovetskiy/lorg"
	"github.com/reconquest/loreley"
	"github.com/seletskiy/hierr"
)

var (
	loggerFormattingBasicLength = 0
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

func setLoggerStyle(logger *lorg.Log, style lorg.Formatter) {
	testLogger := lorg.NewLog()

	testLogger.SetFormat(style)

	buffer := &bytes.Buffer{}
	testLogger.SetOutput(buffer)

	testLogger.Debug(``)

	loggerFormattingBasicLength = len(strings.TrimSuffix(
		loreley.TrimStyles(buffer.String()),
		"\n",
	))

	logger.SetFormat(style)
}

func tracef(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Tracef(`%s`, wrapNewLines(format, args...))

	drawStatus()
}

func debugf(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Debugf(`%s`, wrapNewLines(format, args...))

	drawStatus()
}

func infof(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Infof(`%s`, wrapNewLines(format, args...))

	drawStatus()
}

func warningf(format string, args ...interface{}) {
	args = serializeErrors(args)

	if verbose <= verbosityQuiet {
		return
	}

	logger.Warningf(`%s`, wrapNewLines(format, args...))

	drawStatus()
}

func errorf(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Errorf(`%s`, wrapNewLines(format, args...))
}

func fatalf(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Fatalf(`%s`, wrapNewLines(format, args...))

	exit(1)
}

func wrapNewLines(format string, values ...interface{}) string {
	contents := fmt.Sprintf(format, values...)
	contents = strings.TrimSuffix(contents, "\n")
	contents = strings.Replace(
		contents,
		"\n",
		"\n"+strings.Repeat(" ", loggerFormattingBasicLength),
		-1,
	)

	return contents
}

func serializeErrors(args []interface{}) []interface{} {
	for i, arg := range args {
		if err, ok := arg.(error); ok {
			args[i] = serializeError(err)
		}
	}

	return args
}

func setStatus(status interface{}) {
	if bar == nil {
		return
	}

	bar.SetStatus(status)
}

func shouldDrawStatus() bool {
	if bar == nil {
		return false
	}

	if format != outputFormatText {
		return false
	}

	if verbose <= verbosityQuiet {
		return false
	}

	return true
}

func drawStatus() {
	if !shouldDrawStatus() {
		return
	}

	err := bar.Render(os.Stderr)
	if err != nil {
		errorf(
			"%s", hierr.Errorf(
				err,
				`can't draw status bar`,
			),
		)
	}
}

func clearStatus() {
	if !shouldDrawStatus() {
		return
	}

	bar.Clear(os.Stderr)
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
