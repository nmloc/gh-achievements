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
