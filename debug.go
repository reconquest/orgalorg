package main

func debugf(format string, args ...interface{}) {
	// TODO always write debug to the file
	logger.Debugf(format, args...)
}
