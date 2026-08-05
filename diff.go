package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/cubicdaiya/gonp"
)

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
	if err != nil && err != io.EOF {
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
