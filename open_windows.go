//go:build windows
// +build windows

package main

import "os"

func openRD(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_RDONLY, 0)
}
