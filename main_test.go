package main

import "testing"

func TestReorder_flagsAfterTheUsername(t *testing.T) {
	got := reorder([]string{"torvalds", "--json", "--no-color"})
	if len(got) != 3 || got[0] != "--json" || got[1] != "--no-color" || got[2] != "torvalds" {
		t.Fatalf("reorder = %v", got)
	}
}
