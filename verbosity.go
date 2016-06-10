package main

type (
	verbosity int
)

const (
	verbosityQuiet verbosity = iota
	verbosityNormal
	verbosityDebug
	verbosityTrace
)

func parseVerbosity(args map[string]interface{}) verbosity {
	var (
		quiet = args["--quiet"].(bool)
		level = args["--verbose"].(int)
	)

	if quiet {
		return verbosityQuiet
	}

	if level == 1 {
		return verbosityDebug
	}

	if level > 1 {
		return verbosityTrace
	}

	return verbosityNormal
}
