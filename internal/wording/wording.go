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

// fixed is the status of a badge that can no longer be earned: no progress, no threshold.
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

// count is "count / next threshold to xN": where the count stands against the next tier.
func count(slug string, n, target, nextTier int) string {
	nums := group(n) + " / " + group(target)
	if unit := shortUnit[slug]; unit != "" {
		nums += " " + unit
	}
	return fmt.Sprintf("%s to x%d", nums, nextTier)
}

// topTier adds the count where it means something on its own ("252,192 stars").
func topTier(slug string, n int) string {
	if unit := shortUnit[slug]; unit != "" && n > 0 {
		return "Top tier, " + group(n) + " " + unit
	}
	return "Top tier"
}

// nextTier is for a tiered badge with no counter: the tier GitHub shows, then the next threshold.
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
