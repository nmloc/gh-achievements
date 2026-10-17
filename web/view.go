package web

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/official"
	"github.com/nmloc/gh-achievements/internal/progress"
	"github.com/nmloc/gh-achievements/internal/report"
	"github.com/nmloc/gh-achievements/internal/webui"
)

// Meter feeds a <progress> element: progress from the current tier's floor to the next.
type Meter struct {
	Value, Max int
	Label      string
}

// Link is a milestone row; URL is empty for a private one.
type Link struct{ Label, URL string }

// Card is one badge on the results page.
type Card struct {
	Slug, Name, HowToEarn, Image, ImageAlt, Initials, TierLabel, Line string
	Source, SourceMeaning, BadgeURL, Unlocked, TierText               string
	Tier                                                              int
	Earned, Locked                                                    bool
	Progress                                                          *Meter
	Notes                                                             []string
	Milestones                                                        []Link
}

// ResultView is everything result.html shows.
type ResultView struct {
	User, Avatar                     string
	Private, Hidden, Authenticated   bool
	Earned, EarnableTotal, TierTotal int
	Closest                          *Card
	QuickWins, Earnable, Other       []Card
	Highlights                       []report.HighlightRow
	Verified                         string
}

var tierNames = []string{"", "base", "bronze", "silver", "gold"}

var sourceMeaning = map[progress.Source]string{
	progress.Official:   "Read from GitHub's own profile pages.",
	progress.Calibrated: "Counted through the GitHub API, then checked against a milestone GitHub names.",
	progress.Counted:    "Counted through the GitHub API; nothing public to check it against.",
}

// shortUnit names what a count counts, for badges whose name does not already say it.
var shortUnit = map[string]string{"starstruck": "stars", "galaxy-brain": "answers"}

// yearPattern finds the year in an unlock date such as "2020-02-02" or "Jan 2021".
var yearPattern = regexp.MustCompile(`\b(19|20)\d{2}\b`)

// NewResultView turns a report into the page model. private marks a signed-in result.
func NewResultView(r *report.Report, private bool) ResultView {
	v := ResultView{
		User: r.User, Avatar: "https://github.com/" + r.User + ".png?size=120",
		Private: private, Hidden: r.AchievementsHidden, Authenticated: r.Authenticated,
		Highlights: r.Highlights, Verified: r.CatalogVerified,
	}
	bestRemaining := 0
	for _, row := range r.Achievements {
		c := newCard(row)
		if row.Status == report.Earned {
			v.TierTotal += row.Tier
		}
		if row.Kind != catalog.Earnable {
			v.Other = append(v.Other, c)
			continue
		}
		v.EarnableTotal++
		if c.Earned {
			v.Earned++
		}
		v.Earnable = append(v.Earnable, c)
		if st := row.Standing; st != nil && !st.IsMax && st.Remaining > 0 &&
			(v.Closest == nil || st.Remaining < bestRemaining) {
			cc := c
			v.Closest = &cc
			bestRemaining = st.Remaining
		}
		if len(row.Tiers) == 1 && row.Status == report.NotEarned {
			v.QuickWins = append(v.QuickWins, c)
		}
	}
	return v
}

func newCard(row report.Row) Card {
	c := Card{
		Slug: row.Slug, Name: row.Name, HowToEarn: row.HowToEarn, Tier: row.Tier,
		Earned: row.Status == report.Earned, Locked: row.Status != report.Earned,
		Source: string(row.Source), SourceMeaning: sourceMeaning[row.Source],
		Unlocked: row.Unlocked, Notes: webNotes(row.Notes), TierText: tierText(row.Tiers),
		Image: catalog.Image(row.Slug, row.Tier), Initials: webui.Initials(row.Name),
	}
	if _, known := catalog.BySlug(row.Slug); known {
		c.BadgeURL = webui.Rel(webui.SiteURL + "achievements/" + row.Slug + "/")
	}
	c.ImageAlt = row.Name + " badge, not earned"
	if row.Status == report.Unknown {
		c.ImageAlt = row.Name + " badge, status hidden"
	}
	if c.Earned {
		c.TierLabel = "✓"
		c.ImageAlt = row.Name + " badge"
		if len(row.Tiers) > 1 {
			c.TierLabel = fmt.Sprintf("x%d", row.Tier)
			c.ImageAlt = fmt.Sprintf("%s, %s tier (x%d)", row.Name, tierNames[min(row.Tier, 4)], row.Tier)
		}
	}
	switch s := row.Standing; {
	case row.Kind != catalog.Earnable:
		c.Line = fixedLine(row.Kind, c.Earned, row.Unlocked)
	case s != nil && s.IsMax:
		c.Line = topTierLine(row.Slug, s.Count)
		c.Progress = &Meter{Value: 1, Max: 1, Label: row.Name + ": top tier"}
	case s != nil:
		c.Line = countLine(row.Slug, s.Count, s.Target, s.Tier+1)
		span := max(s.Target-s.Floor, 1)
		c.Progress = &Meter{Value: min(max(s.Count-s.Floor, 0), span), Max: span,
			Label: fmt.Sprintf("%s: %d of %d", row.Name, s.Count, s.Target)}
		if s.IsAhead {
			c.Notes = append(slices.Clone(c.Notes), aheadNote(row.Tier, s.Tier))
		}
	case row.Status == report.Unknown:
		c.Line = "Hidden on this profile"
	case len(row.Tiers) > 1 && row.Tier < len(row.Tiers):
		c.Line = nextTierLine(row.Tier, row.Tiers[row.Tier])
	case len(row.Tiers) > 1:
		c.Line = "Top tier"
	case c.Earned:
		c.Line = sinceLine(row.Unlocked)
	default:
		c.Line = row.HowToEarn
	}
	if row.Official != nil {
		for _, m := range row.Official.Milestones {
			c.Milestones = append(c.Milestones, milestoneLink(m))
		}
	}
	return c
}

// webNotes words the CLI's notes for the web: counts there come from the site's bot token
// (public) or the visitor's sign-in (private), and "gh auth login" cannot apply. It returns
// a new slice and leaves the report's notes alone.
func webNotes(notes []string) []string {
	var out []string
	for _, n := range notes {
		if strings.Contains(n, "`gh auth login`") {
			continue
		}
		out = append(out, strings.ReplaceAll(n, "what your gh login can see", "what this check can see"))
	}
	return out
}

// fixedLine is the status of a badge that can no longer be earned: no progress, no threshold.
func fixedLine(kind catalog.Kind, earned bool, unlocked string) string {
	switch {
	case earned:
		return sinceLine(unlocked)
	case kind == catalog.Disabled:
		return "Switched off by GitHub"
	case kind == catalog.Retired:
		return "No longer earnable"
	case kind == catalog.Internal:
		return "GitHub staff only"
	}
	return "Not yet"
}

// countLine is "count / next threshold to xN": where the count stands against the next tier.
func countLine(slug string, count, target, nextTier int) string {
	nums := group(count) + " / " + group(target)
	if unit := shortUnit[slug]; unit != "" {
		nums += " " + unit
	}
	return fmt.Sprintf("%s to x%d", nums, nextTier)
}

// topTierLine adds the count where it means something on its own ("252,192 stars").
func topTierLine(slug string, count int) string {
	if unit := shortUnit[slug]; unit != "" && count > 0 {
		return "Top tier, " + group(count) + " " + unit
	}
	return "Top tier"
}

// nextTierLine is for a tiered badge with no counter: the tier GitHub shows, then the next threshold.
func nextTierLine(tier, threshold int) string {
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

// sinceLine is the year a badge was first unlocked, or "Earned" when GitHub gives no date.
func sinceLine(unlocked string) string {
	if y := yearPattern.FindString(unlocked); y != "" {
		return "Since " + y
	}
	if unlocked != "" {
		return "Since " + unlocked
	}
	return "Earned"
}

// group writes n with thousands separators: 252192 → "252,192". Counts are never negative.
func group(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func aheadNote(officialTier, countTier int) string {
	shown := "no tier"
	if officialTier > 0 {
		shown = fmt.Sprintf("x%d", officialTier)
	}
	return fmt.Sprintf("GitHub shows %s; the count says x%d", shown, countTier)
}

// milestoneLink emits a URL only for a well-formed "owner/repo" (and "#number"); anything else
// stays label-only rather than becoming a broken link.
func milestoneLink(m official.Milestone) Link {
	switch {
	case m.PR != "":
		repo, num, _ := strings.Cut(m.PR, "#")
		if isRepo(repo) && isDigits(num) {
			return Link{Label: m.Label, URL: "https://github.com/" + repo + "/pull/" + num}
		}
	case isRepo(m.Repo):
		return Link{Label: m.Label, URL: "https://github.com/" + m.Repo}
	}
	return Link{Label: m.Label}
}

func isRepo(s string) bool {
	owner, name, ok := strings.Cut(s, "/")
	return ok && owner != "" && name != "" && !strings.Contains(name, "/")
}

func isDigits(s string) bool {
	return s != "" && strings.Trim(s, "0123456789") == ""
}

func tierText(tiers []int) string {
	if len(tiers) < 2 {
		return ""
	}
	parts := make([]string, len(tiers))
	for i, t := range tiers {
		parts[i] = "?"
		if t > 0 {
			parts[i] = fmt.Sprint(t)
		}
	}
	return strings.Join(parts, " · ")
}
