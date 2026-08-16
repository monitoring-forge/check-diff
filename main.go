package main

import (
	"os"

	"github.com/mackerelio/checkers"
	"github.com/mackerelio/golib/pluginutil"
	"github.com/monitoring-forge/flagrun"
)

var version string

type Opt struct {
	Args       []string
	Command    string
	Identifier string `long:"identifier" description:"identify the file used to store the command result with the given string"`
	Warn       bool   `short:"w" long:"warn" description:"Set the error level to warning"`
	Workdir    string `long:"workdir" description:"Set the working directory"`
	Version    bool   `short:"v" long:"version" description:"Show version"`
}

func (opt *Opt) Run(args []string) *checkers.Checker {
	opt.Args = []string{}
	opt.Command = args[0]
	if len(args) > 1 {
		opt.Args = args[1:]
	}

	if opt.Workdir == "" {
		opt.Workdir = pluginutil.PluginWorkDir()
	}

	return opt.check()
}

func main() {
	os.Exit(flagrun.Check(&Opt{}, flagrun.Version(version), flagrun.ArgsRequired()))
}
