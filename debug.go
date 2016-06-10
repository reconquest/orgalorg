package main

func tracef(format string, args ...interface{}) {
	// TODO always write debug to the file
	if verbose >= verbosityTrace {
		logger.Debugf(format, args...)
	}
}

func debugf(format string, args ...interface{}) {
	// TODO always write debug to the file
	logger.Debugf(format, args...)
}

func infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func warningf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}
