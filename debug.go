package main

import (
	"fmt"

	"github.com/seletskiy/hierr"
)

func tracef(format string, args ...interface{}) {
	if verbose < verbosityTrace {
		return
	}

	args = serializeErrors(args)

	logger.Debugf(format, args...)
}

func debugf(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Debugf(format, args...)
}

func infof(format string, args ...interface{}) {
	args = serializeErrors(args)

	logger.Infof(format, args...)
}

func warningf(format string, args ...interface{}) {
	args = serializeErrors(args)

	if verbose <= verbosityQuiet {
		return
	}

	logger.Warningf(format, args...)
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

func serializeError(err error) string {
	if format == outputFormatText {
		return fmt.Sprint(err)
	}

	if hierarchicalError, ok := err.(hierr.Error); ok {
		serializedError := fmt.Sprint(hierarchicalError.Nested)
		if nested, ok := hierarchicalError.Nested.(error); ok {
			serializedError = serializeError(nested)
		}

		return hierarchicalError.Message + ": " + serializedError
	}

	return err.Error()
}
