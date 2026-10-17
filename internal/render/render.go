// Package render prints a report as text or JSON.
package render

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/report"
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
func (s style) bold(t string) string  { return s.wrap("1", t) }
func (s style) dim(t string) string   { return s.wrap("2", t) }
func (s style) green(t string) string { return s.wrap("32", t) }

// tier marks use GitHub's own label colours: bronze, silver, gold.
func (s style) mark(tier int, text string) string {
	switch tier {
	case 1:
		return s.green(text)
	case 2:
		return s.wrap("38;5;216", text)
	case 3:
		return s.wrap("38;5;252", text)
	case 4:
		return s.wrap("38;5;221", text)
	}
	return s.dim(text)
}

// Text writes the human report. color adds ANSI styling.
func Text(w io.Writer, r *report.Report, color bool) {
	s := style{on: color}
	p := func(format string, a ...interface{}) { fmt.Fprintf(w, format+"\n", a...) }

	p("")
	p("  %s", s.bold("GitHub achievements · @"+r.User))
	if r.AchievementsHidden {
		p("  This profile hides achievements: no official tiers, counters only.")
	}
	if !r.Authenticated {
		p("  %s", s.dim("Not signed in to gh: public data only. Run `gh auth login` for full counts."))
	}

	kinds := append([]catalog.Kind{}, catalog.Kinds...)
	kinds = append(kinds, report.KindNew)
	for _, kind := range kinds {
		rows := rowsOf(r, kind)
		if len(rows) == 0 {
			continue
		}
		title := catalog.KindTitle(kind)
		if kind == report.KindNew {
			title = "NEW · on this profile, not in the catalog yet"
		}
		p("")
		p("  %s", s.bold(title))
		for _, row := range rows {
			writeRow(w, s, row)
		}
	}

	p("")
	p("  %s", s.bold("HIGHLIGHTS"))
	for _, h := range r.Highlights {
		if h.Earned {
			p("  %s   %s", s.green("✓"), h.Name)
		} else {
			p("  %s   %s", s.dim("·"), s.dim(h.Name))
		}
	}
	p("")
	p("  %s", s.dim("official = GitHub's own pages · calibrated = counted, checked against GitHub · counted = counted only"))
	p("")
}

func rowsOf(r *report.Report, kind catalog.Kind) []report.Row {
	var rows []report.Row
	for _, row := range r.Achievements {
		if row.Kind == kind {
			rows = append(rows, row)
		}
	}
	return rows
}

func writeRow(w io.Writer, s style, row report.Row) {
	p := func(format string, a ...interface{}) { fmt.Fprintf(w, format+"\n", a...) }

	mark := "--"
	switch row.Status {
	case report.Earned:
		mark = fmt.Sprintf("x%d", row.Tier)
	case report.Unknown:
		mark = "··"
	}
	when := ""
	if row.Unlocked != "" {
		when = "since " + row.Unlocked
	}
	name := fmt.Sprintf("%-30s", row.Name)
	if row.Status == report.Earned {
		name = s.bold(name)
	}
	head := fmt.Sprintf("  %s %s %s", s.mark(row.Tier, fmt.Sprintf("%-3s", mark)), name, s.dim(when))
	p("%s", strings.TrimRight(head, " "))

	how := row.HowToEarn
	if len(row.Tiers) > 1 {
		how += " · tiers " + tierList(row.Tiers)
	}
	p("      %s", s.dim(how))

	if row.Kind != catalog.Earnable {
		return
	}
	switch {
	case row.Standing != nil:
		p("      %s", standing(s, row))
	case len(row.Tiers) > 1 && row.Counter == catalog.CounterNone && row.Status != report.Unknown:
		if row.Tier >= len(row.Tiers) {
			p("      max tier  [%s]", row.Source)
		} else {
			p("      next x%d at %d  [%s]", row.Tier+1, row.Tiers[row.Tier], row.Source)
		}
	case len(row.Tiers) == 1 && row.Status == report.Earned:
		p("      %s  [%s]", s.green("done"), row.Source)
	case len(row.Tiers) == 1 && row.Status == report.NotEarned:
		p("      0 / 1  [%s]", row.Source)
	}
	for _, note := range row.Notes {
		p("      · %s", note)
	}
}

func tierList(tiers []int) string {
	parts := make([]string, len(tiers))
	for i, t := range tiers {
		if t == 0 {
			parts[i] = "?"
		} else {
			parts[i] = fmt.Sprint(t)
		}
	}
	return strings.Join(parts, " · ")
}

func standing(s style, row report.Row) string {
	st := row.Standing
	ahead := ""
	if st.IsAhead {
		ahead = fmt.Sprintf("  (count says x%d)", st.Tier)
	}
	if st.IsMax {
		return fmt.Sprintf("%6d  max tier  [%s]%s", st.Count, row.Source, ahead)
	}
	return fmt.Sprintf("%6d / %-5d %s %d to x%d  [%s]%s",
		st.Count, st.Target, bar(s, st.Count, st.Floor, st.Target), st.Remaining, st.Tier+1, row.Source, ahead)
}

func bar(s style, value, lo, hi int) string {
	const width = 16
	filled := 0
	if hi > lo {
		filled = (width*(value-lo)*2 + (hi - lo)) / ((hi - lo) * 2) // round half up
	}
	filled = max(0, min(width, filled))
	return s.green(strings.Repeat("█", filled)) + s.dim(strings.Repeat("░", width-filled))
}
