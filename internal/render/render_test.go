package render

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nmloc/gh-achievements/internal/fixtures"
	"github.com/nmloc/gh-achievements/internal/report"
)

var update = flag.Bool("update", false, "rewrite golden files")

func demoReport(t *testing.T) *report.Report {
	t.Helper()
	r, err := report.Build("torvalds", fixtures.Official{}, fixtures.TorvaldsDemo())
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestText_golden(t *testing.T) {
	var buf bytes.Buffer
	Text(&buf, demoReport(t), false)
	golden := filepath.Join("testdata", "torvalds.txt")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run with -update)", err)
	}
	if buf.String() != string(want) {
		t.Fatalf("output differs from %s (run with -update if intended):\n%s", golden, buf.String())
	}
}

func TestText_noColorHasNoEscapes_colorHas(t *testing.T) {
	var plain, color bytes.Buffer
	Text(&plain, demoReport(t), false)
	Text(&color, demoReport(t), true)
	if strings.Contains(plain.String(), "\x1b[") {
		t.Error("plain output contains ANSI escapes")
	}
	if !strings.Contains(color.String(), "\x1b[") {
		t.Error("color output has no ANSI escapes")
	}
}

func TestText_noTrailingSpaces(t *testing.T) {
	var buf bytes.Buffer
	Text(&buf, demoReport(t), false)
	for i, line := range strings.Split(buf.String(), "\n") {
		if strings.HasSuffix(line, " ") {
			t.Errorf("line %d has trailing spaces: %q", i+1, line)
		}
	}
}

func TestJSON_roundTrips(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, demoReport(t)); err != nil {
		t.Fatal(err)
	}
	var back report.Report
	if err := json.Unmarshal(buf.Bytes(), &back); err != nil {
		t.Fatal(err)
	}
	if back.User != "torvalds" || len(back.Achievements) != 15 || len(back.Highlights) != 6 {
		t.Fatalf("decoded %+v", back)
	}
}
