//go:build !windows
// +build !windows

package main

import (
	"os"
	"syscall"
)

func openRD(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
}
