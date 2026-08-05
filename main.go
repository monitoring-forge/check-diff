package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
	"github.com/mackerelio/golib/pluginutil"
)

var version string
var commit string

const (
	OK = iota
	WARNING
	CRITICAL
	UNKNOWN
)

type Opt struct {
	Args       []string
	Command    string
	Identifier string `long:"identifier" description:"identify the file used to store the command result with the given string"`
	Warn       bool   `short:"w" long:"warn" description:"Set the error level to warning"`
	Workdir    string `long:"workdir" description:"Set the working directory"`
	Version    bool   `short:"v" long:"version" description:"Show version"`
}

func (opt *Opt) cmd(file *os.File) error {
	cmd := exec.Command(opt.Command, opt.Args...)
	var stderr bytes.Buffer
	cmd.Stdout = file
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	if err != nil {
		return fmt.Errorf("%w - %s", err, stderr.String())
	}
	return nil
}

func (opt *Opt) run() *checkers.Checker {

	identifier := strings.Join([]string{
		opt.Identifier,
		opt.Command,
		strings.Join(opt.Args, "-"),
	}, "-")
	hasher := sha256.New()
	hasher.Write([]byte(identifier))

	curUser, err := user.Current()
	if err != nil {
		return checkers.Critical(err.Error())
	}

	prevPath := filepath.Join(opt.Workdir, fmt.Sprintf("check-diff-%s-%x", curUser.Uid, hasher.Sum(nil)))
	newFile, err := os.CreateTemp(opt.Workdir, "check-diff-")
	if err != nil {
		return checkers.Critical(err.Error())
	}

	err = opt.cmd(newFile)
	if err != nil {
		newFile.Close()
		_ = os.Remove(newFile.Name())
		return checkers.Critical(err.Error())
	}

	err = newFile.Close()
	if err != nil {
		return checkers.Critical(err.Error())
	}

	if !fileExists(prevPath) {
		err = os.Rename(newFile.Name(), prevPath)
		if err != nil {
			return checkers.Critical(err.Error())
		}
		if len(opt.Args) > 0 {
			return checkers.Ok(fmt.Sprintf("first time execution command: '%s %s'", opt.Command, strings.Join(opt.Args, " ")))
		}
		return checkers.Ok(fmt.Sprintf("first time execution command: '%s'", opt.Command))
	}

	diff, err := diff(prevPath, newFile.Name())
	if err != nil {
		return checkers.Critical(err.Error())
	}

	err = os.Rename(newFile.Name(), prevPath)
	if err != nil {
		return checkers.Critical(err.Error())
	}

	if diff == "" {
		msg, err := buildNoDifferenceMsg(prevPath)
		if err != nil {
			return checkers.Critical(err.Error())
		}
		return checkers.Ok(msg)
	}

	diffMsg := buildDiffMsg(diff)
	if opt.Warn {
		return checkers.Warning(diffMsg)
	}
	return checkers.Critical(diffMsg)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func main() {
	opt := &Opt{}
	psr := flags.NewParser(opt, flags.HelpFlag|flags.PassDoubleDash)
	psr.Usage = "[OPTIONS] -- command args1 args2"
	args, err := psr.Parse()
	if opt.Version {
		if commit == "" {
			commit = "dev"
		}
		fmt.Printf(
			"%s-%s\n%s/%s, %s, %s\n",
			filepath.Base(os.Args[0]),
			version,
			runtime.GOOS,
			runtime.GOARCH,
			runtime.Version(),
			commit)
		os.Exit(OK)
	} else if flags.WroteHelp(err) {
		fmt.Fprintf(os.Stdout, "%v\n", err)
		os.Exit(OK)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(UNKNOWN)
	} else if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "command is required\n")
		psr.WriteHelp(os.Stderr)
		os.Exit(UNKNOWN)
	}

	opt.Args = []string{}
	opt.Command = args[0]
	if len(args) > 1 {
		opt.Args = args[1:]
	}

	if opt.Workdir == "" {
		opt.Workdir = pluginutil.PluginWorkDir()
	}

	ckr := opt.run()
	ckr.Name = "check-diff"
	ckr.Exit()
}
