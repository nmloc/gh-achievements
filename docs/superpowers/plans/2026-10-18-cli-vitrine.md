# CLI and README in the Vitrine style: implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `gh achievements` prints the same shelves, wording and metals as the redesigned site, and the README looks like the site.

**Architecture:** A new `internal/wording` package owns every badge line and the headline numbers, used by both `web/view.go` and `internal/render`. `render.Text` is rewritten around shelves with an `Options{Color, All}` argument. `main.go` gains `--all` and a once-only star line. `cmd/gendocs/svg.go` restyles the generated SVGs; the README is rewritten around a screenshot captured by a committed Node script.

**Tech Stack:** Go standard library, go-gh (already a dependency), Node 22+ built-ins for the screenshot script, headless Google Chrome (local only).

**Spec:** `docs/superpowers/specs/2026-10-18-cli-vitrine-design.md`

## Global Constraints

- Wording is identical on the web and in the CLI and comes from one place (`internal/wording`).
- Metals: bronze x2, silver x3, gold x4. Terminal: ANSI 256 `38;5;173` bronze, `38;5;250` silver, `38;5;178` gold. SVG (site dark tokens): bronze `#d39468`, silver `#c7d0d6`, gold `#e2b53a`, ink `#eaf1f5`, muted `#93a9b7`, background `#0d1c26`, surface `#122633`. Light ink `#142430`.
- `--json` output unchanged. `--no-color` / `NO_COLOR` unchanged.
- Star line exactly `★ Found it useful? Star it: https://github.com/nmloc/gh-achievements`, at most once per machine (marker `<os.UserConfigDir()>/gh-achievements/star-shown`), only when stdout is a terminal and not `--json`; if the marker cannot be written, print nothing.
- No new Go or npm dependencies. CSP and site behaviour untouched except the regenerated README art.
- Commits: author nmloc <nmloc312@gmail.com>, Conventional Commits, no Co-Authored-By trailer, no AI attribution. Never push.
- Regenerate and commit generated files: `go run ./cmd/gendocs` (CI runs `-check`), `go test ./internal/render -run TestText_golden -update` and `go test ./web -run TestRenderResult_golden -update` when output changes. `gofmt -l .` empty. `node --test "internal/webui/jstest/*.test.mjs"` stays green.

---

### Task 1: One wording package for the web and the CLI

**Files:**
- Create: `internal/wording/wording.go`, `internal/wording/wording_test.go`
- Modify: `web/view.go` (use `wording`; delete the moved helpers), `web/view_test.go` (move the line table to `wording_test.go`)

**Interfaces:**
- Produces: `wording.Line(row report.Row) string`; `wording.Since(unlocked string) string`; `type Shelf int` with `Display`, `Collect`, `Archive`; `wording.ShelfOf(row report.Row) Shelf`; `type Summary struct{ Earned, EarnableTotal, TierTotal int; Closest *report.Row }`; `wording.Summarize(r *report.Report) Summary`.

- [ ] **Step 1: Write the failing tests** — `internal/wording/wording_test.go`:

Move `TestCard_lineShowsTheCountAgainstTheNextThreshold` from `web/view_test.go` here as `TestLine_countAgainstTheNextThreshold`, unchanged except `newCard(tc.row).Line` becomes `Line(tc.row)`. Add three more cases to its table: `{"switched off", report.Row{Slug: "x", Kind: catalog.Disabled, Tiers: []int{0, 16, 128, 0}, Status: report.NotEarned}, "Switched off by GitHub"}`, `{"retired, earned", report.Row{Slug: "x", Kind: catalog.Retired, Tiers: []int{1}, Status: report.Earned, Tier: 1, Unlocked: "Jan 2021"}, "Since 2021"}`, `{"hidden, no count", report.Row{Slug: "quickdraw", Kind: catalog.Earnable, Tiers: []int{1}, Status: report.Unknown}, "Hidden on this profile"}`. Then add:

```go
func demo(t *testing.T) *report.Report {
	t.Helper()
	r, err := report.Build("torvalds", fixtures.Official{}, fixtures.TorvaldsDemo())
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestSummarize_matchesTheResultPage(t *testing.T) {
	s := Summarize(demo(t))
	if s.Earned != 3 || s.EarnableTotal != 8 || s.TierTotal != 12 {
		t.Fatalf("summary = %+v", s)
	}
	if s.Closest == nil || s.Closest.Slug != "galaxy-brain" || Line(*s.Closest) != "0 / 2 answers to x1" {
		t.Fatalf("closest = %+v", s.Closest)
	}
}

func TestShelfOf(t *testing.T) {
	got := map[Shelf][]string{}
	for _, row := range demo(t).Achievements {
		got[ShelfOf(row)] = append(got[ShelfOf(row)], row.Slug)
	}
	for shelf, want := range map[Shelf]string{
		Display: "pull-shark,starstruck,pair-extraordinaire,arctic-code-vault-contributor,mars-2020-contributor",
		Collect: "galaxy-brain,quickdraw,yolo,public-sponsor,finishing-touch",
		Archive: "heart-on-your-sleeve,open-sourcerer,proxima-pioneer,proxima-staffshipper,proxima-staffuser",
	} {
		if strings.Join(got[shelf], ",") != want {
			t.Errorf("shelf %d = %v", shelf, got[shelf])
		}
	}
}
```

(Imports: `strings`, `testing`, `catalog`, `fixtures`, `progress`, `report`.)

- [ ] **Step 2: Run** `go test ./internal/wording` → FAIL (package does not exist / undefined: Line).

- [ ] **Step 3: Create `internal/wording/wording.go`**

```go
// Package wording words a report the same way on the web and in the terminal: one line per
// badge, the shelf it stands on and the headline numbers.
package wording

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/report"
)

// shortUnit names what a count counts, for badges whose name does not already say it.
var shortUnit = map[string]string{"starstruck": "stars", "galaxy-brain": "answers"}

// yearPattern finds the year in an unlock date such as "2020-02-02" or "Jan 2021".
var yearPattern = regexp.MustCompile(`\b(19|20)\d{2}\b`)

// Line is a badge's status in one line: "72 / 128 to x3", "Top tier, 252,192 stars",
// "x2, next tier at 24", "Since 2020", or how to earn a single-tier badge not earned yet.
func Line(row report.Row) string {
	earned := row.Status == report.Earned
	switch s := row.Standing; {
	case row.Kind != catalog.Earnable:
		return fixed(row.Kind, earned, row.Unlocked)
	case s != nil && s.IsMax:
		return topTier(row.Slug, s.Count)
	case s != nil:
		return count(row.Slug, s.Count, s.Target, s.Tier+1)
	case row.Status == report.Unknown:
		return "Hidden on this profile"
	case len(row.Tiers) > 1 && row.Tier < len(row.Tiers):
		return nextTier(row.Tier, row.Tiers[row.Tier])
	case len(row.Tiers) > 1:
		return "Top tier"
	case earned:
		return Since(row.Unlocked)
	}
	return row.HowToEarn
}

// Since is the year a badge was first unlocked, or "Earned" when GitHub gives no date.
func Since(unlocked string) string {
	if y := yearPattern.FindString(unlocked); y != "" {
		return "Since " + y
	}
	if unlocked != "" {
		return "Since " + unlocked
	}
	return "Earned"
}

// Shelf is where a badge stands: on display, still to collect, or in the archive.
type Shelf int

const (
	Display Shelf = iota // earned, any kind
	Collect              // earnable, not earned (or hidden)
	Archive              // cannot be earned any more, not earned
)

// ShelfOf places a row on its shelf.
func ShelfOf(row report.Row) Shelf {
	switch {
	case row.Status == report.Earned:
		return Display
	case row.Kind == catalog.Earnable:
		return Collect
	}
	return Archive
}

// Summary is the headline both the result page and the CLI print.
type Summary struct {
	Earned, EarnableTotal, TierTotal int
	Closest                          *report.Row // the counted earnable badge with the fewest to go; nil if none
}

// Summarize counts earned earnable badges, total tiers and the closest next unlock.
func Summarize(r *report.Report) Summary {
	var s Summary
	for i, row := range r.Achievements {
		if row.Status == report.Earned {
			s.TierTotal += row.Tier
		}
		if row.Kind != catalog.Earnable {
			continue
		}
		s.EarnableTotal++
		if row.Status == report.Earned {
			s.Earned++
		}
		if st := row.Standing; st != nil && !st.IsMax && st.Remaining > 0 &&
			(s.Closest == nil || st.Remaining < s.Closest.Standing.Remaining) {
			s.Closest = &r.Achievements[i]
		}
	}
	return s
}

func fixed(kind catalog.Kind, earned bool, unlocked string) string {
	switch {
	case earned:
		return Since(unlocked)
	case kind == catalog.Disabled:
		return "Switched off by GitHub"
	case kind == catalog.Retired:
		return "No longer earnable"
	case kind == catalog.Internal:
		return "GitHub staff only"
	}
	return "Not yet"
}

func count(slug string, n, target, nextTier int) string {
	nums := group(n) + " / " + group(target)
	if unit := shortUnit[slug]; unit != "" {
		nums += " " + unit
	}
	return fmt.Sprintf("%s to x%d", nums, nextTier)
}

func topTier(slug string, n int) string {
	if unit := shortUnit[slug]; unit != "" && n > 0 {
		return "Top tier, " + group(n) + " " + unit
	}
	return "Top tier"
}

func nextTier(tier, threshold int) string {
	switch {
	case tier == 0 && threshold > 0:
		return fmt.Sprintf("First tier at %d", threshold)
	case tier == 0:
		return "First tier: threshold unknown"
	case threshold > 0:
		return fmt.Sprintf("x%d, next tier at %d", tier, threshold)
	}
	return fmt.Sprintf("x%d, next tier unknown", tier)
}

// group writes n with thousands separators: 252192 → "252,192". Counts are never negative.
func group(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
```

- [ ] **Step 4: Use it in `web/view.go`.** Delete `shortUnit`, `yearPattern`, `fixedLine`, `countLine`, `topTierLine`, `nextTierLine`, `sinceLine`, `group` and the `regexp`/`strconv` imports. In `NewResultView`, replace the counting loop with:

```go
	sum := wording.Summarize(r)
	v.Earned, v.EarnableTotal, v.TierTotal = sum.Earned, sum.EarnableTotal, sum.TierTotal
	if sum.Closest != nil {
		c := newCard(*sum.Closest)
		v.Closest = &c
	}
	for _, row := range r.Achievements {
		c := newCard(row)
		switch wording.ShelfOf(row) {
		case wording.Display:
			v.Display = append(v.Display, c)
		case wording.Collect:
			v.Collect = append(v.Collect, c)
		default:
			v.Archive = append(v.Archive, c)
		}
	}
	return v
```

In `newCard`, replace the whole `switch s := row.Standing; { … }` with:

```go
	c.Line = wording.Line(row)
	if s := row.Standing; row.Kind == catalog.Earnable && s != nil {
		if s.IsMax {
			c.Progress = &Meter{Value: 1, Max: 1, Label: row.Name + ": top tier"}
		} else {
			span := max(s.Target-s.Floor, 1)
			c.Progress = &Meter{Value: min(max(s.Count-s.Floor, 0), span), Max: span,
				Label: fmt.Sprintf("%s: %d of %d", row.Name, s.Count, s.Target)}
			if s.IsAhead {
				c.Notes = append(slices.Clone(c.Notes), aheadNote(row.Tier, s.Tier))
			}
		}
	}
```

Remove `TestCard_lineShowsTheCountAgainstTheNextThreshold` from `web/view_test.go` (it moved). Every other web test stays as is.

- [ ] **Step 5: Verify** `go test ./internal/wording ./web` PASS; `go test ./web -run TestRenderResult_golden` PASS **without** `-update` (the page must not change); `go test ./... && gofmt -l . && go vet ./...` clean.

- [ ] **Step 6: Commit** `refactor: share badge wording between web and CLI`.

---

### Task 2: Terminal output on shelves, `--all`, and the star line

**Files:**
- Modify: `internal/render/render.go` (rewrite `Text`), `internal/render/render_test.go`, `internal/render/testdata/torvalds.txt` (golden), `main.go`, `main_test.go`, `cmd/gendocs/main.go` (the one `render.Text` call)
- Test: `internal/render/render_test.go`, `main_test.go`

**Interfaces:**
- Consumes: Task 1's `wording.Line`, `ShelfOf`, `Summarize`.
- Produces: `render.Options{Color, All bool}`; `render.Text(w io.Writer, r *report.Report, opt render.Options)`; `render.StarLine(color bool) string`; `render.StarURL`. The gendocs demo calls `render.Text(&colored, demo, render.Options{Color: true})`.

- [ ] **Step 1: Write the failing tests** in `internal/render/render_test.go` (keep `TestText_golden`, `TestText_noColorHasNoEscapes_colorHas`, `TestText_noTrailingSpaces`, `TestJSON_roundTrips`, changing their `Text(&buf, r, false)` calls to `Text(&buf, r, Options{})` and `true` to `Options{Color: true}`; also run `noTrailingSpaces` with `Options{All: true}`):

```go
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
```

(Add imports `counters` and `unicode/utf8`. If a fixture field name differs, follow `web/view_test.go`'s sindresorhus builder exactly.)

In `main_test.go` add:

```go
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
```

- [ ] **Step 2: Run** `go test ./internal/render . ` → FAIL (`Options` undefined, `starOnce` undefined).

- [ ] **Step 3: Rewrite `internal/render/render.go`** (keep `JSON` as is):

```go
// Package render prints a report as text or JSON.
package render

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/progress"
	"github.com/nmloc/gh-achievements/internal/report"
	"github.com/nmloc/gh-achievements/internal/wording"
)

// Options picks what Text prints.
type Options struct {
	Color bool // ANSI colour and bold
	All   bool // also the archive, every highlight and each badge's notes
}

// StarURL is where the one-time star request points.
const StarURL = "https://github.com/nmloc/gh-achievements"

const (
	barWidth = 12
	legend   = "official = GitHub's own pages · calibrated = counted, checked against GitHub · counted = counted only"
)

// JSON writes the report as indented JSON.
func JSON(w io.Writer, r *report.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

type style struct{ on bool }

func (s style) wrap(code, text string) string {
	if !s.on || text == "" {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}
func (s style) bold(t string) string { return s.wrap("1", t) }
func (s style) dim(t string) string  { return s.wrap("2", t) }
func (s style) gold(t string) string { return s.wrap("38;5;178", t) }

// metal colours a tier the way the site does: bronze x2, silver x3, gold x4.
func (s style) metal(tier int, text string) string {
	switch tier {
	case 2:
		return s.wrap("38;5;173", text)
	case 3:
		return s.wrap("38;5;250", text)
	case 4:
		return s.gold(text)
	}
	return text
}

// StarLine is the one-time request to star the repository.
func StarLine(color bool) string {
	return "  " + style{on: color}.gold("★") + " Found it useful? Star it: " + StarURL
}

// Text writes the human report: the headline, then the shelves the result page shows.
func Text(w io.Writer, r *report.Report, opt Options) {
	s := style{on: opt.Color}
	p := func(format string, a ...any) { fmt.Fprintf(w, format+"\n", a...) }

	shelves := map[wording.Shelf][]report.Row{}
	for _, row := range r.Achievements {
		shelves[wording.ShelfOf(row)] = append(shelves[wording.ShelfOf(row)], row)
	}
	display, collect, archive := shelves[wording.Display], shelves[wording.Collect], shelves[wording.Archive]
	shown := append(append([]report.Row{}, display...), collect...)
	if opt.All {
		shown = append(shown, archive...)
	}
	cols := measure(shown)
	rows := func(list []report.Row) {
		for _, row := range list {
			writeRow(w, s, cols, row, opt.All)
		}
	}

	sum := wording.Summarize(r)
	p("")
	p("  %s", s.bold("🏆 @"+r.User))
	p("  %d of %d badges you can still earn, %d tiers", sum.Earned, sum.EarnableTotal, sum.TierTotal)
	if c := sum.Closest; c != nil {
		p("  Closest next unlock: %s, %s", c.Name, wording.Line(*c))
	}
	if r.AchievementsHidden {
		p("  %s", s.dim("This profile hides its achievements: live counts only, no official tiers."))
	}
	if !r.Authenticated {
		p("  %s", s.dim("Not signed in to gh: public data only. Run `gh auth login` for full counts."))
	}

	switch {
	case len(display) > 0:
		p("")
		p("  %s", s.bold("On display"))
		rows(display)
	case !r.AchievementsHidden:
		p("")
		p("  %s", s.bold("On display"))
		p("  %s", s.dim("Nothing yet. Start with a badge from the list below."))
	}
	if len(collect) > 0 {
		title := "Still to collect"
		if r.AchievementsHidden {
			title = "Live counts"
		}
		p("")
		p("  %s", s.bold(title))
		rows(collect)
	}
	if len(archive) > 0 {
		p("")
		if opt.All {
			p("  %s", s.bold("The archive"))
			rows(archive)
		} else {
			p("  %s %s", s.bold(fmt.Sprintf("The archive: %d %s you can no longer earn", len(archive), plural(len(archive), "badge"))),
				s.dim("(--all lists them)"))
		}
	}
	writeHighlights(w, s, r.Highlights, opt.All)
	p("")
	p("  %s", s.dim(legend))
	p("")
}

type columns struct{ name, line int }

// measure finds the widest name and the widest line that is followed by a source label,
// so names, bars and source labels each line up in one column.
func measure(rows []report.Row) columns {
	var c columns
	for _, row := range rows {
		c.name = max(c.name, width(row.Name))
		if hasSource(row) {
			c.line = max(c.line, width(wording.Line(row)))
		}
	}
	return c
}

// hasSource reports whether a row ends with its source label: earnable tiered badges, whose
// count or tier a reader may want to trace.
func hasSource(row report.Row) bool {
	return row.Kind == catalog.Earnable && row.Source != "" && len(row.Tiers) > 1
}

func writeRow(w io.Writer, s style, c columns, row report.Row, all bool) {
	earned := row.Status == report.Earned
	glyph, tier := s.dim("○"), ""
	name := s.dim(pad(row.Name, c.name))
	if earned {
		glyph, tier = s.metal(row.Tier, "●"), "✓"
		if len(row.Tiers) > 1 {
			tier = fmt.Sprintf("x%d", row.Tier)
		}
		name = s.bold(pad(row.Name, c.name))
	}
	b := "  " + glyph + " " + name + "  " + s.metal(row.Tier, pad(tier, 2)) + "  "
	if row.Standing != nil {
		b += bar(s, *row.Standing) + "  "
	} else {
		b += strings.Repeat(" ", barWidth+2)
	}
	line := wording.Line(row)
	if hasSource(row) {
		b += pad(line, c.line) + "  " + s.dim(string(row.Source))
	} else {
		b += line
	}
	if st := row.Standing; st != nil && st.IsAhead {
		b += s.dim(fmt.Sprintf("  (count says x%d)", st.Tier))
	}
	fmt.Fprintln(w, strings.TrimRight(b, " "))
	if all {
		for _, note := range row.Notes {
			fmt.Fprintf(w, "      %s\n", s.dim("↳ "+note))
		}
	}
}

// bar fills from the current tier's threshold to the next one; a top tier is full.
func bar(s style, st progress.Standing) string {
	filled := barWidth
	if !st.IsMax {
		filled = 0
		if span := st.Target - st.Floor; span > 0 {
			filled = (barWidth*(st.Count-st.Floor)*2 + span) / (span * 2) // round half up
		}
		filled = max(0, min(barWidth, filled))
	}
	return s.gold(strings.Repeat("█", filled)) + s.dim(strings.Repeat("░", barWidth-filled))
}

func writeHighlights(w io.Writer, s style, hs []report.HighlightRow, all bool) {
	fmt.Fprintln(w)
	if all {
		fmt.Fprintf(w, "  %s\n", s.bold("Highlights"))
		for _, h := range hs {
			if h.Earned {
				fmt.Fprintf(w, "  %s %s\n", s.gold("●"), h.Name)
			} else {
				fmt.Fprintf(w, "  %s\n", s.dim("○ "+h.Name))
			}
		}
		return
	}
	var earned []string
	for _, h := range hs {
		if h.Earned {
			earned = append(earned, h.Name)
		}
	}
	text := fmt.Sprintf("none of %d yet", len(hs))
	if len(earned) > 0 {
		text = strings.Join(earned, ", ")
	}
	fmt.Fprintf(w, "  %s %s\n", s.bold("Highlights:"), text)
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

func width(s string) int { return utf8.RuneCountInString(s) }

func pad(s string, n int) string { return s + strings.Repeat(" ", max(n-width(s), 0)) }
```

- [ ] **Step 4: `main.go`.** Add `"bytes"`-free imports `io`, `path/filepath`. Add the flag `all := fs.Bool("all", false, "")`, document it in `usage` as `  --all        also list the archive, every highlight and each badge's notes`. Replace the text output with:

```go
	t := term.FromEnv()
	color := !*noColor && os.Getenv("NO_COLOR") == "" && t.IsTerminalOutput() && t.IsColorEnabled()
	render.Text(os.Stdout, r, render.Options{Color: color, All: *all})
	if t.IsTerminalOutput() {
		starOnce(os.Stdout, color, starMarker())
	}
	return nil
```

and add:

```go
// starOnce prints the request to star the repository the first time a report reaches a
// terminal, and leaves a marker so it never repeats. If the marker cannot be written it
// prints nothing: asking on every run would be worse than not asking.
func starOnce(w io.Writer, color bool, marker string) {
	if marker == "" {
		return
	}
	if _, err := os.Stat(marker); err == nil {
		return // already asked
	}
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		return
	}
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		return
	}
	fmt.Fprintln(w, render.StarLine(color))
	fmt.Fprintln(w)
}

// starMarker is where starOnce remembers that it asked; "" if there is no config directory.
func starMarker() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "gh-achievements", "star-shown")
}
```

In `cmd/gendocs/main.go` change `render.Text(&colored, demo, true)` to `render.Text(&colored, demo, render.Options{Color: true})`. (`cmd/gendocs/svg.go` will still compile but its section filter keys on the old headings; Task 3 replaces it. `go run ./cmd/gendocs` in this task regenerates `docs/demo.svg`/`site/demo.svg`-free outputs with whatever the old filter keeps; that is fine, Task 3 regenerates again.)

- [ ] **Step 5: Verify.** `go test ./internal/render -run TestText_golden -update`, then read `internal/render/testdata/torvalds.txt` and check it matches the spec's example layout (shelves, aligned columns, archive line, highlights line, legend). `go run ./cmd/gendocs`; `go run ./cmd/gendocs -check && go test ./... && go vet ./... && gofmt -l .` clean. Also run `go run . torvalds --no-color` once (live, read-only GitHub calls) and paste its output into the report.

- [ ] **Step 6: Commit** `feat(cli): print the report on shelves like the site` (render, golden, gendocs call, regenerated files) and `feat(cli): --all and a one-time star request` (main.go, main_test.go) — or one commit if they cannot build apart.

---

### Task 3: Generated SVGs in the site's palette

**Files:**
- Modify: `cmd/gendocs/svg.go` (`ansiFill`, `keepDemoLines`, `terminalSVG`, `socialSVG`, `checkButtonSVG`), `cmd/gendocs/readme_test.go`
- Regenerate: `docs/demo.svg`, `docs/social.svg`, `docs/check-button.svg`

**Interfaces:**
- Consumes: Task 2's colour codes (`2`, `1`, `38;5;173`, `38;5;250`, `38;5;178`) and headings (`On display`, `Still to collect`, `The archive…`, `Highlights…`, legend line).

- [ ] **Step 1: Failing tests** in `cmd/gendocs/readme_test.go`:

```go
func TestTerminalSVG_usesTheSitePaletteAndTheNewLayout(t *testing.T) {
	demo, err := report.Build("torvalds", fixtures.Official{}, fixtures.TorvaldsDemo())
	if err != nil {
		t.Fatal(err)
	}
	var colored bytes.Buffer
	render.Text(&colored, demo, render.Options{Color: true})
	svg := string(terminalSVG(colored.String(), ""))
	for _, want := range []string{`fill="#0d1c26"`, `fill="#e2b53a"`, `fill="#d39468"`, `fill="#93a9b7"`,
		"On display", "Still to collect", "72 / 128 to x3", "The archive"} {
		if !strings.Contains(svg, want) {
			t.Errorf("demo SVG lacks %q", want)
		}
	}
	for _, gone := range []string{"#0d1117", "#3fb950", "official = GitHub"} {
		if strings.Contains(svg, gone) {
			t.Errorf("demo SVG still has %q", gone)
		}
	}
}

func TestSVGs_matchTheSite(t *testing.T) {
	for name, svg := range map[string]string{"social": string(socialSVG()), "button": string(checkButtonSVG())} {
		for _, want := range []string{"#e2b53a"} {
			if !strings.Contains(svg, want) {
				t.Errorf("%s SVG lacks the site gold %s", name, want)
			}
		}
		for _, gone := range []string{"#2ea043", "#1a7f37", "#0d1117", "#fae57e"} {
			if strings.Contains(svg, gone) {
				t.Errorf("%s SVG still uses the old colour %s", name, gone)
			}
		}
	}
	if !strings.Contains(string(socialSVG()), "72\u00a0/\u00a0128\u00a0to\u00a0x3") { // the SVG keeps columns with no-break spaces
		t.Error("the social card should show the new progress wording")
	}
}
```

(Imports `bytes`, `fixtures`, `render`, `report`. `TestCheckButtonSVG_isWellFormedAndSaysWhatItDoes` stays and must keep passing.)

- [ ] **Step 2: Run** `go test ./cmd/gendocs -run 'TestTerminalSVG|TestSVGs'` → FAIL.

- [ ] **Step 3: Implement.**
  - `ansiFill` becomes `{"2": "#93a9b7", "38;5;173": "#d39468", "38;5;250": "#c7d0d6", "38;5;178": "#e2b53a"}`; bold (`1`) stays bold.
  - `keepDemoLines`: keep every line of the output except the legend line (`official = …`) and trailing blank lines; drop `demoSections`, `sectionWords` and the "… plus disabled and internal badges" footer (the new output already folds the archive into one line).
  - `terminalSVG`: background `#0d1c26`, default text `#eaf1f5`, title text `#93a9b7`; keep the three window dots.
  - `socialSVG`: background `#0d1c26`, top stripe `#e2b53a`, headline `#eaf1f5`, sub-lines `#93a9b7`, card `#122633` with stroke `#25414f`, install line `#e2b53a`. Rows become the new wording with metal marks: `{"●", bronze "#d39468", "Pull Shark", "x2  72 / 128 to x3  calibrated"}`, `{"●", gold, "Starstruck", "x4  Top tier, 252,192 stars  counted"}`, `{"●", gold, "Pair Extraord.", "x4  Top tier  official"}`, `{"●", ink "#eaf1f5", "Arctic Code Vault", "✓   Since 2020"}`. Keep the 1280×640 size and the existing layout constants.
  - `checkButtonSVG`: replace the green gradient with ink `#142430` (rect fill) and stroke `#e2b53a` (2px); trophy fill/stroke `#e2b53a`; title text `#eaf1f5`; sub-line `#c7d0d6`. Keep the size, texts, `role`/`<title>` and no external references.
- [ ] **Step 4: Verify.** `go run ./cmd/gendocs`; open `docs/demo.svg` text to check no line was cut; `go run ./cmd/gendocs -check && go test ./... && gofmt -l .` clean. Note in the report that `docs/social.png` must be re-rendered with `scripts/social-png.sh` (needs `rsvg-convert`; run it if installed, otherwise say so).
- [ ] **Step 5: Commit** `feat(docs): demo, social card and button in the site palette`.

---

### Task 4: README achievement table with a medal per tier

Loc asked (2026-10-18) to learn from drknzz/GitHub-Achievements: each badge row there holds a small table, one column per tier (Default, Bronze, Silver, Gold) with that tier's medal above the amount it needs. Ours prints `2 · 16 · 128 · 1024` in one cell.

**Files:**
- Modify: `cmd/gendocs/main.go` (`readmeTable`), `cmd/gendocs/site.go` (share the group headings with `listGroups`), `cmd/gendocs/readme_test.go`
- Regenerate: `README.md` (between `<!-- catalog:start -->` and `<!-- catalog:end -->`)

**Interfaces:**
- Produces: `kindGroups` — `[]struct{ Kind catalog.Kind; Heading, Note string }` in catalog order of kinds, used by both `listGroups` (site) and `readmeTable` (README), so the README and `/achievements/` group badges identically.

- [ ] **Step 1: Failing test.** Replace `TestReadmeTable_hasTheBadgeArt` with:

```go
func TestReadmeTable_groupsBadgesAndShowsEveryTierMedal(t *testing.T) {
	table := readmeTable()
	last := -1
	for _, h := range []string{"### Earnable", "### Switched off by GitHub", "### Retired", "### GitHub staff only"} {
		i := strings.Index(table, h)
		if i < 0 {
			t.Errorf("%s missing", h)
			continue
		}
		if i < last {
			t.Errorf("%s out of order", h)
		}
		last = i
	}
	row := func(slug string) string {
		for _, line := range strings.Split(table, "\n") {
			if strings.Contains(line, "achievements/"+slug+"/)") {
				return line
			}
		}
		t.Fatalf("no row for %s", slug)
		return ""
	}
	ps := row("pull-shark")
	for i, want := range []string{"2", "16", "128", "1024"} {
		if !strings.Contains(ps, `<img src="`+catalog.Image("pull-shark", i+1)+`" width="48" height="48" alt="Pull Shark `+metals[i]+`">`) {
			t.Errorf("pull shark lacks its %s medal", metals[i])
		}
		if !strings.Contains(ps, `<td align="center">`+want+`</td>`) {
			t.Errorf("pull shark lacks the amount %s", want)
		}
	}
	if !strings.Contains(ps, "<th>Default</th><th>Bronze</th><th>Silver</th><th>Gold</th>") {
		t.Errorf("pull shark tier headings: %s", ps)
	}
	if heart := row("heart-on-your-sleeve"); strings.Count(heart, `<td align="center">?</td>`) != 2 {
		t.Errorf("heart on your sleeve shows unknown amounts as ?: %s", heart)
	}
	if yolo := row("yolo"); !strings.Contains(yolo, "<th>Default</th></tr>") || strings.Contains(yolo, "Bronze") ||
		!strings.Contains(yolo, `<td align="center">1</td>`) {
		t.Errorf("a single-tier badge has one medal and the amount 1: %s", yolo)
	}
	for _, a := range catalog.Achievements {
		if line := row(a.Slug); strings.Count(line, "|") != 5 {
			t.Errorf("%s row must stay one markdown row of 4 cells: %s", a.Slug, line)
		}
	}
	if !strings.Contains(table, "Profile highlights it also reports:") || !strings.Contains(table, "_Catalog last verified against live profiles on ") {
		t.Error("highlights and verified lines lost")
	}
}
```

Run `go test ./cmd/gendocs -run TestReadmeTable_` → FAIL.

- [ ] **Step 2: Share the groups (`cmd/gendocs/site.go`).** Add above `listGroups`:

```go
// kindGroups names the catalog's sections, in display order, for the site's list and the README.
var kindGroups = []struct {
	Kind          catalog.Kind
	Heading, Note string
}{
	{catalog.Earnable, "Earnable", "You can still earn these."},
	{catalog.Disabled, "Switched off by GitHub", "GitHub tested these and switched them off; nobody can earn them now."},
	{catalog.Retired, "Retired", "No longer awarded. Profiles that have them keep them."},
	{catalog.Internal, "GitHub staff only", "Only GitHub staff could earn these."},
}
```

and rewrite `listGroups` to build its `[]badgeGroup` from `kindGroups` (same headings and notes as today, so the site does not change).

- [ ] **Step 3: Rewrite `readmeTable` (`cmd/gendocs/main.go`):**

```go
// readmeTable groups the catalog like the site's list and, as GitHub's own badge dialog does,
// shows every tier as its medal above the amount it needs.
func readmeTable() string {
	var b strings.Builder
	for _, g := range kindGroups {
		fmt.Fprintf(&b, "### %s\n\n%s\n\n", g.Heading, g.Note)
		b.WriteString("| Badge | Achievement | How to earn it | Tiers |\n|:-:|---|---|:-:|\n")
		for _, a := range catalog.Achievements {
			if a.Kind != g.Kind {
				continue
			}
			art := ""
			if img := catalog.Image(a.Slug, len(a.Tiers)); img != "" {
				art = fmt.Sprintf(`<img src="%s" width="40" height="40" alt="%s badge">`, img, a.Name)
			}
			fmt.Fprintf(&b, "| %s | [%s](%sachievements/%s/) | %s | %s |\n",
				art, a.Name, webui.SiteURL, a.Slug, a.HowToEarn, tierTable(a))
		}
		b.WriteString("\n")
	}
	b.WriteString("Profile highlights it also reports: ")
	names := make([]string, len(catalog.Highlights))
	for i, h := range catalog.Highlights {
		names[i] = h.Name
	}
	b.WriteString(strings.Join(names, ", ") + ".\n\n")
	fmt.Fprintf(&b, "_Catalog last verified against live profiles on %s._\n", catalog.LastVerified)
	return b.String()
}

// tierTable is one inline HTML table: a column per tier, its medal above the amount needed.
// It stays on one line so the markdown row around it does not break.
func tierTable(a catalog.Achievement) string {
	n := max(len(a.Tiers), 1)
	var head, art, need strings.Builder
	for i := 0; i < n; i++ {
		amount := "1"
		if i < len(a.Tiers) {
			amount = "?"
			if a.Tiers[i] > 0 {
				amount = fmt.Sprint(a.Tiers[i])
			}
		}
		head.WriteString("<th>" + metalTitles[i] + "</th>")
		img := catalog.Image(a.Slug, i+1)
		art.WriteString(fmt.Sprintf(`<td align="center"><img src="%s" width="48" height="48" alt="%s %s"></td>`, img, a.Name, metals[i]))
		need.WriteString(`<td align="center">` + amount + "</td>")
	}
	return "<table><tr>" + head.String() + "</tr><tr>" + art.String() + "</tr><tr>" + need.String() + "</tr></table>"
}

// metalTitles head the tier columns, as on GitHub's own badge art.
var metalTitles = []string{"Default", "Bronze", "Silver", "Gold"}
```

(`metals` is the site's `[]string{"base", "bronze", "silver", "gold"}` in `site.go`; the image alt uses it, e.g. `alt="Pull Shark bronze"`. A single-tier badge has one column: Default, its medal, amount `1`. If a badge ever had more than 4 tiers this would index past `metalTitles`; the catalog has at most 4 — add `n = min(n, len(metalTitles))`.)

- [ ] **Step 4: Verify.** `go run ./cmd/gendocs`, then read the regenerated README section: four `###` groups, one row per badge, each row one line. `go run ./cmd/gendocs -check && go test ./... && gofmt -l .` clean. `TestReadmeBadges_showEveryEarnableBadgeAtItsTopTier` and the site list tests still pass (the site list must not change: `git diff --stat site/achievements/index.html` empty).

- [ ] **Step 5: Commit** `docs(readme): a medal per tier in the achievements table`.

---

### Task 5: README and its screenshots

**Files:**
- Create: `scripts/readme-screenshots.mjs`, `docs/screenshot-light.png`, `docs/screenshot-dark.png`
- Modify: `README.md` (hand-written parts only; the `catalog` and `badges` marker sections stay generated), `CONTRIBUTING.md` (one line on retaking screenshots)
- Test: `cmd/gendocs/readme_test.go`

- [ ] **Step 1: Failing test** in `cmd/gendocs/readme_test.go`:

```go
func TestReadme_showsTheSiteAndBothWaysToCheck(t *testing.T) {
	b, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	readme := string(b)
	for _, want := range []string{
		`<source media="(prefers-color-scheme: dark)" srcset="docs/screenshot-dark.png">`,
		`<img src="docs/screenshot-light.png"`,
		"## Two ways to check", "## What it shows", "gh achievements --all",
		"(docs/demo.svg)", "docs/check-button.svg",
		"https://github.com/nmloc/gh-achievements/stargazers",
	} {
		if !strings.Contains(readme, want) {
			t.Errorf("README lacks %q", want)
		}
	}
	for _, png := range []string{"../../docs/screenshot-light.png", "../../docs/screenshot-dark.png"} {
		if fi, err := os.Stat(png); err != nil || fi.Size() < 20_000 {
			t.Errorf("%s missing or too small", png)
		}
	}
}
```

(Import `os`.) Run `go test ./cmd/gendocs -run TestReadme_` → FAIL.

- [ ] **Step 2: Write `scripts/readme-screenshots.mjs`** — a dependency-free Chrome DevTools Protocol client (Node 22+: global `WebSocket` and `fetch`):

```js
#!/usr/bin/env node
// Capture the README screenshots of the result page, light and dark.
// Needs Google Chrome and the dev server: `go run ./cmd/devserver` (http://localhost:8787).
//   node scripts/readme-screenshots.mjs [url]
import { spawn } from "node:child_process";
import { mkdtempSync, writeFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const url = process.argv[2] || "http://localhost:8787/result?u=torvalds";
const chrome = process.env.CHROME || "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const port = 9333;
const profile = mkdtempSync(join(tmpdir(), "gha-shot-"));
const proc = spawn(chrome, ["--headless=new", `--remote-debugging-port=${port}`, `--user-data-dir=${profile}`,
  "--hide-scrollbars", "--no-first-run", "about:blank"], { stdio: "ignore" });

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
async function target() {
  for (let i = 0; i < 50; i++) {
    try {
      const list = await (await fetch(`http://127.0.0.1:${port}/json/list`)).json();
      const page = list.find((t) => t.type === "page");
      if (page) return page.webSocketDebuggerUrl;
    } catch { /* Chrome still starting */ }
    await sleep(200);
  }
  throw new Error("Chrome did not start");
}

const ws = new WebSocket(await target());
await new Promise((r) => ws.addEventListener("open", r, { once: true }));
let seq = 0;
const pending = new Map();
ws.addEventListener("message", (e) => {
  const msg = JSON.parse(e.data);
  if (msg.id && pending.has(msg.id)) {
    const { resolve, reject } = pending.get(msg.id);
    pending.delete(msg.id);
    msg.error ? reject(new Error(msg.error.message)) : resolve(msg.result);
  }
});
const send = (method, params = {}) => new Promise((resolve, reject) => {
  const id = ++seq;
  pending.set(id, { resolve, reject });
  ws.send(JSON.stringify({ id, method, params }));
});
const evaluate = (expression) => send("Runtime.evaluate", { expression, awaitPromise: true, returnByValue: true });

const width = 1280, height = 1100;
await send("Page.enable");
await send("Emulation.setDeviceMetricsOverride", { width, height, deviceScaleFactor: 1.5, mobile: false });
await send("Page.navigate", { url: new URL("/", url).href });
await sleep(1500);
for (const theme of ["light", "dark"]) {
  // No consent banner, no analytics, and the theme under test.
  await evaluate(`localStorage.setItem("gha:consent", "denied"); localStorage.setItem("gha:theme", "${theme}"); sessionStorage.removeItem("gha:stars"); true`);
  await send("Page.navigate", { url });
  await sleep(1500);
  await evaluate(`Promise.all([document.fonts.ready, ...[...document.images].map((i) => i.complete ? 0 : new Promise((r) => { i.onload = i.onerror = r; }))]).then(() => true)`);
  await sleep(300);
  const { data } = await send("Page.captureScreenshot", { format: "png", clip: { x: 0, y: 0, width, height, scale: 1 } });
  const out = `docs/screenshot-${theme}.png`;
  writeFileSync(out, Buffer.from(data, "base64"));
  console.log("wrote", out);
}
ws.close();
proc.kill();
rmSync(profile, { recursive: true, force: true });
```

- [ ] **Step 3: Capture.** Start the dev server in the background (`go run ./cmd/devserver`; it reads `.env.local` itself — never print or open that file), wait for `curl -s -o /dev/null -w '%{http_code}' http://localhost:8787/` to return 200, run `node scripts/readme-screenshots.mjs`, then stop the server. Look at both PNGs (Read them as images): header, @torvalds heading, "On display" shelf and "Still to collect" visible, no consent banner, the right theme in each. Each file under 1.5 MB; if one is larger, lower `deviceScaleFactor` to 1.25 and capture again.

- [ ] **Step 4: Rewrite the README's hand-written parts.** Keep the H1, the `badges` and `catalog` marker blocks, the shields line, and these sections unchanged: Install, the generated "All GitHub achievements" block, How accurate is it?, FAQ, How it compares, JSON output, Contributing, Credits, License. Replace everything between the `<!-- badges:end -->` line and `## Install` with:

```markdown
<p align="center"><strong>Your GitHub trophy case: every achievement, your tier and how far the next one is, verified against GitHub's own data.</strong></p>

<p align="center">
  <a href="https://gh-achievements.vercel.app/check/"><img src="docs/check-button.svg" width="520" alt="🏆 Check your GitHub achievements → no install; sign in with GitHub to include private work"></a>
</p>

<p align="center">
  <a href="https://gh-achievements.vercel.app/check/"><strong>gh-achievements.vercel.app/check</strong></a> · no install · or use the CLI below
</p>

<p align="center">
  <a href="https://gh-achievements.vercel.app/result?u=torvalds">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="docs/screenshot-dark.png">
      <img src="docs/screenshot-light.png" width="880" alt="The gh-achievements result page for @torvalds: Pull Shark x2 at 72 / 128 to x3, Starstruck and Pair Extraordinaire at their top tier, and the retired Arctic Code Vault and Mars 2020 badges on a lit shelf; Galaxy Brain, Quickdraw, YOLO, Public Sponsor and Finishing Touch still to collect">
    </picture>
  </a>
</p>

[![Release](https://img.shields.io/github/v/release/nmloc/gh-achievements)](https://github.com/nmloc/gh-achievements/releases)
[![Tests](https://github.com/nmloc/gh-achievements/actions/workflows/test.yml/badge.svg)](https://github.com/nmloc/gh-achievements/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Two ways to check

| On the web | In your terminal |
|---|---|
| **[gh-achievements.vercel.app/check](https://gh-achievements.vercel.app/check/)** | `gh extension install nmloc/gh-achievements` then `gh achievements` |
| Type any username, or sign in with GitHub to include private repositories. The token is used for one check and revoked at once; nothing is stored. | Reuses your `gh auth login`, so it counts everything that login can see, including organizations that restrict third-party apps. |

![Terminal output of gh achievements torvalds: on display Pull Shark x2 at 72 / 128 to x3 (calibrated), Starstruck at the top tier with 252,192 stars, Pair Extraordinaire at the top tier, the retired Arctic Code Vault and Mars 2020 badges; still to collect Galaxy Brain at 0 / 2 answers, Quickdraw, YOLO, Public Sponsor and Finishing Touch](docs/demo.svg)

## What it shows

- **Every badge, not just yours.** All 15 known achievements (8 earnable, 2 switched off by GitHub, 2 retired, 3 for GitHub staff) and 6 profile highlights such as Pro.
- **How far to the next tier.** Your count against the next threshold, such as `72 / 128 to x3` for Pull Shark.
- **Checked against GitHub.** Tiers, unlock dates and milestones come from GitHub's own badge history, and every number is labelled `official`, `calibrated` or `counted`.
```

In **Usage**, add `gh achievements --all        # also the archive, every highlight and each badge's notes` after the `--json` line and keep the alignment of the `#` comments. Delete the old `## Web` section (its content is now in "Two ways to check") and the old intro paragraph + bullet list that followed the demo image (replaced by "What it shows"). Before `## Credits`, add:

```markdown
## Like it?

If gh-achievements helped you, a ⭐ helps other people find it: [star it on GitHub](https://github.com/nmloc/gh-achievements/stargazers).
```

and in **FAQ** keep every question. In `CONTRIBUTING.md`, next to the site/test instructions, add one line: "README screenshots: `go run ./cmd/devserver` in one terminal, then `node scripts/readme-screenshots.mjs` (needs Google Chrome)."

- [ ] **Step 5: Verify.** `go run ./cmd/gendocs` (the marker sections regenerate; nothing else should change), `go run ./cmd/gendocs -check && go test ./... && gofmt -l . && node --test "internal/webui/jstest/*.test.mjs"` clean. Render the README locally in your head against GitHub's rules: no HTML GitHub strips (`<picture>`, `<source>`, `<img>`, `<a>`, `<p align>` are all allowed).
- [ ] **Step 6: Commit** `docs: README in the site's style, with screenshots` (README, CONTRIBUTING, PNGs, script).

---

### Task 6: Check it (controller)

- [ ] `go run . torvalds` in a real terminal tab (colour) and `--no-color`, `--all`, `--json | head`; a second run shows no star line; delete the marker afterwards only if it was created by this check (`<UserConfigDir>/gh-achievements/star-shown`, ask before deleting anything else).
- [ ] View the README on the branch on GitHub after push (light and dark), and `docs/demo.svg` as an image.
