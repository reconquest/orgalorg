package main

import (
	"os"
	"os/user"
	"strings"
	"sync"

	"github.com/docopt/docopt-go"
	"github.com/kovetskiy/lorg"
	"github.com/seletskiy/hierr"
	"github.com/theairkit/runcmd"
)

const version = "1.0"

const usage = `orgalorg - synchronizing files on many hosts.

First of all, orgalorg will try to acquire global cluster lock by flock'ing
root directory on each host. If at least one flock fails, then orgalorg
will stop, unless '-k' flag is specified.

orgalorg will create tar-archive from specified files, keeping file attributes
and ownerships, then upload archive in parallel to the specified hosts and
unpacks it in the temporary directory. No further actions will be done until
all hosts unpacks the archive.

Then, gunter will be launched with that temporary directory as templates source
directory with empty data file (e.g. no template processing will be done).
No further actions will be taken until gunter finishes processing without
error. All modified files will be logged host-wise in temporary log file.

Then, guntalina will be launched with that log file and will apply actions,
that are specified in guntalina config files (each host may have different
actions).

All output from guntalina will be passed back and returned on stdout.

Finally, all temporary files will be removed from hosts, optionally keeping
backup of the modified files on every host, and global lock is freed.

Restrictions:

    * only one authentication method can be used, and corresponding
      authentication data used for all specified hosts;

Usage:
    orgalorg -h | --help
    orgalorg [options] (-o <host>...|-s)... -S <files>...
    orgalorg [options] (-o <host>...|-s)... --stop-after-lock

Operation modes:
    -S                   Sync.
                          Synchronizes files on the specified hosts via 4-stage
                          process:
                          * global cluster locking;
                          * tar-ing files on local machine, transmitting and
                            unpacking files to the intermediate directory;
                          * launching copy tool (e.g. gunter/rsync);
                          * launching action tool (e.g. guntalina);

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
                          authentication.
                          [default: ~/.ssh/id_rsa]
    -p --password        Use password authentication. Password will be
                          requested on stdin after program start.
                          Excludes '-i' option.
    -x --no-sudo         Do not try to obtain root (via 'sudo -i').
                          By default, orgalorg will try to obtain root and do
                          all actions from root, because it's most common use
                          case. To prevent that behaviour, this option can be
                          used.
    -l --no-lock-abort   Try to obtain global lock, but only print warning if
                          it cannot be done, do not stop execution.
    -r --root <root>     Specify root dir to extract files into.
                          [default: /]
    -u --user <user>     Username used for connecting to all hosts by default.
                          [default: $USER]
    -v --verbose         Print debug information on stderr.
    -V --version         Print program version.

Advanced options:
    --no-preserve-uid    Do not preserve UIDs for transferred files.
    --no-preserve-gid    Do not preserve GIDs for transferred files.
    --backups-dir <dir>  Directory, that will be used on the remote hosts for
                          storing backups. Backups will be stored in the
                          subdirectory, uniquely named with source hostname
                          and timestamp.
                          This option is only useful without '-b'.
                          [default: /var/orgalorg/backups/]
    --temp-dir <dir>     Use specified directory for storing temporary data
                          on each host.
                          [default: /tmp/orgalorg/runs/]
    --stop-after-lock    Will stop right after locking, e.g. will not try to
                          do sync whatsoever. Will keep lock until interrupted.
    --stop-after-upload  Will lock and upload files into specified intermediate
                          directory, then stop.
    --stop-after-copy    Will lock, upload files and then run copy tool to sync
                          files in specified directory with provided files.
    --copy-tool <copy>   Tool for copying files from intermediate directory to
                          the target directory.
                          Tool will accept two arguments: source and target
                          directories.
                          [default: orgalorg-sync]
    --act-tool <act>     Tool for running post-actions after synchronizing
                          files.
                          Tool will accept one argument: file containing files
                          changed while synchronization process.
                          [default: orgalorg-act]
`

const (
	defaultSSHPort = 22
)

var (
	log = lorg.NewLog()
)

func main() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(hierr.Errorf(
			err,
			`can't get current user`,
		))
	}

	log.SetLevel(lorg.LevelDebug)

	usage := strings.Replace(usage, "$USER", currentUser.Username, -1)
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	switch {
	case args["--stop-after-lock"].(bool):
		fallthrough
	case args["-S"].(bool):
		err = synchronize(args)

	}

	log.Fatal(err)
}

func synchronize(args map[string]interface{}) error {
	var (
		lockOnly = args["--stop-after-lock"].(bool)
	)

	_, err := acquireDistributedLock(args)
	if err != nil {
		return hierr.Errorf(
			err,
			`acquiring global cluster lock failed`,
		)
	}

	if lockOnly {
		log.Info("--stop-after-lock was passed, waiting for interrupt...")

		wait := sync.WaitGroup{}
		wait.Add(1)
		wait.Wait()

		os.Exit(0)
	}

	return nil
}

func acquireDistributedLock(
	args map[string]interface{},
) (*distributedLock, error) {
	var (
		defaultUser = args["--user"].(string)
		targets     = args["-o"].([]string)
		rootDir     = args["--root"].(string)
	)

	lock := &distributedLock{}

	for _, host := range targets {
		user, domain, port, err := parseAddress(
			host, defaultUser, defaultSSHPort,
		)
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't parse specified host: '%s'`, host,
			)
		}
		runner, err := runcmd.NewLocalRunner()
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't create runner for host: '%s'`, host,
			)
		}

		err = lock.addNodeRunner(runner, user, domain, port)
		if err != nil {
			return nil, hierr.Errorf(
				err,
				`can't add host to the global cluster lock: '%s'`, host,
			)
		}
	}

	err := lock.acquire(rootDir)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't acquire global cluster lock on %d hosts`,
			len(targets),
		)
	}

	return lock, nil
}
