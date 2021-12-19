package main

import (
	"bufio"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/kovetskiy/lorg"
	"github.com/mattn/go-shellwords"
	"github.com/reconquest/barely"
	"github.com/reconquest/hierr-go"
	"github.com/reconquest/loreley"
	"github.com/reconquest/runcmd"
	"golang.org/x/crypto/ssh/agent"
)

var version = "[manual build]"

const usage = `orgalorg - files synchronization on many hosts.

First of all, orgalorg will try to acquire global cluster lock by flock'ing
file, specified by '--lock-file' on each host. If at least one flock fails,
then orgalorg will stop, unless '-t' flag is specified.

orgalorg will create tar-archive from specified files, keeping file attributes
and ownerships, then upload archive in parallel to the specified hosts and
unpack it in the temporary directory (see '-r'). No further actions will be
done until all hosts will unpack the archive.

If '-S' flag specified, then sync command tool will be launched after upload
(see '--sync-cmd'). Sync command tool can send stdout and stderr back to the
orgalorg, but it needs to be compatible with following procotol.

First of all, sync command tool and orgalorg communicate through stdout/stdin.
All lines, that are not match protocol will be printed untouched.

orgalorg first send hello message to the each running node, where '<prefix>'
is an unique string

  <prefix> HELLO

All consequent communication must be prefixed by that prefix, followed by
space.

Then, orgalorg will pass nodes list to each running node by sending 'NODE'
commands, where '<node>' is unique node identifier:

  <prefix> NODE <node> [CURRENT]

'CURRENT' will present next to the node record which currently executes sync
tool.

After nodes list is exchausted, orgalorg will send 'START' marker, that means
sync tool may proceed with execution.

  <prefix> START

Then, sync command tool can reply with 'SYNC' messages, that will be
broadcasted to all connected nodes by orgalorg:

  <prefix> SYNC <description>

Broadcasted sync message will contain source node:

  <prefix> SYNC <node> <description>

Each node can decide, when to wait synchronizations, based on amount of
received sync messages.

Usage:
  orgalorg -h | --help
  orgalorg [options] [-v]... (-o <host>...|-s) -r= -U <files>...
  orgalorg [options] [-v]... (-o <host>...|-s) [-r=] [-g=]... -S [<files>...|--no-upload]
  orgalorg [options] [-v]... (-o <host>...|-s) [-r=] -C [--] <command>...
  orgalorg [options] [-v]... (-o <host>...|-s) -L

Operation mode options:
  -S --sync               Sync.
                           Synchronizes files on the specified hosts via 3-stage
                           process:
                           * global cluster locking (use -L to stop here);
                           * tar-ing files on local machine, transmitting and
                           unpacking files to the intermediate directory
                           (-U to stop here);
                           * launching sync command tool such as gunter;
  -L --lock               Will stop right after locking, e.g. will not try to
                           do sync whatsoever. Will keep lock until interrupted.
  -U --upload             Upload files to specified directory and exit.
  -C --command            Run specified command on all hosts and exit.

Required options:
  -o --host <host>        Target host in format [<username>@]<domain>[:<port>].
                           If value is started from '/' or from './', then it's
                           considered file which should be used to read hosts
                           from.
  -s --read-stdin         Read hosts from stdin in addition to other flags.

Options:
  -h --help               Show this help.
  -k --key <identity>     Identity file (private key), which will be used for
                           authentication. This is default way of
                           authentication.
                           [default: $HOME/.ssh/id_rsa]
  -p --password           Enable password authentication.
                           Exclude '-k' option.
                           Interactive TTY is required for reading password.
  -x --sudo               Obtain root via 'sudo -n'.
                           By default, orgalorg will not obtain root and do
                           all actions from specified user. To change that
                           behaviour, this option can be used.
  -y --no-lock            Do not lock at all.
  -t --no-lock-fail       Try to obtain global lock, but only print warning if
                           it cannot be done, do not stop execution.
  -w --no-conn-fail       Skip unreachable servers whatsoever.
  -r --root <root>        Specify root dir to extract files into.
                           By default, orgalorg will create temporary directory
                           inside of '$ROOT'.
                           Removal of that directory is up to sync tool.
  -u --user <user>        Username used for connecting to all hosts by default.
                           [default: $USER]
  -i --stdin <file>       Pass specified file as input for the command.
  -l --serial             Run commands in serial mode, so they output will not
                           interleave each other. Only one node is allowed to
                           output, all other nodes will wait that node to
                           finish.
  -q --quiet              Be quiet, in command mode do not use prefixes.
  -v --verbose            Print debug information on stderr.
  -V --version            Print program version.

Advanced options:
  --lock-file <path>      File to put lock onto. If not specified, value of '-r'
                           will be used. If '-r' is not specified too, then
                           use "$LOCK" as lock file.
  -e --relative           Upload files by relative path. By default, all
                           specified files will be uploaded on the target
                           hosts by absolute paths, e.g. if you running
                           orgalorg from '/tmp' dir with argument '-S x',
                           then file will be uploaded into '/tmp/x' on the
                           remote hosts. That option switches off that
                           behavior.
  -n --sync-cmd <cmd>     Run specified sync command tool on each remote node.
                           Orgalorg will communicate with sync command tool via
                           stdin. See 'Protocol commands' below.
                           [default: /usr/lib/orgalorg/sync "${@}"]
  -g --arg <arg>          Arguments to pass untouched to the sync command tool.
                           No modification will be done to the passed arg, so
                           take care about escaping.
  -m --simple             Treat sync command as simple tool, which is not
                           support specified protocol messages. No sync
                           is possible in that case and all stdout and stderr
                           will be passed untouched back to the orgalorg.
  --shell <shell>         Use following shell wrapper. '{}' will be replaced
                           with properly escaped command. If empty, then no
                           shell wrapper will be used. If any args are given
                           using '-g', they will be appended to shell
                           invocation.
                           [default: ` + defaultRemoteExecutionShell + `]
  -d --threads <n>        Set threads count which will be used for connection,
                           locking and execution commands.
                           [default: 16].
` + helpNoPreserveUidGuid + `  --no-upload             Do not upload files while syncing.

Output format and colors options:
    --json                Output everything in line-by-line JSON format,
                           printing objects with fields:
                           * 'stream' = 'stdout' | 'stderr';
                           * 'node' = <node-name> | null (if internal output);
                           * 'body' = <string>
    --bar-format <f>      Format for the status bar.
                           Full Go template syntax is available with delims
                           of '{' and '}'.
                           See https://github.com/reconquest/barely for more
                           info.
                           For example, run orgalorg with '-vv' flag.
                           Two embedded themes are available by their names:
                           ` + themeDark + ` and ` + themeLight + `
                           [default: ` + themeDefault + `]
    --log-format <f>      Format for the logs.
                           See https://github.com/reconquest/colorgful  for more
                           info.
                           [default: ` + themeDefault + `]
    --colors-dark         Set all available formats to predefined dark theme.
    --colors-light        Set all available formats to predefined light theme.
    --color <mode>        Specify, whether to use colors:
                           * never - disable colors;
                           * auto - use colors only when TTY presents.
                           * always - always use colorized output.
                           [default: auto]

Timeout options:
  -c --conn-timeout <ms>  Remote host connection timeout in milliseconds.
                           [default: 10000]
  -j --send-timeout <ms>  Remote host connection data sending timeout in
                           milliseconds. [default: 60000]
                           NOTE: send timeout will be also used for the
                           heartbeat messages, that orgalorg and connected nodes
                           exchanges through synchronization process.
  -z --recv-timeout <ms>  Remote host connection data receiving timeout in
                           milliseconds. [default: 60000]
  -a --keep-alive <ms>    How long to keep connection keeped alive after session
                           ends. [default: 10000]
`

const (
	defaultSSHPort = 22

	// heartbeatTimeoutCoefficient will be multiplied to send timeout and
	// resulting value will be used as time interval between heartbeats.
	heartbeatTimeoutCoefficient = 0.8

	runsDirectory = "/var/run/orgalorg/"

	defaultLockFile = "/"

	defaultRemoteExecutionShell = "bash -c '{}'"
)

var (
	sshPasswordPrompt   = "Password: "
	sshPassphrasePrompt = "Key passphrase: "
)

var (
	logger  = lorg.NewLog()
	verbose = verbosityNormal
	format  = outputFormatText

	pool      *threadPool
	statusbar *barely.StatusBar
)

var exit = os.Exit

func main() {
	args := parseArgs()

	verbose = parseVerbosity(args)

	setLoggerVerbosity(verbose, logger)

	format = parseOutputFormat(args)

	setLoggerOutputFormat(logger, format)

	loreley.Colorize = parseColorMode(args)

	loggerStyle, err := getLoggerTheme(parseTheme("log", args))
	if err != nil {
		fatalln(hierr.Errorf(
			err,
			`can't use given logger style`,
		))
	}

	setLoggerStyle(logger, loggerStyle)

	poolSize, err := parseThreadPoolSize(args)
	if err != nil {
		errorln(hierr.Errorf(
			err,
			`--threads given invalid value`,
		))
	}

	pool = newThreadPool(poolSize)

	setupInteractiveMode(args)

	switch {
	case args["--upload"].(bool):
		fallthrough
	case args["--lock"].(bool):
		fallthrough
	case args["--sync"].(bool):
		err = handleSynchronize(args)
	case args["--command"].(bool):
		err = handleEvaluate(args)
	}

	if err != nil {
		fatalln(err)
	}

	clearStatus()
}

func parseArgs() map[string]interface{} {
	usage, err := formatUsage(string(usage))
	if err != nil {
		fatalln(hierr.Errorf(
			err,
			`can't format usage`,
		))
	}

	args, err := docopt.Parse(usage, nil, true, version, true)
	if err != nil {
		panic(err)
	}

	return args
}

func formatUsage(template string) (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't get current user`,
		)
	}

	usage := template

	usage = strings.Replace(usage, "$USER", currentUser.Username, -1)
	usage = strings.Replace(usage, "$HOME", currentUser.HomeDir, -1)
	usage = strings.Replace(usage, "$ROOT", runsDirectory, -1)
	usage = strings.Replace(usage, "$LOCK", defaultLockFile, -1)

	return usage, nil
}

func handleEvaluate(args map[string]interface{}) error {
	var (
		stdin, hasStdin = args["--stdin"].(string)
		rootDir, _      = args["--root"].(string)
		sudo            = args["--sudo"].(bool)
		shell           = args["--shell"].(string)
		serial          = args["--serial"].(bool)

		command = args["<command>"].([]string)
	)

	canceler := sync.NewCond(&sync.Mutex{})

	cluster, err := connectAndLock(args, canceler)
	if err != nil {
		return err
	}

	runner := &remoteExecutionRunner{
		shell:     shell,
		sudo:      sudo,
		command:   command,
		directory: rootDir,
		serial:    serial,
		term:      !hasStdin,
	}

	return run(cluster, runner, stdin)
}

func run(
	cluster *distributedLock,
	runner *remoteExecutionRunner,
	stdin string,
) error {
	execution, err := runner.run(cluster, nil)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't run remote execution on %d nodes`,
			len(cluster.nodes),
		)
	}

	if stdin != "" {
		var inputFile *os.File

		inputFile, err = os.Open(stdin)
		if err != nil {
			return hierr.Errorf(
				err,
				`can't open file for passing as stdin: '%s'`,
				stdin,
			)
		}

		defer inputFile.Close()

		_, err = io.Copy(execution.stdin, inputFile)
		if err != nil {
			return hierr.Errorf(
				err,
				`can't copy input file to the execution processes`,
			)
		}

	}

	debugf(`commands are running, waiting for finish`)

	err = execution.stdin.Close()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't close stdin`,
		)
	}

	err = execution.wait()
	if err != nil {
		return hierr.Errorf(
			err,
			`remote execution failed, because one of `+
				`command has been exited with non-zero exit `+
				`code (or timed out) at least on one node`,
		)
	}

	return nil
}

func handleSynchronize(args map[string]interface{}) error {
	var (
		stdin, _   = args["--stdin"].(string)
		rootDir, _ = args["--root"].(string)
		lockOnly   = args["--lock"].(bool)
		uploadOnly = args["--upload"].(bool)
		relative   = args["--relative"].(bool)

		isSimpleCommand = args["--simple"].(bool)

		commandString = args["--sync-cmd"].(string)
		commandArgs   = args["--arg"].([]string)

		shell = args["--shell"].(string)

		sudo   = args["--sudo"].(bool)
		serial = args["--serial"].(bool)

		fileSources = args["<files>"].([]string)

		noUpload = args["--no-upload"].(bool)
	)

	var (
		filesList = []file{}

		err error
	)

	if !noUpload {
		if !lockOnly {
			debugf(`building files list from %d sources`, len(fileSources))
			filesList, err = getFilesList(relative, fileSources...)
			if err != nil {
				return hierr.Errorf(
					err,
					`can't build files list`,
				)
			}

			debugf(`file list contains %d files`, len(filesList))
			tracef(`files to upload: %+v`, filesList)
		}
	}

	canceler := sync.NewCond(&sync.Mutex{})

	cluster, err := connectAndLock(args, canceler)
	if err != nil {
		return err
	}

	if lockOnly {
		warningf("-L|--lock was passed, waiting for interrupt...")

		canceler.L.Lock()
		canceler.Wait()
		canceler.L.Unlock()

		return nil
	}

	if !noUpload {
		err = upload(args, cluster, filesList)
		if err != nil {
			return hierr.Errorf(
				err,
				`can't upload files on the remote nodes`,
			)
		}

		tracef(`upload done`)
	}

	if uploadOnly {
		return nil
	}

	tracef(`starting sync tool`)

	command, err := shellwords.NewParser().Parse(commandString)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't parse sync tool command: '%s'`,
			commandString,
		)
	}

	runner := &remoteExecutionRunner{
		shell:     shell,
		sudo:      sudo,
		command:   command,
		args:      commandArgs,
		directory: rootDir,
		serial:    serial,
	}

	if isSimpleCommand {
		return run(cluster, runner, stdin)
	}

	err = runSyncProtocol(cluster, runner)
	if err != nil {
		return hierr.Errorf(
			err,
			`failed to run sync command`,
		)
	}

	return nil
}

func upload(
	args map[string]interface{},
	cluster *distributedLock,
	filesList []file,
) error {
	var (
		rootDir, _ = args["--root"].(string)
		sudo       = args["--sudo"].(bool)

		preserveUID = !args["--no-preserve-uid"].(bool)
		preserveGID = !args["--no-preserve-gid"].(bool)

		serial = args["--serial"].(bool)
	)

	if rootDir == "" {
		rootDir = filepath.Join(runsDirectory, generateRunID())
	}

	debugf(`file upload started into: '%s'`, rootDir)

	receivers, err := startArchiveReceivers(cluster, rootDir, sudo, serial)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't start archive receivers on the cluster`,
		)
	}

	err = archiveFilesToWriter(
		receivers.stdin,
		filesList,
		preserveUID,
		preserveGID,
	)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't archive files and send to the remote nodes`,
		)
	}

	tracef(`waiting file upload to finish`)

	err = receivers.stdin.Close()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't close archive receiver stdin`,
		)
	}

	err = receivers.wait()
	if err != nil {
		return hierr.Errorf(
			err,
			`archive upload failed`,
		)
	}

	return nil
}

func connectAndLock(
	args map[string]interface{},
	canceler *sync.Cond,
) (*distributedLock, error) {
	var (
		hosts = args["--host"].([]string)

		sendTimeout = args["--send-timeout"].(string)
		defaultUser = args["--user"].(string)

		askPassword = args["--password"].(bool)
		fromStdin   = args["--read-stdin"].(bool)

		rootDir, _    = args["--root"].(string)
		sshKeyPath, _ = args["--key"].(string)
		lockFile, _   = args["--lock-file"].(string)

		noConnFail = args["--no-conn-fail"].(bool)
		noLockFail = args["--no-lock-fail"].(bool)

		noLock = args["--no-lock"].(bool)
	)

	addresses, err := parseAddresses(hosts, defaultUser, fromStdin)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't parse all specified addresses`,
		)
	}

	timeouts, err := makeTimeouts(args)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't parse SSH connection timeouts`,
		)
	}

	runners, err := createRunnerFactory(timeouts, sshKeyPath, askPassword)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't create runner factory`,
		)
	}

	debugf(`using %d threads`, pool.size)

	debugf(`connecting to %d nodes`, len(addresses))

	if lockFile == "" {
		if rootDir == "" {
			lockFile = defaultLockFile
		} else {
			lockFile = rootDir
		}
	}

	heartbeatMillisecondsBase, err := strconv.Atoi(sendTimeout)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't use --send-timeout as heartbeat timeout`,
		)
	}

	heartbeatMilliseconds := time.Duration(
		float64(heartbeatMillisecondsBase)*heartbeatTimeoutCoefficient,
	) * time.Millisecond

	cluster, err := connectToCluster(
		lockFile,
		runners,
		addresses,
		noLock,
		noLockFail,
		noConnFail,
		func(node *distributedLockNode) {
			heartbeat(heartbeatMilliseconds, node, canceler)
		},
	)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`connecting to cluster failed`,
		)
	}

	if noLock {
		debugf(`connection established to %d nodes`, len(cluster.nodes))
	} else {
		debugf(`global lock acquired on %d nodes`, len(cluster.nodes))
	}

	return cluster, nil
}

func createRunnerFactory(
	timeout *runcmd.Timeout,
	sshKeyPath string,
	askPassword bool,
) (runnerFactory, error) {
	if askPassword {
		password, err := readPassword(sshPasswordPrompt)
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't read password`,
			)
		}

		return createRemoteRunnerFactoryWithPassword(
			string(password),
			timeout,
		), nil
	}

	keyring, err := getSshAgent()
	if err != nil {
		debugf(`failed to connect to a ssh-agent: %s`, err)
	}
	keyListEmpty := true
	if keyring == nil {
		keyring = agent.NewKeyring()
	} else {
		li, err := keyring.List()
		if err != nil {
			debugf(`    could not get agent keys list: %s`, err)
		}
		keyListEmpty = len(li) < 1
		if keyListEmpty {
			debugf(`    agent keys list is empty`)
		}
	}

	if keyListEmpty && sshKeyPath != "" {
		err := readSSHKey(keyring, sshKeyPath)
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't read SSH key: '%s'`,
				sshKeyPath,
			)
		}
	}

	return createRemoteRunnerFactoryWithAgent(
		keyring,
		timeout,
	), nil
}

func parseAddresses(
	hosts []string,
	defaultUser string,
	fromStdin bool,
) ([]address, error) {
	hostsToParse := []string{}

	if fromStdin {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			hostsToParse = append(hostsToParse, scanner.Text())
		}
	}

	for _, host := range hosts {
		if strings.HasPrefix(host, "/") || strings.HasPrefix(host, "./") {
			hostsFile, err := os.Open(host)
			if err != nil {
				return nil, hierr.Errorf(
					err,
					`can't open hosts file: '%s'`,
					host,
				)
			}

			scanner := bufio.NewScanner(hostsFile)
			for scanner.Scan() {
				hostsToParse = append(hostsToParse, scanner.Text())
			}
		} else {
			hostsToParse = append(hostsToParse, host)
		}
	}

	addresses := []address{}

	for _, host := range hostsToParse {
		parsedAddress, err := parseAddress(
			host, defaultUser, defaultSSHPort,
		)
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't parse specified address '%s'`,
				host,
			)
		}

		addresses = append(addresses, parsedAddress)
	}

	return getUniqueAddresses(addresses), nil
}

func setupInteractiveMode(args map[string]interface{}) {
	var (
		_, hasStdin = args["--stdin"].(string)

		barLock = &sync.Mutex{}
	)

	barStyle, err := getStatusBarTheme(parseTheme("bar", args))
	if err != nil {
		errorln(hierr.Errorf(
			err,
			`can't use given status bar style`,
		))
	}

	if loreley.HasTTY(int(os.Stderr.Fd())) {
		statusbar = barely.NewStatusBar(barStyle.Template)
		statusbar.SetLock(barLock)
	} else {
		statusbar = nil

		sshPasswordPrompt = ""
		sshPassphrasePrompt = ""
	}

	if hasStdin && loreley.HasTTY(int(os.Stdin.Fd())) {
		statusbar = nil
	}
}

func generateRunID() string {
	return time.Now().Format("20060102150405.999999")
}
