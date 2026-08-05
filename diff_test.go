package main

import (
	"strings"
	"testing"
)

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
