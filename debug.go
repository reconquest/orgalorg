package main

import (
	"sync"
)

var (
	// wait until lorg will be thread-safe
	loggerLock = sync.Mutex{}
)

func tracef(format string, args ...interface{}) {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	// TODO always write debug to the file
	if verbose >= verbosityTrace {
		logger.Debugf(format, args...)
	}
}

func debugf(format string, args ...interface{}) {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	// TODO always write debug to the file
	logger.Debugf(format, args...)
}

func infof(format string, args ...interface{}) {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	logger.Infof(format, args...)
}
