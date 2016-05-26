package main

import (
	"os"
	"os/user"
	"strings"
	"sync"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/kovetskiy/lorg"
	"github.com/seletskiy/hierr"
)

const version = "1.0"

const usage = `orgalorg - synchronizing files on many hosts.

First of all, orgalorg will try to acquire global cluster lock by flock'ing
file, specified by '--lock-file' on each host. If at least one flock fails,
then orgalorg will stop, unless '-t' flag is specified.

orgalorg will create tar-archive from specified files, keeping file attributes
and ownerships, then upload archive in parallel to the specified hosts and
unpacks it in the temporary directory (see '-r'). No further actions will be
done until all hosts unpacks the archive.

If no '-d' flag specified, after upload post-action tool will be launched (
see '--post-action'). Post-action tool can send stdout and stderr back to
the orgalorg, but it need to be prefixed with special prefix, passed in the
first argument. Additionally, post-action tool can separate it's actions into
several phases, which should be synced across all cluster. To do it, tool
should send 'SYNC' message to the stdout as soon as phase is done. Then, tool
can receive 'SYNC <host>' back from orgalorg as many times as there are
messages sent from other hosts. Tool can decide how soon it should continue
based on the ammount received 'SYNC' messages.

Restrictions:

    * only one authentication method can be used, and corresponding
      authentication data used for all specified hosts;

Usage:
    orgalorg -h | --help
    orgalorg [options] (-o <host>...|-s) -S <files>... [-d]
    orgalorg [options] (-o <host>...|-s) -C <command>...
    orgalorg [options] (-o <host>...|-s) (-L | --stop-at-lock)

Operation modes:
    -S --sync            Sync.
                          Synchronizes files on the specified hosts via 4-stage
                          process:
                          * global cluster locking (use -L to stop here);
                          * tar-ing files on local machine, transmitting and
                            unpacking files to the intermediate directory
                            (-d to stop here);
                          * launching post-action tool such as gunter;
    -L --stop-at-lock    Will stop right after locking, e.g. will not try to
                          do sync whatsoever. Will keep lock until interrupted.
    -d --stop-at-upload  Will lock and upload files into specified intermediate
                          directory, then stop.
    -C --command         Run specified command on all hosts.

Required options:
    -o <host>            Target host in format [<username>@]<domain>[:<port>].
    -s                   Read hosts from stdin in addition to other flags.

Options:
    -h --help            Show this help.
    -n --dry-run         Dry run: upload files on hosts and run gunter in dry
                          run mode. No real files will be replaced. Temporary
                          files will be deleted. Guntalina will be launched in
                          dry mode too.
    -b --no-backup       Do not backup of modified files on each target host.
    -k --key <identity>  Identity file (private key), which will be used for
                          authentication. If '-k' option is not used, then
                          password authentication will be used instead.
                          [default: $HOME/.ssh/id_rsa]
    -p --pasword         Enable password authentication.
                          Exclude '-k' option.
    -x --no-sudo         Do not try to obtain root (via 'sudo -i').
                          By default, orgalorg will try to obtain root and do
                          all actions from root, because it's most common use
                          case. To prevent that behaviour, this option can be
                          used.
    -t --no-lock-abort   Try to obtain global lock, but only print warning if
                          it cannot be done, do not stop execution.
    -r --root <root>     Specify root dir to extract files into.
                          [default: /var/run/orgalorg/files/$RUNID]
    -u --user <user>     Username used for connecting to all hosts by default.
                          [default: $USER]
    -v --verbose         Print debug information on stderr.
    -V --version         Print program version.

Advanced options:
    --lock-file <path>   File to put lock onto.
                         [default: /]
    -e --relative        Upload files by relative path. By default, all
                          specified files will be uploaded on the target
                          hosts by absolute paths, e.g. if you running
                          orgalorg from '/tmp' dir with argument '-S x',
                          then file will be uploaded into '/tmp/x' on the
                          remote hosts. That option switches off that
                          behavior.
    --no-preserve-uid    Do not preserve UIDs for transferred files.
    --no-preserve-gid    Do not preserve GIDs for transferred files.
    --post-action <exe>  Run specified post-action tool on each remote node.
                          Post-action tool should accept followin arguments:
                          * string prefix, that should be used to prefix all
                            stdout and stderr from the process; all unprefixed
                            data will be treated as control commands.
                          * path to directory which contain received files.
                          * additional arguments from the '-g' flag.
                          * --
                          * hosts to sync, one per argument.
                          [default: /usr/lib/orgalorg/post-action]
    -g --args <args>     Arguments to pass untouched to the post-action tool.
                          No modification will be done to the passed arg, so
                          take care about escaping.
    -m --simple <exe>    Treat post-action as simple tool, which is not
                          support specified protocol messages. No syncc
                          is possible in that case and all stdout and stderr
                          will be passed untouched back to the orgalorg.
                          Exclude '--post-action'.

Timeout options:
    --conn-timeout <t>   Remote host connection timeout in milliseconds.
                          [default: 10000]
    --send-timeout <t>   Remote host connection data sending timeout in
                          milliseconds. [default: 10000]
    --recv-timeout <t>   Remote host connection data receiving timeout in
                          milliseconds. [default: 10000]
    --keep-alive <t>     How long to keep connection keeped alive after session
                          end in milliseconds. [default: 60000]
Control commands:

    SYNC                 Cause orgalorg to broadcast 'SYNC <host>' to all
                          connected nodes.
`

const (
	defaultSSHPort    = 22
	SSHPasswordPrompt = "Password: "
)

var (
	logger = lorg.NewLog()
)

func main() {
	logger.SetFormat(lorg.NewFormat("* ${time} ${level:[%s]:left} %s"))
	currentUser, err := user.Current()
	if err != nil {
		logger.Fatal(hierr.Errorf(
			err,
			`can't get current user`,
		))
	}

	logger.SetLevel(lorg.LevelDebug)

	usage := strings.Replace(usage, "$USER", currentUser.Username, -1)
	usage = strings.Replace(usage, "$HOME", currentUser.HomeDir, -1)
	usage = strings.Replace(usage, "$RUNID", generateRunID(), -1)
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	switch {
	case args["-L"].(bool):
		// because of docopt
		args["--stop-at-lock"] = true
		fallthrough
	case args["--stop-at-lock"].(bool):
		fallthrough
	case args["-S"].(bool):
		err = synchronize(args)

	}

	if err != nil {
		logger.Fatal(err)
	}
}

func synchronize(args map[string]interface{}) error {
	var (
		SSHKeyPath, _ = args["--key"].(string)
		lockOnly      = args["--stop-at-lock"].(bool)
		fileSources   = args["<files>"].([]string)
		relative      = args["--relative"].(bool)
	)

	addresses, err := parseAddresses(args)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't parse all specified addresses`,
		)
	}

	timeouts, err := makeTimeouts(args)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't parse SSH connection timeouts`,
		)
	}

	filesList := []string{}
	if !lockOnly {
		logger.Info(`building files list`)
		filesList, err = getFilesList(relative, fileSources...)
		if err != nil {
			return hierr.Errorf(
				err,
				`can't obtain files list to sync from localhost`,
			)
		}

		logger.Infof(`file list contains %d files`, len(filesList))
	}

	var runnerFactory runnerFactory

	switch {
	case SSHKeyPath != "":
		runnerFactory = createRemoteRunnerFactoryWithKey(
			SSHKeyPath,
			timeouts,
		)

	default:
		runnerFactory = createRemoteRunnerFactoryWithAskedPassword(
			SSHPasswordPrompt,
			timeouts,
		)
	}

	cluster, err := acquireDistributedLock(args, runnerFactory, addresses)
	if err != nil {
		return hierr.Errorf(
			err,
			`acquiring global cluster lock failed`,
		)
	}

	logger.Infof(`global lock acquired on %d nodes`, len(cluster.nodes))

	if lockOnly {
		logger.Info("--stop-at-lock was passed, waiting for interrupt...")

		wait := sync.WaitGroup{}
		wait.Add(1)
		wait.Wait()

		os.Exit(0)
	}

	receivers, err := startArchiveReceivers(cluster, args)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't start archive receivers on the cluster`,
		)
	}

	logger.Info(`file upload started`)

	err = archiveFilesToWriter(receivers.stdin, filesList)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't archive files and send to the remote nodes`,
		)
	}

	logger.Info(`waiting file upload to finish`)
	err = receivers.wait()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't finish files archive`,
		)
	}

	logger.Info(`upload done`)

	return nil
}

func parseAddresses(args map[string]interface{}) ([]address, error) {
	var (
		defaultUser = args["--user"].(string)
		hosts       = args["-o"].([]string)
	)

	addresses := []address{}

	for _, host := range hosts {
		address, err := parseAddress(
			host, defaultUser, defaultSSHPort,
		)
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't parse specified address '%s'`,
				host,
			)
		}

		addresses = append(addresses, address)
	}

	return uniqAddresses(addresses), nil
}

func generateRunID() string {
	return time.Now().Format("20060102150405.999999")
}
