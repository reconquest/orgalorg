package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kovetskiy/lorg"
	"github.com/reconquest/hierr-go"
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
	logger.SetFormat(style)
	logger.SetIndentLines(true)

	logger.SetShiftIndent(28)
}

func tracef(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Tracef(format, args...)

	drawStatus()
}

func traceln(args ...interface{}) {
	tracef("%s", fmt.Sprint(serializeErrors(args)...))
}

func debugf(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Debugf(format, args...)

	drawStatus()
}

func debugln(args ...interface{}) {
	debugf("%s", fmt.Sprint(serializeErrors(args)...))
}

func infof(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Infof(format, args...)

	drawStatus()
}

func infoln(args ...interface{}) {
	infof("%s", fmt.Sprint(serializeErrors(args)...))
}

func warningf(format string, args ...interface{}) {
	args = serializeErrors(args)

	if verbose <= verbosityQuiet {
		return
	}

	logger.Warningf(format, args...)

	drawStatus()
}

func warningln(args ...interface{}) {
	warningf("%s", fmt.Sprint(serializeErrors(args)...))
}

func errorf(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Errorf(format, args...)
}

func errorln(args ...interface{}) {
	errorf("%s", fmt.Sprint(serializeErrors(args)...))
}

func fatalf(format string, args ...interface{}) {
	args = serializeErrors(args)

	clearStatus()

	logger.Fatalf(format, args...)

	exit(1)
}

func fatalln(args ...interface{}) {
	fatalf("%s", fmt.Sprint(serializeErrors(args)...))
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
	if statusbar == nil {
		return
	}

	clearStatus()

	statusbar.SetStatus(status)

	drawStatus()
}

func shouldDrawStatus() bool {
	if statusbar == nil {
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

	err := statusbar.Render(os.Stderr)
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

	statusbar.Clear(os.Stderr)
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
