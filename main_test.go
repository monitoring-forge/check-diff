package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mackerelio/checkers"
)

func getCmdOpt(t *testing.T, command string, args []string) (Opt, *os.File) {
	workdir := t.TempDir()
	filename := filepath.Join(workdir, "test.txt")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	return Opt{
		Command:    command,
		Args:       args,
		Identifier: "test",
		Workdir:    workdir,
	}, file
}

// Test cmd
func TestCmd(t *testing.T) {
	opt, file := getCmdOpt(t, "echo", []string{"Hello, World!"})
	defer file.Close()

	if err := opt.cmd(file); err != nil {
		t.Fatalf("cmd failed: %v", err)
	}

	file.Seek(0, 0) // Reset file pointer to the beginning
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		output := scanner.Text()
		expected := "Hello, World!"
		if output != expected {
			t.Errorf("Expected output %q, got %q", expected, output)
		}
	} else {
		t.Errorf("No output from command")
	}
}

func TestCmdFailedCase(t *testing.T) {
	opt, file := getCmdOpt(t, "nonexistent_command", []string{})
	defer file.Close()

	if err := opt.cmd(file); err == nil {
		t.Fatalf("Expected cmd to fail, but it succeeded")
	}
}

func TestCmdFailedCommand(t *testing.T) {
	opt, file := getCmdOpt(t, "ls", []string{"nonexistent_directory"})
	defer file.Close()
	if err := opt.cmd(file); err == nil {
		t.Fatalf("Expected cmd to fail, but it succeeded")
	}
}

func createTempFile(t *testing.T, content string) (*os.File, func()) {
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return tmpfile, func() {
		os.Remove(tmpfile.Name())
	}
}

func TestBuildNoDifferenceMsg(t *testing.T) {
	shortMsg := "This is a test message."
	file, cleanup := createTempFile(t, shortMsg+"\n")
	defer cleanup()
	expected := "no difference: ```" + shortMsg + "```"
	result, err := buildNoDifferenceMsg(file.Name())
	if err != nil {
		t.Fatalf("buildNoDifferenceMsg failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestBuildNoDifferenceMsgLargeResult(t *testing.T) {
	largeMsg := strings.Repeat("A", 600)
	file, cleanup := createTempFile(t, largeMsg+"\n")
	defer cleanup()

	expected := "no difference: ```" + largeMsg[:128] + "...```"
	result, err := buildNoDifferenceMsg(file.Name())
	if err != nil {
		t.Fatalf("buildNoDifferenceMsg failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestBuildDiffMsg(t *testing.T) {
	diff := "This is a test diff message.\nThis is the second line of the diff.\n"
	expected := "found difference: ```This is a test diff message.\nThis is the second line of the diff.```"
	result := buildDiffMsg(diff)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestBuildDiffMsgLargeResult(t *testing.T) {
	diff := strings.Repeat("A", 600)
	expected := "found difference: ```" + diff[:512] + "...```"
	result := buildDiffMsg(diff)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestDiff(t *testing.T) {
	file1, cleanup1 := createTempFile(t, "Hello, World!\n")
	defer cleanup1()

	file2, cleanup2 := createTempFile(t, "Hello, Go!\n")
	defer cleanup2()

	diffResult, err := diff(file1.Name(), file2.Name())
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if diffResult == "" {
		t.Errorf("Expected a diff result, got empty string")
	}
	if !strings.Contains(diffResult, "Hello, World!") || !strings.Contains(diffResult, "Hello, Go!") {
		t.Errorf("Diff result does not contain expected content: %q", diffResult)
	}
}

func TestDiffNoDifference(t *testing.T) {
	file1, cleanup1 := createTempFile(t, "Hello, World!\n")
	defer cleanup1()

	file2, cleanup2 := createTempFile(t, "Hello, World!\n")
	defer cleanup2()

	diffResult, err := diff(file1.Name(), file2.Name())
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if diffResult != "" {
		t.Errorf("Expected no diff result, got: %q", diffResult)
	}
}

func TestRun(t *testing.T) {
	opt, file := getCmdOpt(t, "echo", []string{"Hello, World!"})
	defer file.Close()
	ckr := opt.run()
	if ckr.Status != checkers.OK {
		t.Errorf("Expected OK status, got %v", ckr.Status)
	}
	if !strings.Contains(ckr.Message, "first time execution command") {
		t.Errorf("Expected first time execution message, got %q", ckr.Message)
	}

	// Run again to check for no difference
	ckr = opt.run()
	if ckr.Status != checkers.OK {
		t.Errorf("Expected OK status, got %v", ckr.Status)
	}
	if !strings.Contains(ckr.Message, "no difference") {
		t.Errorf("Expected no difference message, got %q", ckr.Message)
	}
}

func TestRunWithDifference(t *testing.T) {
	opt, file := getCmdOpt(t, "date", []string{})
	defer file.Close()
	// First run to create the initial state
	opt.run()

	// Change the command to produce a different output
	time.Sleep(2 * time.Second) // Ensure the date command produces a different output
	opt.Command = "date"
	ckr := opt.run()
	if ckr.Status != checkers.CRITICAL {
		t.Errorf("Expected CRITICAL status, got %v", ckr.Status)
	}
	if !strings.Contains(ckr.Message, "found difference") {
		t.Errorf("Expected found difference message, got %q", ckr.Message)
	}
}
