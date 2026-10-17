package main

import "testing"

func TestIsLogin(t *testing.T) {
	// Regression: "torvalds?x" became /torvalds?x?tab=achievements and printed a wrong
	// report for torvalds under another name.
	valid := []string{"torvalds", "nmloc", "a", "Schweinepriester", "dependabot-preview", "a-b-c", "x1234567890123456789012345678901234567"}
	invalid := []string{"", "torvalds?x", "../etc", "a b", "-lead", "trail-", "dou--ble", "a/b", "x12345678901234567890123456789012345678901"}
	for _, s := range valid {
		if !isLogin(s) {
			t.Errorf("isLogin(%q) = false", s)
		}
	}
	for _, s := range invalid {
		if isLogin(s) {
			t.Errorf("isLogin(%q) = true", s)
		}
	}
}

func TestReorder_flagsAfterTheUsername(t *testing.T) {
	got := reorder([]string{"torvalds", "--json", "--no-color"})
	if len(got) != 3 || got[0] != "--json" || got[1] != "--no-color" || got[2] != "torvalds" {
		t.Fatalf("reorder = %v", got)
	}
}
