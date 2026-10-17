package web

import (
	"fmt"
	"slices"
	"strings"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/official"
	"github.com/nmloc/gh-achievements/internal/progress"
	"github.com/nmloc/gh-achievements/internal/report"
	"github.com/nmloc/gh-achievements/internal/webui"
	"github.com/nmloc/gh-achievements/internal/wording"
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
	Standalone                       bool // the full /result page: @login is its <h1>
	Earned, EarnableTotal, TierTotal int
	Closest                          *Card
	Display                          []Card // "On display": every earned badge, any kind
	Collect                          []Card // "Still to collect": earnable, not earned
	Archive                          []Card // "The archive": cannot be earned, not earned
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
