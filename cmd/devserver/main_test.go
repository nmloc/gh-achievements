package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	p := filepath.Join(t.TempDir(), ".env.local")
	_ = os.WriteFile(p, []byte("# comment\nA=1\nB = \"two words\"\n\nC='x=y'\n"), 0o600)
	env, err := loadEnv(p)
	if err != nil || env["A"] != "1" || env["B"] != "two words" || env["C"] != "x=y" || len(env) != 3 {
		t.Fatalf("loadEnv = %v, %v", env, err)
	}
}
