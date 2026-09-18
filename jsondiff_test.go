package main

import (
	"strings"
	"testing"

	"github.com/itchyny/gojq"
	"github.com/stretchr/testify/require"
)

func TestJsonDiff(t *testing.T) {
	file1, cleanup1 := createTempFile(t, `{"key": "value", "another_key": ["item1", "item2"]}`)
	defer cleanup1()

	file2, cleanup2 := createTempFile(t, `{"key": "value", "another_key": ["item3", "item4"]}`)
	defer cleanup2()

	diff, err := diffJson(file1.Name(), file2.Name(), nil)
	if err != nil {
		t.Fatalf("diffjson failed: %v", err)
	}
	splited := strings.Split(diff, "\n")
	require.Contains(t, splited[0], "@@ ")
	require.Contains(t, diff, `-    "item1",`)
	require.Contains(t, diff, `-    "item2"`)
	require.Contains(t, diff, `+    "item3",`)
	require.Contains(t, diff, `+    "item4"`)
	require.NotContains(t, diff, `new`)
	require.NotContains(t, diff, `prev`)
}

func TestJsonDiffIgnore(t *testing.T) {
	file1, cleanup1 := createTempFile(t, `{"key": "value", "another_key": ["item1", "item2"]}`)
	defer cleanup1()

	file2, cleanup2 := createTempFile(t, `{"key": "value", "another_key": ["item3", "item4"]}`)
	defer cleanup2()

	jq, err := gojq.Parse(".another_key")
	if err != nil {
		t.Fatalf("gojq.Parse failed: %v", err)
	}

	diff, err := diffJson(file1.Name(), file2.Name(), jq)
	if err != nil {
		t.Fatalf("diffjson failed: %v", err)
	}
	require.Equal(t, "", diff)
}
