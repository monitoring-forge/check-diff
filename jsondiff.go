package main

import (
	"strings"

	"github.com/aereal/jsondiff"
	"github.com/itchyny/gojq"
	"github.com/monitoring-forge/saferio"
)

func diffJsonInput(path, name string) (*jsondiff.Input, error) {
	file, err := saferio.OpenRD(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	input, err := jsondiff.NewInputFromFile(file)
	if err != nil {
		return nil, err
	}
	input.Name = name
	return input, nil
}

func diffJson(prevFilePath, newFilePath string, jq *gojq.Query) (string, error) {
	prevDiff, err := diffJsonInput(prevFilePath, "prev")
	if err != nil {
		return "", err
	}
	newDiff, err := diffJsonInput(newFilePath, "new")
	if err != nil {
		return "", err
	}

	options := []jsondiff.Option{}
	if jq != nil {
		options = append(options, jsondiff.Ignore(jq))
	}
	diff, err := jsondiff.Diff(prevDiff, newDiff, options...)
	if err != nil {
		return "", err
	}
	if diff == "" {
		return "", nil
	}
	split := strings.Split(strings.TrimSuffix(diff, "\n"), "\n")
	if len(split) < 3 || !strings.HasPrefix(split[0], "--- ") || !strings.HasPrefix(split[1], "+++ ") {
		return diff, nil
	}
	return strings.Join(split[2:], "\n"), nil
}
