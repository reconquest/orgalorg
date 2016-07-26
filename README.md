# orgalorg [![goreport](https://goreportcard.com/badge/github.com/reconquest/orgalorg)](https://goreportcard.com/report/github.com/reconquest/orgalorg) [![MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/reconquest/orgalorg/master/LICENSE)

<p align="center">
<b>orgalorg can run command and upload files in parallel by SSH on many hosts</b>
</p>

<p align="center">
<img src="https://raw.githubusercontent.com/reconquest/orgalorg/master/demo.gif" />
</p>

# Features

* Zero-configuration. No config files. Everything is done via command line
  flags.

* Running SSH commands or shell scripts on any number of hosts in parallel. All
  output from nodes will be returned back, keeping stdout and stderr streams
  mapping of original commands.

* Synchronizing files and directories across cluster with prior global cluster
  locking.
  After synchronization is done, arbitrary command can be evaluated.

* Synchronizing files and directories with subsequent run of complex multi-step
  scenario with steps synchronization across cluster.

* User-friendly progress indication.

* Both strict or loose modes of failover to be sure that everything will either
  fail on any error or try to complete, no matter of what.

* Interactive password authentication as well as SSH public key authentication.

* Ability to run commands through `sudo`.

* Grouped mode of output, so stdout and stderr from nodes will be grouped by
  node name. Alternatively, output can be returned as soon as node returns
  something.

# Installation

## go get

```bash
go get github.com/reconquest/orgalorg
```

# Alternatives

* ansible: intended to apply complex DSL-based scenarios of actions;  
  orgalorg aimed only on running commands and synchronizing files in parallel.
  orgalorg can accept target hosts list on stdin and can provide realtime
  output from commands, which ansible can't do (like running `tail -f`).
  orgalorg also uses same argument semantic as `ssh`:  
  `orgalorg ... -C tail -f '/var/log/*.log'` will do exactly the same.

* clusterssh / cssh: will open number of xterm terminals to all nodes.  
  orgalorg intended to use in batch mode, no GUI is assumed. orgalorg, however,
  can be used in interactive mode (see example section below).

* pssh: buggy, uses binary ssh, which is not resource efficient.  
  orgalorg uses native SSH protocol implementation, so safe and fast to use
  on thousand of nodes.

* dsh / gsh / pdsh: not maintained.

# Example usages

`-o <host>...` in later examples will mean any supported combination of
host-specification arguments, like  
`-o node1.example.com -o node2.example.com`.

## Evaluating command on hosts in parallel

```bash
orgalorg -o <host>... -C uptime
```

## Evaluating command on hosts given by stdin

`axfr` is a tool of your choice for retrieving domain information from your
infrastructure DNS.

```bash
axfr | grep phpnode | orgalorg -s -C uptime
```

## Evaluate command under root (passwordless sudo required)

```bash
orgalorg -o <host>... -x -C whoami
```

## Tailing logs from many hosts in realtime

```bash
orgalorg -o <host>... -C tail -f /var/log/syslog
```

## Copying SSH public key for remote authentication

```bash
orgalorg -o <host>... -p -i ~/.ssh/id_rsa.pub -C tee -a ~/.ssh/authorized_keys
```

## Synchronizing configs and then reloading service (like nginx)

```bash
orgalorg -o <host>... -xn 'systemctl reload nginx' -S /etc/nginx.conf
```

## Evaluating shell script

```bash
orgalorg -o <host>... -i script.bash -C bash
```

## Install package on all nodes and get combined output from each node

```bash
orgalorg -o <host>... -lx -C pacman -Sy my-package --noconfirm
```

## Evaluating shell oneliner

```bash
orgalorg -o <host>... -C sleep '$(($RANDOM % 10))' '&&' echo done
```

## Running poor-man interactive parallel shell

```bash
orgalorg -o <host>... -i /dev/stdin -C bash -s
```

## Obtaining global cluster lock

```bash
orgalorg -o <host>... -L
```

Next orgalorg calls will fail with message, that lock is already acquired,
until first instance will be stopped.

Useful for setting cluster into maintenance state.

## Obtaining global cluster lock on custom directory

```bash
orgalorg -o <host>... -L -r /etc
```

# Description

orgalorg provides easy way of synchronizing files across cluster and running
arbitrary SSH commands.

orgalorg works through SSH & tar, so no unexpected protocol errors will arise.

In default mode of operation (lately referred as sync mode) orgalorg will
perform steps in the following order:

1. Acquire global cluster lock (check more detailed info above).
2. Create, upload and extract specified files in streaming mode to the
   specified nodes into temporary run directory.
3. Start synchronization tool on each node, that should relocate files from
   temporary run directory to the destination.

So, orgalorg expected to work with third-party synchronization tool, that
will do actual files relocation and can be quite intricate, **but orgalorg can
work without that tool and perform simple files sync (more on this later)**.


## Global Cluster Lock

Before doing anything else orgalorg will perform global cluster lock. That lock
is acquired atomically, and no other orgalorg instance can acquire lock if it
is already acquired.

Locking is done via flock'ing specified file or directory on each of target
nodes, and will fail, if flock fails on at least one node.

Directory can be used as lock target as well as ordinary file. `--lock-file`
can be used to specify lock target different from `/`.

After acquiring lock, orgalorg will run heartbeat process, which will check,
that lock is still intact. By default, that check will be performed every 10
seconds. If at least one heartbeat is failed, then orgalorg will abort entire
sync procedure.

User can stop there by using `--lock` or `-L` flag, effectively transform
orgalorg to the distributed locking tool.


## File Upload

Files will be sent from local node to the amount of specified nodes.

orgalorg will perform streaming transfer, so it's safe to synchronize large
files without major memory consumption.

By default, orgalorg will upload files to the temporary run directory. That
behaviour can be changed by using `--root` or `-r` flag. Then, files will be
uploaded to the specified directory.

User can specify `--upload` or `-U` flag to transform orgalorg to the simple
file upload tool. In that mode orgalorg will upload files to the specified
directory and then exit.

orgalorg preserves all file attributes while transfer as well as user and group
IDs. That behaviour can be changed by using `--no-preserve-uid` and
`--no-preseve-gid` command line options.

By default, orgalorg will keep source file paths as is, creating same directory
layout on the target nodes. E.g., if orgalorg told to upload file `a` while
current working directory is `/b/c/`, orgalorg will upload file to the
`<root>/b/c/a` on the remote nodes. That behaviour can be changed by
specifying `--relative` or `-e` flag. Then, orgalorg will not preserve source
file base directory.

orgalorg will try to upload files under specified user (current user by
default). However, if user has `NOPASSWD` record in the sudoers file on the
remote nodes, `--sudo` or `-x` can be used to elevate to root before uploading
files. It makes possible to login to the remote nodes under normal user and
rewrite system files.


## Synchronization Tool

After file upload orgalorg will execute synchronization tool
(`/usr/lib/orgalorg/sync`). That tool is expected to relocate synced files from
temporary directory to the target directory. However, that tool can perform
arbitrary actions, like reloading system services.

To specify custom synchronization tool user can use `--sync-cmd` or `-n` flag.
Full shell syntax is supported in the argument to that option.

Tool is also expected to communicate with orgalorg using sync protocol
(described below), however, it's not required. If not specified, orgalorg will
communicate with that tool using stdin/stdout streams. User can change that
behaviour using `--simple` or `-m` flag, which will cause orgalorg to treat
specified sync tool as simple shell command. User can even provide stdin
to that program by using `--stdin` or `-i` flag.

Tool can accept number of arguments, which can be specified  by using `-g` or
`--arg` flags.


# Synchronization Protocol

orgalorg will communicate with given sync tool using special sync protocol,
which gives possibility to perform some actions with synchronization across
entire cluster.

orgalorg will start sync tool as it specified in the command line, without
any modification.

After start, orgalorg will communicate with running sync tool using stdin
and stdout streams. stderr will be passed to user untouched.

All communication messages should be prefixed by special prefix, which is
send by orgalorg in the hello message. All lines on stdout that are not match
given prefix will be printed as is, untouched.

Communication begins from the hello message.


## Protocol

### HELLO

`orgalorg -> sync tool`

```
<prefix> HELLO
```

Start communication session. All further messages should be prefixed with given
prefix.

### NODE

`orgalorg -> sync tool`

```
<prefix> NODE <node>
```

orgalorg will send node list to the sync tools on each running node.

### START

`orgalorg -> sync tool`

```
<prefix> START
```

Start messages will be sent at the end of the nodes list and means that sync
tool can start doing actions.

### SYNC

`sync tool -> orgalorg`

```
<prefix> SYNC <description>
```

Sync tool can send sync messages after some steps are done to be sure, that
every node in cluster are performing steps gradually, in order.

When orgalorg receives sync message, it will be broadcasted to every connected
sync tool.

### SYNC (broadcasted)

`orgalorg -> sync tool`

```
<prefix> SYNC <node> <description>
```

orgalorg will retransmit incoming sync message from one node to every connected
node (including node, that is sending sync).

Sync tools can wait for specific number of the incoming sync messages to
continue to the next step of execution process.

## Example

`<-` are outgoing messages (from orgalorg to sync tools).

```
<- ORGALORG:132464327653 HELLO
<- ORGALORG:132464327653 NODE [user@node1:22]
<- ORGALORG:132464327653 NODE [user@node2:1234]
<- ORGALORG:132464327653 START
-> (from node1) ORGALORG:132464327653 SYNC phase 1 completed
<- ORGALORG:132464327653 SYNC [user@node1:22] phase 1 completed
-> (from node2) ORGALORG:132464327653 SYNC phase 1 completed
<- ORGALORG:132464327653 SYNC [user@node2:1234] phase 1 completed
```

# Testing

To run tests it's enough to:

```
./run_tests
```

## Requirements

Testcases are run through [tests.sh](https://github.com/reconquest/tests.sh)
library.

For every testcase new set of temporary containers will be initialized through
[hastur](https://github.com/seletskiy/hastur), so `systemd` is required for
running test suite.

orgalorg testcases are close to reality as possible, so orgalorg will really
connect via SSH to cluster of containers in each testcase.

## Coverage

Run following command to calculate total coverage (available after running
testsuite):

```bash
make coverage.total
```

Current coverage level is something about **85%**.
