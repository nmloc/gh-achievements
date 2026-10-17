package render

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/nmloc/gh-achievements/internal/counters"
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
	Text(&buf, demoReport(t), Options{})
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
	Text(&plain, demoReport(t), Options{})
	Text(&color, demoReport(t), Options{Color: true})
	if strings.Contains(plain.String(), "\x1b[") {
		t.Error("plain output contains ANSI escapes")
	}
	if !strings.Contains(color.String(), "\x1b[") {
		t.Error("color output has no ANSI escapes")
	}
}

func TestText_noTrailingSpaces(t *testing.T) {
	for _, opt := range []Options{{}, {All: true}} {
		for i, line := range strings.Split(text(t, demoReport(t), opt), "\n") {
			if strings.HasSuffix(line, " ") {
				t.Errorf("%+v: line %d has trailing spaces: %q", opt, i+1, line)
			}
		}
	}
}

func text(t *testing.T, r *report.Report, opt Options) string {
	t.Helper()
	var buf bytes.Buffer
	Text(&buf, r, opt)
	return buf.String()
}

func TestText_readsLikeTheResultPage(t *testing.T) {
	out := text(t, demoReport(t), Options{})
	for _, want := range []string{
		"🏆 @torvalds", "3 of 8 badges you can still earn, 12 tiers",
		"Closest next unlock: Galaxy Brain, 0 / 2 answers to x1",
		"  On display\n", "  Still to collect\n",
		"72 / 128 to x3", "Top tier, 252,192 stars", "Since 2020",
		"Close an issue or pull request within 5 minutes of opening it",
		"The archive: 5 badges you can no longer earn (--all lists them)",
		"Highlights: none of 6 yet",
		"official = GitHub's own pages · calibrated = counted, checked against GitHub · counted = counted only",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q", want)
		}
	}
	if strings.Index(out, "On display") > strings.Index(out, "Still to collect") {
		t.Error("On display comes first")
	}
	for _, hidden := range []string{"Heart On Your Sleeve", "Proxima Pioneer", "subsurface/subsurface#1506"} {
		if strings.Contains(out, hidden) {
			t.Errorf("%q belongs to --all", hidden)
		}
	}
}

func TestText_rowsShowGlyphTierBarLineAndSource(t *testing.T) {
	out := text(t, demoReport(t), Options{})
	var shark, pair, galaxy string
	for _, l := range strings.Split(out, "\n") {
		switch {
		case strings.Contains(l, "Pull Shark"):
			shark = l
		case strings.Contains(l, "Pair Extraordinaire"):
			pair = l
		case strings.Contains(l, "Galaxy Brain") && strings.Contains(l, "counted"):
			galaxy = l
		}
	}
	if !strings.HasPrefix(shark, "  ● Pull Shark") || !strings.Contains(shark, " x2  ██████░░░░░░  72 / 128 to x3") || !strings.HasSuffix(shark, "calibrated") {
		t.Errorf("pull shark row = %q", shark)
	}
	if strings.Contains(pair, "█") || !strings.Contains(pair, "Top tier") || !strings.HasSuffix(pair, "official") {
		t.Errorf("pair row = %q (no counter, so no bar)", pair)
	}
	if !strings.HasPrefix(galaxy, "  ○ Galaxy Brain") || !strings.Contains(galaxy, "░░░░░░░░░░░░  0 / 2 answers to x1") {
		t.Errorf("galaxy row = %q", galaxy)
	}
	// The source labels line up in one column.
	col := func(line, word string) int { return utf8.RuneCountInString(line[:strings.Index(line, word)]) }
	if col(shark, "calibrated") != col(pair, "official") {
		t.Errorf("source columns differ:\n%s\n%s", shark, pair)
	}
}

func TestText_allListsTheArchiveHighlightsAndNotes(t *testing.T) {
	out := text(t, demoReport(t), Options{All: true})
	for _, want := range []string{"  The archive\n", "Heart On Your Sleeve", "Switched off by GitHub", "GitHub staff only",
		"↳ checked at GitHub's #16 subsurface/subsurface#1506", "  Highlights\n", "○ Pro"} {
		if !strings.Contains(out, want) {
			t.Errorf("--all output lacks %q", want)
		}
	}
	if strings.Contains(out, "(--all lists them)") {
		t.Error("--all must not tell you to use --all")
	}
}

func TestText_hiddenProfileSaysLiveCounts(t *testing.T) {
	r, err := report.Build("sindresorhus", fixtures.Official{}, &fixtures.Counts{Authed: true,
		PRs: counters.MergedPRs{Total: 2241}, FactsValue: &counters.Facts{}})
	if err != nil {
		t.Fatal(err)
	}
	out := text(t, r, Options{})
	if !strings.Contains(out, "  Live counts\n") || strings.Contains(out, "Still to collect") || strings.Contains(out, "On display") {
		t.Fatalf("hidden profile output:\n%s", out)
	}
	if !strings.Contains(out, "This profile hides its achievements: live counts only, no official tiers.") {
		t.Error("hidden profile lacks its notice")
	}
}

func TestText_nothingEarnedSaysWhereToStart(t *testing.T) {
	r := demoReport(t)
	rows := make([]report.Row, len(r.Achievements))
	for i, row := range r.Achievements {
		row.Status, row.Tier, row.Unlocked = report.NotEarned, 0, ""
		rows[i] = row
	}
	empty := *r
	empty.Achievements = rows
	if out := text(t, &empty, Options{}); !strings.Contains(out, "  On display\n  Nothing yet. Start with a badge from the list below.\n") {
		t.Fatalf("output:\n%s", out)
	}
}

func TestText_aheadOfGitHubSaysSo(t *testing.T) {
	r := demoReport(t)
	rows := append([]report.Row(nil), r.Achievements...)
	for i := range rows {
		if rows[i].Slug == "pull-shark" {
			st := *rows[i].Standing
			st.Tier, st.Count, st.Floor, st.Target, st.IsAhead = 3, 130, 128, 1024, true
			rows[i].Standing = &st
		}
	}
	ahead := *r
	ahead.Achievements = rows
	if out := text(t, &ahead, Options{}); !strings.Contains(out, "(count says x3)") {
		t.Fatalf("an ahead count must say so:\n%s", out)
	}
}

func TestStarLine(t *testing.T) {
	if got := StarLine(false); got != "  ★ Found it useful? Star it: https://github.com/nmloc/gh-achievements" {
		t.Fatalf("star line = %q", got)
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

// withRows returns a copy of the demo report with edit applied to a copy of its rows.
func withRows(t *testing.T, edit func(rows []report.Row) []report.Row) *report.Report {
	t.Helper()
	r := demoReport(t)
	copyOf := *r
	copyOf.Achievements = edit(append([]report.Row(nil), r.Achievements...))
	return &copyOf
}

func TestText_failureNotesShowWithoutAll(t *testing.T) {
	r := withRows(t, func(rows []report.Row) []report.Row {
		for i := range rows {
			if rows[i].Slug == "pull-shark" {
				rows[i].Standing, rows[i].Source = nil, ""
				rows[i].Notes = []string{"could not count: rate limited"}
			}
		}
		return rows
	})
	if out := text(t, r, Options{}); !strings.Contains(out, "↳ could not count: rate limited") {
		t.Fatalf("a failed count must show without --all:\n%s", out)
	}
	if out := text(t, demoReport(t), Options{}); strings.Contains(out, "↳") {
		t.Errorf("information notes belong to --all:\n%s", out)
	}
}

func TestText_aBadgeTheCatalogLacksShowsItsTierAndAsksForAnIssue(t *testing.T) {
	r := withRows(t, func(rows []report.Row) []report.Row {
		return append(rows, report.Row{Slug: "brand-new", Name: "Brand New", Kind: report.KindNew,
			Status: report.Earned, Tier: 3, Unlocked: "2026-10-01",
			HowToEarn: "Not in the catalog yet: please open an issue"})
	})
	out := text(t, r, Options{})
	onDisplay := out[strings.Index(out, "On display"):strings.Index(out, "Still to collect")]
	var line string
	for _, l := range strings.Split(onDisplay, "\n") {
		if strings.Contains(l, "Brand New") {
			line = l
		}
	}
	for _, want := range []string{"● Brand New", " x3 ", "Since 2026 · Not in the catalog yet: please open an issue"} {
		if !strings.Contains(line, want) {
			t.Errorf("the Brand New line in On display lacks %q: %q", want, line)
		}
	}
	for i, line := range strings.Split(out, "\n") {
		if strings.HasSuffix(line, " ") {
			t.Errorf("line %d has trailing spaces: %q", i+1, line)
		}
	}
}

func TestText_coloursEachTierInItsMetal(t *testing.T) {
	line := func(out, name string) string {
		for _, l := range strings.Split(out, "\n") {
			if strings.Contains(l, name) {
				return l
			}
		}
		t.Fatalf("no line for %s", name)
		return ""
	}
	out := text(t, demoReport(t), Options{Color: true})
	if l := line(out, "Pull Shark"); !strings.Contains(l, "\x1b[38;5;173m●") {
		t.Errorf("tier 2 is bronze: %q", l)
	}
	if l := line(out, "Starstruck"); !strings.Contains(l, "\x1b[38;5;178m●") {
		t.Errorf("tier 4 is gold: %q", l)
	}
	silver := withRows(t, func(rows []report.Row) []report.Row {
		for i := range rows {
			if rows[i].Slug == "pull-shark" {
				rows[i].Tier = 3
			}
		}
		return rows
	})
	if l := line(text(t, silver, Options{Color: true}), "Pull Shark"); !strings.Contains(l, "\x1b[38;5;250m●") {
		t.Errorf("tier 3 is silver: %q", l)
	}
}
