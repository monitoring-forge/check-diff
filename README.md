# check-diff

`check-diff` is a [Mackerel](https://mackerel.io/) check plugin that detects changes in the output of a command by comparing it with the previous execution.

This is a rewrite of [kazeburo/diff-detector](https://github.com/kazeburo/diff-detector) with the same functionality.

## Installation

Download the latest binary from the [releases page](https://github.com/monitoring-forge/check-diff/releases), or install it with the [Mackerel CLI](https://github.com/mackerelio/mkr):

```sh
mkr plugin install monitoring-forge/check-diff
```

## Usage

```
% ./check-diff -h
Usage:
  check-diff [OPTIONS] -- command args1 args2

Application Options:
  -w, --warn        Set the error level to warning
      --identifier= Identify the file used to store the command result with the given string
      --workdir=    Set the working directory
  -v, --version     Show version

Help Options:
  -h, --help        Show this help message
```

### Basic usage

Place `--` between options and the command to run. The plugin executes the command, compares its output with the previous run, and reports the result.

```sh
./check-diff -- cat /path/to/file
```

### First run

On the first execution, `check-diff` stores the command output and returns `OK` because there is nothing to compare with.

```sh
% ./check-diff -- cat date.txt
check-diff OK: first time execution command: 'cat date.txt'
```

### Detecting differences

On subsequent runs, `check-diff` compares the current output with the stored output. If the output is unchanged, it returns `OK`; otherwise, it returns `CRITICAL` (or `WARNING` if `-w` is used).

```sh
% echo "Wed Aug 12 00:39:23 JST 2020" > date.txt
% ./check-diff -- cat date.txt
check-diff OK: no difference: ```Wed Aug 12 00:39:23 JST 2020```

% echo "Wed Aug 12 00:39:40 JST 2020" > date.txt
% ./check-diff -- cat date.txt
check-diff CRITICAL: found difference: ```@@ -1 +1 @@
-Wed Aug 12 00:39:23 JST 2020
+Wed Aug 12 00:39:40 JST 2020```
```

## Options

### `-w`, `--warn`

Report differences as `WARNING` instead of `CRITICAL`.

```sh
./check-diff -w -- cat /path/to/file
```

### `--identifier`

Use a custom identifier for the file that stores the command result. This is useful when you want to run the same command multiple times with different identities.

```sh
./check-diff --identifier=production -- cat /etc/passwd
./check-diff --identifier=staging -- cat /etc/passwd
```

### `--workdir`

Change the directory where the previous output is stored. By default, the plugin uses Mackerel's plugin working directory.

```sh
./check-diff --workdir=/var/lib/check-diff -- cat /etc/passwd
```

## Mackerel configuration example

Add entries to `mackerel-agent.conf` to monitor file or command output changes:

```conf
[plugin.checks.uname-changed]
command = "/usr/local/bin/check-diff -- uname -a"

[plugin.checks.passwd-changed]
command = "/usr/local/bin/check-diff -- cat /etc/passwd"
```

## License

See [LICENSE](LICENSE).
