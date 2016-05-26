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
		quiet   = args["--quiet"].(bool)
		verbose = args["--verbose"].(int)
	)

	if quiet {
		return verbosityQuiet
	}

	if verbose == 1 {
		return verbosityDebug
	}

	if verbose > 1 {
		return verbosityTrace
	}

	return verbosityNormal
}
