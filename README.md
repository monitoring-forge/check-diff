# check-diff

Check the difference between the results of the command. check-diff is written to work as a mackerel check plugin.

This is a rewrite of kazeburo/diff-detector with the same functionality.

## Usage

```
% ./check-diff -h
Usage:
  check-diff [OPTIONS] -- command args1 args2

Application Options:
      --identifier= indetify a file store the command result with given string

Help Options:
  -h, --help        Show this help message

```

## Example

```
% echo $(date) > date.txt
% ./check-diff -- cat date.txt 
check-diff OK: first time execution command: 'cat date.txt'
% ./check-diff -- cat date.txt
check-diff OK: no difference: ```Wed Aug 12 00:39:23 JST 2020```

% echo $(date) > date.txt     
% ./check-diff -- cat date.txt
check-diff CRITICAL: found difference: ```@@ -1 +1 @@
-Wed Aug 12 00:39:23 JST 2020
+Wed Aug 12 00:39:40 JST 2020```
```

## mackerel.conf example

```
[plugin.checks.uname-changed]
command = "/usr/local/bin/check-diff -- uname -a"

[plugin.checks.passwd-changed]
command = "/usr/local/bin/check-diff -- cat /etc/passwd"
```

## Install

Please download release page or `mkr plugin install monitoring-forge/check-diff`.