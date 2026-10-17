// Package progress turns official tiers and live counts into one honest number.
//
// The rule: a count may raise the shown tier (GitHub can lag) but a tier GitHub granted
// is never taken back, so a count below the tier's floor means part of the activity is
// invisible to the API, and the floor is shown instead.
package progress

import (
	"fmt"

	"github.com/nmloc/gh-achievements/internal/official"
)

// Source says how far a number can be trusted.
type Source string

const (
	Official   Source = "official"   // read from GitHub's own pages
	Calibrated Source = "calibrated" // counted, then checked against an official milestone
	Counted    Source = "counted"    // counted, nothing public to check it against
)

// TierFromCount is the highest tier whose threshold count reaches. An unknown
// threshold (0) is skipped, so a count past a known higher one still reaches that tier.
func TierFromCount(tiers []int, count int) int {
	tier := 0
	for i, t := range tiers {
		if t > 0 && count >= t {
			tier = i + 1
		}
	}
	return tier
}

// Standing is where a count sits between two tiers.
type Standing struct {
	Tier      int  `json:"tier"`             // effective tier: max(official, implied by count)
	Count     int  `json:"count"`            // never below the floor of Tier
	Floor     int  `json:"floor"`            // threshold of Tier, 0 below x1
	Target    int  `json:"target,omitempty"` // threshold of the next tier
	Remaining int  `json:"remaining,omitempty"`
	IsMax     bool `json:"isMax"`
	// IsAhead is true when the count implies a higher tier than GitHub shows.
	IsAhead bool `json:"isAhead"`
}

// Stand places count against the tiers, given the tier GitHub shows (0 if none or hidden).
func Stand(tiers []int, officialTier, count int) Standing {
	officialTier = min(max(officialTier, 0), len(tiers)) // a tier GitHub adds later must not panic
	implied := TierFromCount(tiers, count)
	tier := max(officialTier, implied)
	s := Standing{Tier: tier, IsAhead: implied > officialTier}
	if tier > 0 {
		s.Floor = tiers[tier-1]
	}
	s.Count = max(count, s.Floor)
	if tier >= len(tiers) {
		s.IsMax = true
		return s
	}
	s.Target = tiers[tier]
	s.Remaining = s.Target - s.Count
	return s
}

// Calibration compares GitHub's count with ours at a milestone both can see.
type Calibration struct {
	At     int    `json:"at"`     // GitHub's ordinal, e.g. 16
	PR     string `json:"pr"`     // the pull request GitHub names for it
	Offset int    `json:"offset"` // how many more GitHub counts than we can see
}

// Calibrate uses the highest milestone whose pull request appears in refs (oldest first).
// If GitHub says PR X was its Nth and X is the i-th in our list, GitHub counts N - i
// pull requests we cannot see.
func Calibrate(refs []string, milestones []official.Milestone) *Calibration {
	position := make(map[string]int, len(refs))
	for i, r := range refs {
		position[r] = i + 1
	}
	var best *Calibration
	for _, m := range milestones {
		if m.N == 0 || m.PR == "" {
			continue
		}
		if i, ok := position[m.PR]; ok && (best == nil || m.N > best.At) {
			best = &Calibration{At: m.N, PR: m.PR, Offset: m.N - i}
		}
	}
	return best
}

// Ordinal formats 1 → 1st, 2 → 2nd, 11 → 11th, 23 → 23rd.
func Ordinal(n int) string {
	suffix := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", n, suffix)
}
