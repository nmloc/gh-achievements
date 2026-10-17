// Package report joins GitHub's official badge data with live counts into one report.
package report

import (
	"fmt"
	"sync"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/counters"
	"github.com/nmloc/gh-achievements/internal/official"
	"github.com/nmloc/gh-achievements/internal/progress"
)

// KindNew marks a badge on the profile that the catalog does not know yet.
const KindNew catalog.Kind = "new"

// Status of a badge for this user.
type Status string

const (
	Earned    Status = "earned"
	NotEarned Status = "not-earned"
	Unknown   Status = "unknown" // the profile hides achievements
)

// Row is one achievement in the report.
type Row struct {
	Slug      string             `json:"slug"`
	Name      string             `json:"name"`
	Kind      catalog.Kind       `json:"kind"`
	HowToEarn string             `json:"howToEarn"`
	Tiers     []int              `json:"tiers"`
	Status    Status             `json:"status"`
	Tier      int                `json:"officialTier"` // GitHub's tier, 0 if not earned or hidden
	Unlocked  string             `json:"unlocked,omitempty"`
	Standing  *progress.Standing `json:"standing,omitempty"`
	Source    progress.Source    `json:"source,omitempty"`
	Notes     []string           `json:"notes,omitempty"`
	Official  *official.Detail   `json:"official,omitempty"`
	Counter   catalog.Counter    `json:"counter"`
}

// HighlightRow is one profile highlight.
type HighlightRow struct {
	Name   string `json:"name"`
	Earned bool   `json:"earned"`
}

// Report is everything the CLI prints, and the --json output.
type Report struct {
	User               string         `json:"user"`
	AchievementsHidden bool           `json:"achievementsHidden"`
	Authenticated      bool           `json:"authenticated"`
	CatalogVerified    string         `json:"catalogVerified"`
	Achievements       []Row          `json:"achievements"`
	Highlights         []HighlightRow `json:"highlights"`
}

// Official is where GitHub's own data comes from.
type Official interface {
	Profile(user string) (official.Profile, error)
	Detail(user, slug string) (*official.Detail, error)
}

// Build makes the report for user.
func Build(user string, off Official, counts counters.Source) (*Report, error) {
	profile, err := off.Profile(user)
	if err != nil {
		return nil, err
	}
	details, detailErrs := fetchDetails(user, off, profile.Earned)
	listed := make(map[string]bool, len(profile.Earned))
	for _, slug := range profile.Earned {
		listed[slug] = true
	}

	var flagFields []string
	for _, h := range catalog.Highlights {
		if h.GraphQLField != "" {
			flagFields = append(flagFields, h.GraphQLField)
		}
	}
	facts, factsErr := counts.Facts(user, flagFields)

	r := &Report{
		User:               user,
		AchievementsHidden: profile.Hidden,
		Authenticated:      counts.IsAuthenticated(),
		CatalogVerified:    catalog.LastVerified,
	}
	b := builder{user: user, counts: counts, facts: facts, factsErr: factsErr, hidden: profile.Hidden,
		listed: listed, detailErrs: detailErrs}

	known := map[string]bool{}
	for _, a := range catalog.Achievements {
		known[a.Slug] = true
		r.Achievements = append(r.Achievements, b.row(a, details[a.Slug]))
	}
	for _, slug := range profile.Earned {
		if known[slug] {
			continue
		}
		d := details[slug]
		var notes []string
		if d == nil {
			notes = append(notes, b.missingDetailNote(slug))
			d = &official.Detail{Slug: slug, Name: slug, Tier: 1}
		}
		r.Achievements = append(r.Achievements, Row{
			Slug: slug, Name: d.Name, Kind: KindNew, HowToEarn: "Not in the catalog yet: please open an issue",
			Status: Earned, Tier: d.Tier, Unlocked: d.Unlocked, Official: details[slug], Source: progress.Official,
			Counter: catalog.CounterNone, Notes: notes,
		})
	}

	for _, h := range catalog.Highlights {
		has := official.HasHighlight(profile.Highlights, h.Name)
		if facts != nil && h.GraphQLField != "" && facts.Flags[h.GraphQLField] {
			has = true
		}
		r.Highlights = append(r.Highlights, HighlightRow{Name: h.Name, Earned: has})
	}
	return r, nil
}

// fetchDetails reads each badge's history. A failure belongs to its badge only: the
// row says so, and the rest of the report still prints.
func fetchDetails(user string, off Official, slugs []string) (map[string]*official.Detail, map[string]error) {
	details := make(map[string]*official.Detail, len(slugs))
	errs := map[string]error{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4) // be polite to github.com
	for _, slug := range slugs {
		wg.Add(1)
		go func(slug string) {
			defer wg.Done()
			// Deferred after wg.Done, so it runs before it. A panic here would end the
			// whole process (and every in-flight request's deferred token revoke), so it
			// becomes this badge's failure: the type only, never the value.
			defer func() {
				if p := recover(); p != nil {
					mu.Lock()
					errs[slug] = fmt.Errorf("detail panic: %T", p)
					mu.Unlock()
				}
			}()
			sem <- struct{}{}
			defer func() { <-sem }()
			d, err := off.Detail(user, slug)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs[slug] = err
			}
			if d != nil {
				details[slug] = d
			}
		}(slug)
	}
	wg.Wait()
	return details, errs
}

type builder struct {
	user       string
	counts     counters.Source
	facts      *counters.Facts
	factsErr   error
	hidden     bool
	listed     map[string]bool // slugs the profile shows as earned
	detailErrs map[string]error

	prsOnce sync.Once
	prs     counters.MergedPRs
	prsErr  error
}

func (b *builder) mergedPRs() (counters.MergedPRs, error) {
	b.prsOnce.Do(func() { b.prs, b.prsErr = b.counts.MergedPRs(b.user) })
	return b.prs, b.prsErr
}

func (b *builder) row(a catalog.Achievement, d *official.Detail) Row {
	row := Row{
		Slug: a.Slug, Name: a.Name, Kind: a.Kind, HowToEarn: a.HowToEarn, Tiers: a.Tiers,
		Official: d, Counter: a.Counter, Status: NotEarned,
	}
	switch {
	case d != nil:
		row.Status, row.Tier, row.Unlocked, row.Name = Earned, d.Tier, d.Unlocked, d.Name
	case b.listed[a.Slug]:
		// The profile shows it, so it is earned; only the tier is unknown beyond x1.
		row.Status, row.Tier = Earned, 1
		row.Notes = append(row.Notes, b.missingDetailNote(a.Slug))
	case b.hidden:
		row.Status = Unknown
	}
	if a.Kind != catalog.Earnable {
		return row
	}
	if !a.IsTiered() {
		b.single(&row)
		return row
	}
	switch a.Counter {
	case catalog.CounterMergedPRs:
		b.pullShark(&row)
	case catalog.CounterStars:
		b.starstruck(&row)
	case catalog.CounterAnswers:
		b.galaxyBrain(&row)
	default:
		// A hidden profile has no official tier either, so nothing here is "official".
		if b.hidden {
			row.Notes = append(row.Notes, "hidden on this profile, and GitHub's rule is not public: no data")
		} else {
			row.Source = progress.Official
			row.Notes = append(row.Notes, a.NoCounterReason)
		}
	}
	return row
}

func (b *builder) missingDetailNote(slug string) string {
	if err := b.detailErrs[slug]; err != nil {
		return "could not read GitHub's history for this badge (" + err.Error() + "): tier shown as at least x1"
	}
	return "GitHub lists this badge but served no history for it: tier shown as at least x1"
}

func (b *builder) single(row *Row) {
	if row.Status == Unknown {
		return
	}
	row.Source = progress.Official
	if row.Slug == "public-sponsor" && row.Status == NotEarned && b.facts != nil && b.facts.Sponsoring > 0 {
		row.Notes = append(row.Notes, "sponsors someone, but no badge: only a public sponsorship earns it")
	}
}

func (b *builder) stand(row *Row, count int, source progress.Source) {
	s := progress.Stand(row.Tiers, row.Tier, count)
	row.Standing, row.Source = &s, source
	if s.Count > count {
		row.Notes = append(row.Notes, fmt.Sprintf(
			"the API sees %d; GitHub's x%d guarantees at least %d, so %d is shown", count, row.Tier, s.Floor, s.Count))
	}
}

func (b *builder) pullShark(row *Row) {
	prs, err := b.mergedPRs()
	if err != nil {
		row.Notes = append(row.Notes, "could not count: "+err.Error())
		return
	}
	var milestones []official.Milestone
	if row.Official != nil {
		milestones = row.Official.Milestones
	}
	var cal *progress.Calibration
	if b.counts.IsAuthenticated() && prs.IsComplete() {
		cal = progress.Calibrate(prs.Refs, milestones)
	}
	offset := 0
	switch {
	case cal != nil:
		offset = cal.Offset
		b.stand(row, prs.Total+offset, progress.Calibrated)
		row.Notes = append(row.Notes, fmt.Sprintf(
			"checked at GitHub's #%d %s: GitHub counts %+d vs what your gh login can see", cal.At, cal.PR, cal.Offset))
	default:
		b.stand(row, prs.Total, progress.Counted)
		switch {
		case !b.counts.IsAuthenticated():
			row.Notes = append(row.Notes, "run `gh auth login` to check this against GitHub's milestones")
		case !prs.IsComplete():
			row.Notes = append(row.Notes, "more than 1000 PRs: search cannot list them all, so this is not checked")
		}
	}
	if cal == nil {
		return // without a checked offset, a position guess would be a calibration claim
	}
	for _, m := range milestones {
		if m.N == 0 || m.PR != "" {
			continue
		}
		if i := m.N - offset; i > 0 && i <= len(prs.Refs) {
			row.Notes = append(row.Notes, fmt.Sprintf(
				"GitHub's #%d is private, most likely %s", m.N, prs.Refs[i-1]))
		}
	}
}

// starstruck takes the most-starred repository among the ones GitHub's history names
// (they can be organisation repositories the user created, e.g. enzymejs/enzyme) and
// the ones the user owns.
func (b *builder) starstruck(row *Row) {
	best, bestRepo, named := -1, "", false
	if row.Official != nil {
		seen := map[string]bool{}
		for _, m := range row.Official.Milestones {
			if m.Repo == "" || seen[m.Repo] {
				continue
			}
			seen[m.Repo] = true
			stars, err := b.counts.RepoStars(m.Repo)
			if err != nil { // e.g. deleted since: the others still count
				row.Notes = append(row.Notes, fmt.Sprintf("could not read %s: %v", m.Repo, err))
				continue
			}
			if stars > best {
				best, bestRepo, named = stars, m.Repo, true
			}
		}
	}
	stars, repo, err := b.counts.MaxOwnedStars(b.user)
	if err != nil {
		row.Notes = append(row.Notes, "could not count: "+err.Error())
		return
	}
	if repo != "" && stars > best {
		best, bestRepo, named = stars, repo, false
	}
	// The star count is live API data even for a repository GitHub's history names.
	b.stand(row, max(best, 0), progress.Counted)
	if bestRepo != "" {
		which := ""
		if named {
			which = " (named in GitHub's history)"
		}
		row.Notes = append(row.Notes, fmt.Sprintf("%s%s has %d stars", bestRepo, which, best))
	}
}

func (b *builder) galaxyBrain(row *Row) {
	switch {
	case b.factsErr != nil:
		row.Notes = append(row.Notes, "could not count: "+b.factsErr.Error())
	case b.facts == nil:
		row.Notes = append(row.Notes, "run `gh auth login` to count accepted answers")
	default:
		b.stand(row, b.facts.AcceptedAnswers, progress.Counted)
		// The API cannot leave them out, so the count can run high.
		row.Notes = append(row.Notes, "includes answers in GitHub Community discussions, which stopped counting on 2024-02-06")
	}
}
