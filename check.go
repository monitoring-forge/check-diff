package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/mackerelio/checkers"
	"github.com/monitoring-forge/saferio"
)

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

func (opt *Opt) check() *checkers.Checker {

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

	prevFileName := fmt.Sprintf("check-diff-%s-%x", curUser.Uid, hasher.Sum(nil))
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

	if !saferio.FileExists(opt.Workdir, prevFileName) {
		err = os.Rename(newFile.Name(), filepath.Join(opt.Workdir, prevFileName))
		if err != nil {
			return checkers.Critical(err.Error())
		}
		if len(opt.Args) > 0 {
			return checkers.Ok(fmt.Sprintf("first time execution command: '%s %s'", opt.Command, strings.Join(opt.Args, " ")))
		}
		return checkers.Ok(fmt.Sprintf("first time execution command: '%s'", opt.Command))
	}

	diff, err := diff(filepath.Join(opt.Workdir, prevFileName), newFile.Name())
	if err != nil {
		return checkers.Critical(err.Error())
	}

	err = os.Rename(newFile.Name(), filepath.Join(opt.Workdir, prevFileName))
	if err != nil {
		return checkers.Critical(err.Error())
	}

	if diff == "" {
		msg, err := buildNoDifferenceMsg(filepath.Join(opt.Workdir, prevFileName))
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
