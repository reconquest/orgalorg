package main

import (
	"fmt"
	"strings"
)

var (
	sudoCommand = []string{"sudo", "-n", "-E", "-H"}
)

type remoteExecutionRunner struct {
	shell     string
	sudo      bool
	command   []string
	args      []string
	directory string
}

func (runner *remoteExecutionRunner) run(
	cluster *distributedLock,
	setupCallback func(*remoteExecutionNode),
) (*remoteExecution, error) {
	command := joinCommand(runner.command)

	if runner.directory != "" {
		command = fmt.Sprintf("cd %s && { %s; }",
			escapeCommandArgumentStrict(runner.directory),
			command,
		)
	}

	if runner.shell != "" {
		command = wrapCommandIntoShell(command, runner.shell, runner.args)
	}

	if runner.sudo {
		command = joinCommand(sudoCommand) + " " + command
	}

	return runRemoteExecution(cluster, command, setupCallback)
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
	argument = strings.Replace(argument, `\`, `\\`, -1)
	argument = strings.Replace(argument, ` `, `\ `, -1)

	return argument
}

func escapeCommandArgumentStrict(argument string) string {
	argument = strings.Replace(argument, `\`, `\\`, -1)
	argument = strings.Replace(argument, "`", "\\`", -1)
	argument = strings.Replace(argument, `"`, `\"`, -1)
	argument = strings.Replace(argument, `$`, `\$`, -1)

	return `"` + argument + `"`
}
