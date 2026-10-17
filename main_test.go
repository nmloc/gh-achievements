package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReorder_flagsAfterTheUsername(t *testing.T) {
	got := reorder([]string{"torvalds", "--json", "--no-color"})
	if len(got) != 3 || got[0] != "--json" || got[1] != "--no-color" || got[2] != "torvalds" {
		t.Fatalf("reorder = %v", got)
	}
}

func TestStarOnce_printsOnceAndRemembers(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "gh-achievements", "star-shown")
	var first, second bytes.Buffer
	starOnce(&first, false, marker)
	starOnce(&second, false, marker)
	if !strings.Contains(first.String(), "Star it: https://github.com/nmloc/gh-achievements") || second.Len() != 0 {
		t.Fatalf("first %q, second %q", first.String(), second.String())
	}
}

func TestStarOnce_staysQuietWhenItCannotRemember(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "file")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	starOnce(&out, false, filepath.Join(blocker, "star-shown")) // parent is a file: cannot be created
	starOnce(&out, false, "")
	if out.Len() != 0 {
		t.Fatalf("printed %q although it could not remember", out.String())
	}
}
