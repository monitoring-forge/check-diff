package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cubicdaiya/gonp"
	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
	"github.com/mackerelio/golib/pluginutil"
)

var version string
var commit string

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
		return fmt.Errorf("%s - %s", err, stderr.String())
	}
	return nil
}

func (opt *Opt) run() *checkers.Checker {

	hasher := sha256.New()
	hasher.Write([]byte(opt.Identifier))
	hasher.Write([]byte("-"))
	hasher.Write([]byte(opt.Command))
	hasher.Write([]byte("-"))
	for _, v := range opt.Args {
		hasher.Write([]byte(v))
		hasher.Write([]byte("-"))
	}
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
		os.Remove(newFile.Name())
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
		msg := ""
		if len(opt.Args) > 0 {
			msg = fmt.Sprintf("first time execution command: '%s %s'", opt.Command, strings.Join(opt.Args, " "))
		} else {
			msg = fmt.Sprintf("first time execution command: '%s'", opt.Command)
		}
		return checkers.Ok(msg)
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

func buildNoDifferenceMsg(filename string) (string, error) {
	file, err := openRD(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		return "", err
	}
	b := make([]byte, 128)
	count, err := file.Read(b)
	if err != nil {
		return "", err
	}
	o := string(strings.TrimRight(string(b[0:count]), "\r\n"))
	if fileinfo.Size() > 128 {
		return fmt.Sprintf("no difference: ```%s...```", o), nil
	}
	return fmt.Sprintf("no difference: ```%s```", o), nil
}

func buildDiffMsg(diff string) string {
	o := diff
	if len(diff) > 512 {
		o = diff[0:512]
	}
	o = strings.TrimRight(o, "\r\n")

	if len(diff) > 512 {
		return fmt.Sprintf("found difference: ```%s...```", o)
	}
	return fmt.Sprintf("found difference: ```%s```", o)
}

func getLines(filename string) ([]string, error) {
	file, err := openRD(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func diff(prev, new string) (string, error) {
	prevLines, err := getLines(prev)
	if err != nil {
		return "", err
	}
	newLines, err := getLines(new)
	if err != nil {
		return "", err
	}

	diff := gonp.New(prevLines, newLines)
	diff.Compose()

	return diff.SprintUniHunks(diff.UnifiedHunks()), nil
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
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if len(args) == 0 {
		psr.WriteHelp(os.Stderr)
		os.Exit(1)
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
