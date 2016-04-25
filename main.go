package main

import (
	"log"

	"github.com/docopt/docopt-go"
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
    orgalorg [options] (-d <hosts-file-dir>|-f <hosts-file>|-o <host>|-s)...
                       -S <files>...

Required options:
    -o <host>            Target host in format [<username>@]<domain>[:<port>].
    -f <hosts_file>      File to read target hosts from. One host per line.
                          Format for one record same as for flag '-o'.
    -d <hosts-file-dir>  Directory to read hosts file from. It's OK to store
                          symlinks to real files in that directory.
    -s                   Read hosts from stdin in addition to other flags.

Options:
    -h --help            Show this help.
    -n                   Dry run: upload files on hosts and run gunter in dry
                          run mode. No real files will be replaced. Temporary
                          files will be deleted. Guntalina will be launched in
                          dry mode too.
    -b                   Do not backup of modified files on each target host.
    -i <identity>        Identity file (private key), which will be used for
                          authentication.
                          [default: ~/.ssh/id_rsa]
    -p                   Use password authentication. Password will be
                          requested on stdin after program start.
                          Excludes '-i' option.
    -x                   Do not try to obtain root (via 'sudo -i').
                          By default, orgalorg will try to obtain root and do
                          all actions from root, because it's most common use
                          case. To prevent that behaviour, this option can be
                          used.
    -k                   Try to obtain global lock, but only print warning if
                          it cannot be done, do not stop execution.
    -v                   Print debug information on stderr.
    -V                   Print program version.

Advanced options:
    --backups-dir <dir>  Directory, that will be used on the remote hosts for
                          storing backups. Backups will be stored in the
                          subdirectory, uniquely named with source hostname
                          and timestamp.
                          This option is only useful with '-b', which is off
                          by default.
                          [default: /var/orgalorg/backups/]
    --temp-dir <dir>     Use specified directory for storing temporary data
                          on each host.
                          [default: /tmp/orgalorg/runs/]
`

func main() {
	args, err := docopt.Parse(usage, nil, true, "1.0", false)
	if err != nil {
		panic(err)
	}

	log.Printf("main.go:39 %#v", args)
}
