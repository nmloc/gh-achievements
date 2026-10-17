package web

import (
	"fmt"
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

// NewResultView turns a report into the page model. private marks a signed-in result.
func NewResultView(r *report.Report, private bool) ResultView {
	v := ResultView{
		User: r.User, Avatar: "https://github.com/" + r.User + ".png?size=120",
		Private: private, Hidden: r.AchievementsHidden, Authenticated: r.Authenticated,
		Highlights: r.Highlights, Verified: r.CatalogVerified,
	}
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
		if row.Standing != nil && !row.Standing.IsMax && row.Standing.Remaining > 0 &&
			(v.Closest == nil || row.Standing.Remaining < remaining(r, v.Closest.Slug)) {
			cc := c
			v.Closest = &cc
		}
		if len(row.Tiers) == 1 && row.Status == report.NotEarned {
			v.QuickWins = append(v.QuickWins, c)
		}
	}
	return v
}

func remaining(r *report.Report, slug string) int {
	for _, row := range r.Achievements {
		if row.Slug == slug && row.Standing != nil {
			return row.Standing.Remaining
		}
	}
	return 1 << 30
}

func newCard(row report.Row) Card {
	c := Card{
		Slug: row.Slug, Name: row.Name, HowToEarn: row.HowToEarn, Tier: row.Tier,
		Earned: row.Status == report.Earned, Locked: row.Status != report.Earned,
		Source: string(row.Source), SourceMeaning: sourceMeaning[row.Source],
		Unlocked: row.Unlocked, Notes: row.Notes, TierText: tierText(row.Tiers),
		Image: catalog.Image(row.Slug, row.Tier), Initials: webui.Initials(row.Name),
	}
	if _, known := catalog.BySlug(row.Slug); known {
		c.BadgeURL = webui.Rel(webui.SiteURL + "achievements/" + row.Slug + "/")
	}
	c.ImageAlt = row.Name + " badge, not earned"
	if c.Earned {
		c.TierLabel = "✓"
		c.ImageAlt = row.Name + " badge"
		if len(row.Tiers) > 1 {
			c.TierLabel = fmt.Sprintf("x%d", row.Tier)
			c.ImageAlt = fmt.Sprintf("%s, %s tier (x%d)", row.Name, tierNames[min(row.Tier, 4)], row.Tier)
		}
	}
	switch s := row.Standing; {
	case row.Status == report.Unknown:
		c.Line = "hidden on this profile"
	case s != nil && s.IsMax:
		c.Line = "max tier"
		c.Progress = &Meter{Value: 1, Max: 1, Label: row.Name + ": max tier"}
	case s != nil:
		c.Line = fmt.Sprintf("%d to x%d", s.Remaining, s.Tier+1)
		c.Progress = &Meter{Value: s.Count - s.Floor, Max: s.Target - s.Floor,
			Label: fmt.Sprintf("%s: %d of %d", row.Name, s.Count, s.Target)}
	case len(row.Tiers) > 1 && row.Tier < len(row.Tiers):
		c.Line = fmt.Sprintf("next x%d at %d", row.Tier+1, row.Tiers[row.Tier])
	case len(row.Tiers) > 1:
		c.Line = "max tier"
	case c.Earned && row.Unlocked != "":
		c.Line = "since " + row.Unlocked
	case c.Earned:
		c.Line = "earned"
	default:
		c.Line = "not yet"
	}
	if row.Official != nil {
		for _, m := range row.Official.Milestones {
			c.Milestones = append(c.Milestones, milestoneLink(m))
		}
	}
	return c
}

func milestoneLink(m official.Milestone) Link {
	switch {
	case m.PR != "":
		repo, num, _ := strings.Cut(m.PR, "#")
		return Link{Label: m.Label, URL: "https://github.com/" + repo + "/pull/" + num}
	case m.Repo != "":
		return Link{Label: m.Label, URL: "https://github.com/" + m.Repo}
	}
	return Link{Label: m.Label}
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
