package report

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/nmloc/gh-achievements/internal/counters"
	"github.com/nmloc/gh-achievements/internal/fixtures"
	"github.com/nmloc/gh-achievements/internal/official"
	"github.com/nmloc/gh-achievements/internal/progress"
)

type fakeCounts = fixtures.Counts

func build(t *testing.T, user string, counts *fakeCounts) *Report {
	t.Helper()
	r, err := Build(user, fixtures.Official{}, counts)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func row(t *testing.T, r *Report, slug string) Row {
	t.Helper()
	for _, row := range r.Achievements {
		if row.Slug == slug {
			return row
		}
	}
	t.Fatalf("no row %s", slug)
	return Row{}
}

// brannonPRs puts DapperLib/Dapper#32 second, as GitHub's own history says it was.
func brannonPRs(total int) counters.MergedPRs {
	refs := make([]string, total)
	for i := range refs {
		refs[i] = "o/r#" + string(rune('a'+i%26)) + strings.Repeat("x", i/26)
	}
	refs[1] = "DapperLib/Dapper#32"
	return counters.MergedPRs{Total: total, Refs: refs}
}

func TestPullShark_calibratedAtPublicMilestone(t *testing.T) {
	r := build(t, "brannon", &fakeCounts{Authed: true, PRs: brannonPRs(155), FactsValue: &counters.Facts{}})
	ps := row(t, r, "pull-shark")
	if ps.Source != progress.Calibrated || ps.Standing.Count != 155 || ps.Standing.Target != 1024 {
		t.Fatalf("got %+v %+v", ps, ps.Standing)
	}
	notes := strings.Join(ps.Notes, "\n")
	for _, want := range []string{"checked at GitHub's #2 DapperLib/Dapper#32", "GitHub's #16 is private", "GitHub's #128 is private"} {
		if !strings.Contains(notes, want) {
			t.Errorf("notes %q lack %q", notes, want)
		}
	}
}

func TestPullShark_pastTheSearchCapIsNotCalibrated(t *testing.T) {
	prs := brannonPRs(1000)
	prs.Total = 2241 // search lists 1000, issueCount says 2241
	r := build(t, "brannon", &fakeCounts{Authed: true, PRs: prs, FactsValue: &counters.Facts{}})
	ps := row(t, r, "pull-shark")
	if ps.Source != progress.Counted || ps.Standing.Count != 2241 {
		t.Fatalf("got %+v %+v", ps, ps.Standing)
	}
	notes := strings.Join(ps.Notes, "\n")
	if !strings.Contains(notes, "1000") || strings.Contains(notes, "private") || strings.Contains(notes, "checked at") {
		t.Fatalf("notes = %q", notes)
	}
}

func TestWithoutGhLogin_countsPublicDataAndSaysSo(t *testing.T) {
	r := build(t, "brannon", &fakeCounts{Authed: false, PRs: counters.MergedPRs{Total: 150}})
	if r.Authenticated {
		t.Fatal("report claims authentication")
	}
	ps := row(t, r, "pull-shark")
	if ps.Source != progress.Counted || !strings.Contains(strings.Join(ps.Notes, " "), "gh auth login") {
		t.Fatalf("pull shark = %+v", ps)
	}
	gb := row(t, r, "galaxy-brain")
	if gb.Standing != nil || !strings.Contains(strings.Join(gb.Notes, " "), "gh auth login") {
		t.Fatalf("galaxy brain = %+v", gb)
	}
}

func TestHiddenProfile_claimsNothingOfficial(t *testing.T) {
	// Regression: a hidden profile showed Pair Extraordinaire as "[official]".
	r := build(t, "sindresorhus", &fakeCounts{Authed: true, PRs: counters.MergedPRs{Total: 2241}, FactsValue: &counters.Facts{Sponsoring: 3}})
	if !r.AchievementsHidden {
		t.Fatal("not marked hidden")
	}
	for _, row := range r.Achievements {
		if row.Status != Unknown {
			t.Errorf("%s status = %s, want unknown", row.Slug, row.Status)
		}
		if row.Source == progress.Official {
			t.Errorf("%s claims an official source on a hidden profile", row.Slug)
		}
	}
	if ps := row(t, r, "pull-shark"); ps.Standing.Tier != 4 || !ps.Standing.IsAhead {
		t.Fatalf("pull shark = %+v", ps.Standing)
	}
	if sp := row(t, r, "public-sponsor"); len(sp.Notes) != 0 {
		t.Fatalf("hidden profile guessed about sponsorship: %q", sp.Notes)
	}
}

func TestPairExtraordinaire_neverCounted(t *testing.T) {
	// Regression: counting Co-authored-by commits gave 60-79 where GitHub's tier says
	// 10-23. The badge must show GitHub's tier and no number.
	r := build(t, "brannon", &fakeCounts{Authed: true, PRs: brannonPRs(155), FactsValue: &counters.Facts{}})
	pe := row(t, r, "pair-extraordinaire")
	if pe.Standing != nil || pe.Source != progress.Official || pe.Tier != 3 {
		t.Fatalf("got %+v", pe)
	}
}

func TestStarstruck_usesTheRepositoriesGitHubNames(t *testing.T) {
	counts := &fakeCounts{Authed: true, FactsValue: &counters.Facts{},
		Owned: 3000, OwnedRepo: "ljharb/qs",
		RepoStar: map[string]int{"es-shims/es6-shim": 3100, "ljharb/qs": 3000, "enzymejs/enzyme": 19900}}
	ss := row(t, build(t, "ljharb", counts), "starstruck")
	if ss.Source != progress.Counted || ss.Standing.Count != 19900 || !ss.Standing.IsMax {
		t.Fatalf("got %+v %+v", ss, ss.Standing)
	}
	if !strings.Contains(ss.Notes[0], "enzymejs/enzyme (named in GitHub's history)") {
		t.Fatalf("notes = %q", ss.Notes)
	}
}

func TestStarstruck_ownedRepoBeatsTheNamedOnes(t *testing.T) {
	counts := &fakeCounts{Authed: true, FactsValue: &counters.Facts{},
		Owned: 500, OwnedRepo: "brannon/new",
		RepoStar: map[string]int{"github/gh-models": 400}}
	ss := row(t, build(t, "brannon", counts), "starstruck")
	if ss.Source != progress.Counted || ss.Standing.Count != 500 {
		t.Fatalf("got %+v %+v", ss, ss.Standing)
	}
}

func TestPublicSponsor_privateSponsorshipHint(t *testing.T) {
	r := build(t, "torvalds", &fakeCounts{Authed: true, PRs: counters.MergedPRs{Total: 72}, FactsValue: &counters.Facts{Sponsoring: 1}})
	if sp := row(t, r, "public-sponsor"); sp.Status != NotEarned || len(sp.Notes) != 1 {
		t.Fatalf("got %+v", sp)
	}
}

func TestEveryCatalogBadgeAppears_andRetiredOnesAreRead(t *testing.T) {
	r := build(t, "torvalds", &fakeCounts{Authed: true, PRs: counters.MergedPRs{Total: 72}, FactsValue: &counters.Facts{}})
	if len(r.Achievements) != 15 {
		t.Fatalf("%d achievements, want 15", len(r.Achievements))
	}
	if m := row(t, r, "mars-2020-contributor"); m.Status != Earned || m.Unlocked != "2021-04-19" {
		t.Fatalf("mars = %+v", m)
	}
}

func TestHighlights_fromPageAndFromGraphQL(t *testing.T) {
	r := build(t, "ljharb", &fakeCounts{Authed: true, PRs: counters.MergedPRs{Total: 1}, FactsValue: &counters.Facts{
		Flags: map[string]bool{"isBountyHunter": true},
	}})
	got := map[string]bool{}
	for _, h := range r.Highlights {
		got[h.Name] = h.Earned
	}
	want := map[string]bool{
		"Developer Program Member": true, "Security Bug Bounty Hunter": true, "GitHub Campus Expert": false,
		"GitHub Star": false, "Pro": true, "Security advisory credit": true,
	}
	for name, w := range want {
		if got[name] != w {
			t.Errorf("%s = %v, want %v", name, got[name], w)
		}
	}
}

type extraBadge struct{ fixtures.Official }

func (e extraBadge) Profile(user string) (official.Profile, error) {
	p, err := e.Official.Profile(user)
	p.Earned = append(p.Earned, "brand-new-badge")
	return p, err
}

func (e extraBadge) Detail(user, slug string) (*official.Detail, error) {
	if slug == "brand-new-badge" {
		return &official.Detail{Slug: slug, Name: "Brand New", Tier: 2}, nil
	}
	return e.Official.Detail(user, slug)
}

func TestUnknownBadgeOnProfileIsListedAsNew(t *testing.T) {
	r, err := Build("brannon", extraBadge{}, &fakeCounts{Authed: true, PRs: brannonPRs(155), FactsValue: &counters.Facts{}})
	if err != nil {
		t.Fatal(err)
	}
	nb := row(t, r, "brand-new-badge")
	if nb.Kind != KindNew || nb.Name != "Brand New" || nb.Tier != 2 {
		t.Fatalf("got %+v", nb)
	}
}

// missingDetail serves fixtures, except that some badges have no detail or fail.
type missingDetail struct {
	fixtures.Official
	none map[string]bool
	fail map[string]error
}

func (m missingDetail) Detail(user, slug string) (*official.Detail, error) {
	if err := m.fail[slug]; err != nil {
		return nil, err
	}
	if m.none[slug] {
		return nil, nil
	}
	return m.Official.Detail(user, slug)
}

func TestListedBadgeWithoutHistory_isStillEarned(t *testing.T) {
	// Regression: earned came only from the detail fragment, so a badge the profile
	// shows but whose history 404s was printed as "0 / 1 [official]".
	off := missingDetail{none: map[string]bool{"yolo": true}}
	r, err := Build("brannon", off, &fakeCounts{Authed: true, PRs: brannonPRs(155), FactsValue: &counters.Facts{}})
	if err != nil {
		t.Fatal(err)
	}
	y := row(t, r, "yolo")
	if y.Status != Earned || y.Tier != 1 || !strings.Contains(strings.Join(y.Notes, " "), "served no history") {
		t.Fatalf("yolo = %+v", y)
	}
}

func TestOneFailedHistory_degradesOnlyItsRow(t *testing.T) {
	// Regression: one 429 on a detail fragment aborted the whole report.
	off := missingDetail{fail: map[string]error{"yolo": errors.New("HTTP 429")}}
	r, err := Build("brannon", off, &fakeCounts{Authed: true, PRs: brannonPRs(155), FactsValue: &counters.Facts{}})
	if err != nil {
		t.Fatalf("Build failed on one badge: %v", err)
	}
	if y := row(t, r, "yolo"); y.Status != Earned || !strings.Contains(strings.Join(y.Notes, " "), "HTTP 429") {
		t.Fatalf("yolo = %+v", y)
	}
	if ps := row(t, r, "pull-shark"); ps.Source != progress.Calibrated {
		t.Fatalf("pull shark lost its calibration: %+v", ps)
	}
}

func TestStarstruck_deletedNamedRepoDoesNotWipeTheRow(t *testing.T) {
	counts := &fakeCounts{Authed: true, FactsValue: &counters.Facts{},
		Owned: 20, OwnedRepo: "brannon/small",
		RepoErr: map[string]error{"github/gh-models": errors.New("HTTP 404")}}
	ss := row(t, build(t, "brannon", counts), "starstruck")
	if ss.Standing == nil || ss.Source != progress.Counted {
		t.Fatalf("got %+v", ss)
	}
	notes := strings.Join(ss.Notes, " ")
	if !strings.Contains(notes, "could not read github/gh-models") || !strings.Contains(notes, "brannon/small has 20 stars") {
		t.Fatalf("notes = %q", notes)
	}
}

func TestRaisedFloor_saysTheAPISawLess(t *testing.T) {
	// torvalds holds Pull Shark x2 (16+); an API that sees 0 must not pass 16 off as a count.
	r := build(t, "torvalds", &fakeCounts{Authed: true, PRs: counters.MergedPRs{Total: 0}, FactsValue: &counters.Facts{}})
	ps := row(t, r, "pull-shark")
	if ps.Standing.Count != 16 || !strings.Contains(strings.Join(ps.Notes, " "), "the API sees 0") {
		t.Fatalf("got %+v %q", ps.Standing, ps.Notes)
	}
}

func TestPrivateMilestoneGuess_onlyAfterCalibration(t *testing.T) {
	// Regression: with every visible milestone unmatched, it still guessed positions
	// with offset 0, which is a calibration claim nothing checked.
	refs := make([]string, 200)
	for i := range refs {
		refs[i] = fmt.Sprintf("x/y#%d", i)
	}
	r := build(t, "brannon", &fakeCounts{Authed: true, PRs: counters.MergedPRs{Total: 200, Refs: refs}, FactsValue: &counters.Facts{}})
	ps := row(t, r, "pull-shark")
	if ps.Source != progress.Counted || strings.Contains(strings.Join(ps.Notes, " "), "private") {
		t.Fatalf("got %+v %q", ps.Source, ps.Notes)
	}
}

func TestGalaxyBrain_warnsTheCountCanRunHigh(t *testing.T) {
	r := build(t, "ljharb", &fakeCounts{Authed: true, PRs: counters.MergedPRs{Total: 1}, FactsValue: &counters.Facts{AcceptedAnswers: 40}})
	if gb := row(t, r, "galaxy-brain"); !strings.Contains(strings.Join(gb.Notes, " "), "2024-02-06") {
		t.Fatalf("notes = %q", gb.Notes)
	}
}
