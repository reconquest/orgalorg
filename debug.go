package main

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

	logger.Warningf(format, args...)
}

func errorf(format string, args ...interface{}) {
	loggerLock.Lock()
	defer loggerLock.Unlock()

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
