package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cubicdaiya/gonp"
	"github.com/monitoring-forge/saferio"
)

func getLines(filename string) ([]string, error) {
	file, err := saferio.OpenRD(filename)
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

func diff(prev, current string) (string, error) {
	prevLines, err := getLines(prev)
	if err != nil {
		return "", err
	}
	newLines, err := getLines(current)
	if err != nil {
		return "", err
	}

	d := gonp.New(prevLines, newLines)
	d.Compose()

	return d.SprintUniHunks(d.UnifiedHunks()), nil
}

func buildNoDifferenceMsg(filename string) (string, error) {
	file, err := saferio.OpenRD(filename)
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
	if err != nil && !errors.Is(err, io.EOF) {
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
