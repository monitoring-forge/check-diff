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

// Test cmd
func TestCmd(t *testing.T) {
	workdir := t.TempDir()
	opt := Opt{
		Command:    "echo",
		Args:       []string{"Hello, World!"},
		Identifier: "test",
		Workdir:    workdir,
	}
	filename := filepath.Join(workdir, "test.txt")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = opt.cmd(file)
	if err != nil {
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
	workdir := t.TempDir()
	opt := Opt{
		Command:    "nonexistent_command",
		Args:       []string{},
		Identifier: "test",
		Workdir:    workdir,
	}
	filename := filepath.Join(workdir, "test.txt")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = opt.cmd(file)
	if err == nil {
		t.Fatalf("Expected cmd to fail, but it succeeded")
	}
}

func TestCmdFailedCommand(t *testing.T) {
	workdir := t.TempDir()
	opt := Opt{
		Command:    "ls",
		Args:       []string{"nonexistent_directory"},
		Identifier: "test",
		Workdir:    workdir,
	}
	filename := filepath.Join(workdir, "test.txt")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = opt.cmd(file)
	if err == nil {
		t.Fatalf("Expected cmd to fail, but it succeeded")
	}
}

func TestBuildNoDifferenceMsg(t *testing.T) {
	file, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()
	shortMsg := "This is a test message."
	file.WriteString(shortMsg + "\n")
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
	file, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()
	largeMsg := make([]byte, 600)
	for i := range largeMsg {
		largeMsg[i] = 'A'
	}
	file.Write(largeMsg)
	expected := "no difference: ```" + string(largeMsg)[:128] + "...```"
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
	diff := make([]byte, 600)
	for i := range diff {
		diff[i] = 'A'
	}
	expected := "found difference: ```" + string(diff)[:512] + "...```"
	result := buildDiffMsg(string(diff))
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestDiff(t *testing.T) {
	file1, err := os.CreateTemp("", "file1")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file1.Name())
	defer file1.Close()
	file1.WriteString("Hello, World!\n")

	file2, err := os.CreateTemp("", "file2")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file2.Name())
	defer file2.Close()
	file2.WriteString("Hello, Go!\n")

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
	file1, err := os.CreateTemp("", "file1")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file1.Name())
	defer file1.Close()
	file1.WriteString("Hello, World!\n")

	file2, err := os.CreateTemp("", "file2")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file2.Name())
	defer file2.Close()
	file2.WriteString("Hello, World!\n")

	diffResult, err := diff(file1.Name(), file2.Name())
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if diffResult != "" {
		t.Errorf("Expected no diff result, got: %q", diffResult)
	}
}

func TestRun(t *testing.T) {
	workdir := t.TempDir()
	opt := Opt{
		Command:    "echo",
		Args:       []string{"Hello, World!"},
		Identifier: "test",
		Workdir:    workdir,
	}
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
	workdir := t.TempDir()
	opt := Opt{
		Command:    "date",
		Args:       []string{},
		Identifier: "test",
		Workdir:    workdir,
	}
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
