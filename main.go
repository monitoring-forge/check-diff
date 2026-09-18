package main

import (
	"fmt"
	"os"

	"github.com/itchyny/gojq"
	"github.com/mackerelio/checkers"
	"github.com/mackerelio/golib/pluginutil"
	"github.com/monitoring-forge/flagrun"
)

var version string

type Opt struct {
	Args            []string
	Command         string
	Identifier      string `long:"identifier" description:"identify the file used to store the command result with the given string"`
	Warn            bool   `short:"w" long:"warn" description:"Set the error level to warning"`
	Workdir         string `long:"workdir" description:"Set the working directory"`
	Version         bool   `short:"v" long:"version" description:"Show version"`
	JSON            bool   `short:"j" long:"json" description:"Calculate the diff treating the command output as JSON"`
	JSONIgnoreQuery string `long:"json-ignore" description:"jq query to ignore structures in the JSON output. To use this option, the json option must be enabled"`
	jq              *gojq.Query
}

func (opt *Opt) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("a command is required")
	}
	opt.Args = []string{}
	opt.Command = args[0]
	if len(args) > 1 {
		opt.Args = args[1:]
	}

	if opt.Workdir == "" {
		opt.Workdir = pluginutil.PluginWorkDir()
	}

	if opt.JSONIgnoreQuery != "" && !opt.JSON {
		return fmt.Errorf("the --json-ignore option requires the --json option to be enabled")
	}
	if opt.JSONIgnoreQuery != "" {
		query, err := gojq.Parse(opt.JSONIgnoreQuery)
		if err != nil {
			return fmt.Errorf("failed to parse JSON ignore query: %w", err)
		}
		opt.jq = query
	}
	return nil
}

func (opt *Opt) Run(_ []string) *checkers.Checker {
	return opt.check()
}

func main() {
	opt := &Opt{}
	os.Exit(flagrun.Check(opt, flagrun.Version(version), flagrun.Validator(opt.Validate), flagrun.ArgsRequired()))
}
