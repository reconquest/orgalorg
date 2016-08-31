package main

import (
	"fmt"
	"strings"

	"github.com/mattn/go-shellwords"
	"github.com/reconquest/hierr-go"
)

var (
	sudoCommand = []string{"sudo", "-n", "-E", "-H"}
)

type remoteExecutionRunner struct {
	command   []string
	args      []string
	shell     string
	directory string
	sudo      bool
	serial    bool
}

func (runner *remoteExecutionRunner) run(
	cluster *distributedLock,
	setupCallback func(*remoteExecutionNode),
) (*remoteExecution, error) {
	commandline := joinCommand(runner.command)

	if runner.directory != "" {
		commandline = fmt.Sprintf("cd %s && { %s; }",
			escapeCommandArgumentStrict(runner.directory),
			commandline,
		)
	}

	if len(runner.shell) != 0 {
		commandline = wrapCommandIntoShell(
			commandline,
			runner.shell,
			runner.args,
		)
	}

	if runner.sudo {
		commandline = joinCommand(sudoCommand) + " " + commandline
	}

	command, err := shellwords.Parse(commandline)
	if err != nil {
		return nil, hierr.Errorf(
			err, "unparsable command line: %s", commandline,
		)
	}

	return runRemoteExecution(cluster, command, setupCallback, runner.serial)
}

func wrapCommandIntoShell(command string, shell string, args []string) string {
	if shell == "" {
		return command
	}

	command = strings.Replace(shell, `{}`, command, -1)

	if len(args) == 0 {
		return command
	}

	escapedArgs := []string{}
	for _, arg := range args {
		escapedArgs = append(escapedArgs, escapeCommandArgumentStrict(arg))
	}

	return command + " _ " + strings.Join(escapedArgs, " ")
}

func joinCommand(command []string) string {
	escapedParts := []string{}

	for _, part := range command {
		escapedParts = append(escapedParts, escapeCommandArgument(part))
	}

	return strings.Join(escapedParts, ` `)
}

func escapeCommandArgument(argument string) string {
	argument = strings.Replace(argument, `'`, `'\''`, -1)

	return argument
}

func escapeCommandArgumentStrict(argument string) string {
	escaper := strings.NewReplacer(
		`\`, `\\`,
		"`", "\\`",
		`"`, `\"`,
		`'`, `'\''`,
		`$`, `\$`,
	)

	escaper.Replace(argument)

	return `"` + argument + `"`
}
