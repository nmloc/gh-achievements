package web

import (
	"slices"
	"strings"
	"testing"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/counters"
	"github.com/nmloc/gh-achievements/internal/fixtures"
	"github.com/nmloc/gh-achievements/internal/official"
	"github.com/nmloc/gh-achievements/internal/progress"
	"github.com/nmloc/gh-achievements/internal/report"
)

func demoView(t *testing.T, private bool) ResultView {
	t.Helper()
	r, err := report.Build("torvalds", fixtures.Official{}, fixtures.TorvaldsDemo())
	if err != nil {
		t.Fatal(err)
	}
	return NewResultView(r, private)
}

func card(t *testing.T, cards []Card, slug string) Card {
	t.Helper()
	for _, c := range cards {
		if c.Slug == slug {
			return c
		}
	}
	t.Fatalf("no card %s", slug)
	return Card{}
}

func TestResultView_summary(t *testing.T) {
	v := demoView(t, false)
	// torvalds: Pull Shark x2, Starstruck x4, Pair x4 among earnable; Arctic, Mars retired.
	if v.Earned != 3 || v.EarnableTotal != 8 || v.TierTotal != 12 || v.Private {
		t.Fatalf("summary = earned %d of %d, tiers %d", v.Earned, v.EarnableTotal, v.TierTotal)
	}
	if v.Avatar != "https://github.com/torvalds.png?size=120" {
		t.Fatalf("avatar = %s", v.Avatar)
	}
}

func TestResultView_closestNextUnlockIsTheSmallestRemaining(t *testing.T) {
	v := demoView(t, false)
	// Galaxy Brain needs 2, Pull Shark 56: Galaxy Brain is closest.
	if v.Closest == nil || v.Closest.Slug != "galaxy-brain" || v.Closest.Line != "0 / 2 answers to x1" {
		t.Fatalf("closest = %+v", v.Closest)
	}
}

func TestResultView_quickWinsAreUnearnedSingleTierBadges(t *testing.T) {
	v := demoView(t, false)
	var slugs []string
	for _, c := range v.QuickWins {
		slugs = append(slugs, c.Slug)
	}
	if strings.Join(slugs, ",") != "quickdraw,yolo,public-sponsor" {
		t.Fatalf("quick wins = %v", slugs)
	}
}

func TestResultView_cards(t *testing.T) {
	v := demoView(t, false)
	ps := card(t, v.Earnable, "pull-shark")
	if !ps.Earned || ps.TierLabel != "x2" || !strings.Contains(ps.Image, "-bronze-") || ps.Source != "calibrated" ||
		ps.Progress == nil || ps.Progress.Value != 56 || ps.Progress.Max != 112 || ps.Line != "72 / 128 to x3" {
		t.Fatalf("pull shark = %+v %+v", ps, ps.Progress)
	}
	if ps.ImageAlt != "Pull Shark, bronze tier (x2)" {
		t.Fatalf("alt = %q", ps.ImageAlt)
	}
	if len(ps.Milestones) == 0 || ps.Milestones[1].URL != "https://github.com/subsurface/subsurface/pull/1506" {
		t.Fatalf("milestones = %+v", ps.Milestones)
	}
	gb := card(t, v.Earnable, "galaxy-brain")
	if gb.Earned || !gb.Locked || !strings.Contains(gb.Image, "-default-") {
		t.Fatalf("galaxy brain = %+v", gb)
	}
	if mars := card(t, v.Other, "mars-2020-contributor"); !mars.Earned || mars.TierLabel != "✓" {
		t.Fatalf("mars = %+v", mars)
	}
	// GitHub deleted this art; the pinned community copy stands in (catalog.communityArt).
	if heart := card(t, v.Other, "heart-on-your-sleeve"); !strings.HasSuffix(heart.Image, "/heart-on-your-sleeve-default.png") || !heart.Locked {
		t.Fatalf("heart = %+v", heart)
	}
}

func TestResultView_hiddenProfileHasNoQuickWins(t *testing.T) {
	r, err := report.Build("sindresorhus", fixtures.Official{}, &fixtures.Counts{Authed: true,
		PRs: counters.MergedPRs{Total: 2241}, FactsValue: &counters.Facts{}})
	if err != nil {
		t.Fatal(err)
	}
	v := NewResultView(r, false)
	if !v.Hidden || len(v.QuickWins) != 0 {
		t.Fatalf("hidden %v quick wins %v", v.Hidden, v.QuickWins)
	}
}

func TestCard_nonEarnableKindsNeverShowProgress(t *testing.T) {
	v := demoView(t, false)
	heart := card(t, v.Other, "heart-on-your-sleeve")
	if heart.Line != "Switched off by GitHub" || heart.Progress != nil || strings.Contains(heart.Line, "0") {
		t.Fatalf("heart = %q %+v", heart.Line, heart.Progress)
	}
	// Every fixture profile has the Arctic badge, so build the unearned rows directly.
	for kind, want := range map[catalog.Kind]string{
		catalog.Disabled: "Switched off by GitHub",
		catalog.Retired:  "No longer earnable",
		catalog.Internal: "GitHub staff only",
	} {
		row := report.Row{Slug: "x", Name: "X Badge", Kind: kind, Tiers: []int{0, 16, 128, 0}, Status: report.NotEarned}
		if c := newCard(row); c.Line != want || c.Progress != nil {
			t.Errorf("%s: line %q progress %+v, want %q", kind, c.Line, c.Progress, want)
		}
	}
	earned := report.Row{Kind: catalog.Retired, Name: "X", Tiers: []int{1}, Status: report.Earned, Unlocked: "Jan 2021"}
	if c := newCard(earned); c.Line != "Since 2021" {
		t.Errorf("earned retired line = %q", c.Line)
	}
	earned.Unlocked = ""
	if c := newCard(earned); c.Line != "Earned" {
		t.Errorf("earned retired line = %q", c.Line)
	}
}

func TestResultView_hiddenProfileKeepsCountedClosest(t *testing.T) {
	r, err := report.Build("sindresorhus", fixtures.Official{}, &fixtures.Counts{Authed: true,
		PRs: counters.MergedPRs{Total: 2241}, FactsValue: &counters.Facts{}})
	if err != nil {
		t.Fatal(err)
	}
	v := NewResultView(r, false)
	if v.Closest == nil || v.Closest.Line == "Hidden on this profile" || !strings.Contains(v.Closest.Line, " to x") {
		t.Fatalf("closest = %+v", v.Closest)
	}
	if v.Closest.Progress == nil || !strings.Contains(v.Closest.ImageAlt, "status hidden") {
		t.Fatalf("closest = %+v", v.Closest)
	}
	// A hidden row with no count still says so.
	c := newCard(report.Row{Name: "Quickdraw", Kind: catalog.Earnable, Tiers: []int{1}, Status: report.Unknown})
	if c.Line != "Hidden on this profile" || c.ImageAlt != "Quickdraw badge, status hidden" {
		t.Fatalf("unknown without count = %q %q", c.Line, c.ImageAlt)
	}
}

func TestCard_aheadOfGitHubExplainsItself(t *testing.T) {
	row := report.Row{Slug: "pull-shark", Name: "Pull Shark", Kind: catalog.Earnable, Tiers: []int{2, 16, 128, 1024},
		Status: report.Earned, Tier: 2,
		Standing: &progress.Standing{Tier: 3, Count: 130, Floor: 128, Target: 1024, Remaining: 894, IsAhead: true}}
	c := newCard(row)
	want := "GitHub shows x2; the count says x3"
	if len(c.Notes) == 0 || c.Notes[len(c.Notes)-1] != want {
		t.Fatalf("notes = %v", c.Notes)
	}
}

func TestMilestoneLink_onlyWellFormedReferencesBecomeLinks(t *testing.T) {
	for _, pr := range []string{"me/repo#12a", "norepo#3", "a/b/c#3", "me/repo#", "/repo#3", "me/#3"} {
		if l := milestoneLink(official.Milestone{PR: pr, Label: "L"}); l.URL != "" || l.Label != "L" {
			t.Errorf("%q -> %+v", pr, l)
		}
	}
	if l := milestoneLink(official.Milestone{PR: "me/repo#12"}); l.URL != "https://github.com/me/repo/pull/12" {
		t.Errorf("good PR -> %+v", l)
	}
	if l := milestoneLink(official.Milestone{Repo: "a/b/c"}); l.URL != "" {
		t.Errorf("bad repo -> %+v", l)
	}
	if l := milestoneLink(official.Milestone{Repo: "a/b"}); l.URL != "https://github.com/a/b" {
		t.Errorf("good repo -> %+v", l)
	}
}

func TestCard_meterMaxIsNeverZero(t *testing.T) {
	row := report.Row{Name: "X", Kind: catalog.Earnable, Tiers: []int{5, 10}, Status: report.Earned, Tier: 1,
		Standing: &progress.Standing{Tier: 1, Count: 5, Floor: 5, Target: 5, Remaining: 1}}
	if c := newCard(row); c.Progress == nil || c.Progress.Max < 1 {
		t.Fatalf("progress = %+v", c.Progress)
	}
}

func TestCard_notesSpeakForTheWeb(t *testing.T) {
	row := report.Row{Slug: "pull-shark", Name: "Pull Shark", Kind: catalog.Earnable, Tiers: []int{2, 16, 128, 1024},
		Status: report.Earned, Tier: 2, Notes: []string{
			"checked at GitHub's #16 org/repo: GitHub counts +3 vs what your gh login can see",
			"run `gh auth login` to check this against GitHub's milestones",
			"run `gh auth login` to count accepted answers",
		}}
	c := newCard(row)
	want := []string{"checked at GitHub's #16 org/repo: GitHub counts +3 vs what this check can see"}
	if !slices.Equal(c.Notes, want) {
		t.Fatalf("notes = %q, want %q", c.Notes, want)
	}
	if len(row.Notes) != 3 || !strings.Contains(row.Notes[0], "gh login") {
		t.Fatalf("the report row was modified: %q", row.Notes)
	}
}

func TestCard_lineShowsTheCountAgainstTheNextThreshold(t *testing.T) {
	pullShark := []int{2, 16, 128, 1024}
	stars := []int{16, 128, 512, 4096}
	cases := []struct {
		name string
		row  report.Row
		want string
	}{
		{"count / target", report.Row{Slug: "pull-shark", Kind: catalog.Earnable, Tiers: pullShark, Status: report.Earned, Tier: 2,
			Standing: &progress.Standing{Tier: 2, Count: 72, Floor: 16, Target: 128, Remaining: 56}}, "72 / 128 to x3"},
		{"with a unit", report.Row{Slug: "galaxy-brain", Kind: catalog.Earnable, Tiers: []int{2, 8, 16, 32}, Status: report.NotEarned,
			Standing: &progress.Standing{Tier: 0, Count: 0, Floor: 0, Target: 2, Remaining: 2}}, "0 / 2 answers to x1"},
		{"thousands", report.Row{Slug: "starstruck", Kind: catalog.Earnable, Tiers: stars, Status: report.Earned, Tier: 3,
			Standing: &progress.Standing{Tier: 3, Count: 1200, Floor: 512, Target: 4096, Remaining: 2896}}, "1,200 / 4,096 stars to x4"},
		{"top tier with a unit", report.Row{Slug: "starstruck", Kind: catalog.Earnable, Tiers: stars, Status: report.Earned, Tier: 4,
			Standing: &progress.Standing{Tier: 4, Count: 252192, Floor: 4096, IsMax: true}}, "Top tier, 252,192 stars"},
		{"top tier without a unit", report.Row{Slug: "pull-shark", Kind: catalog.Earnable, Tiers: pullShark, Status: report.Earned, Tier: 4,
			Standing: &progress.Standing{Tier: 4, Count: 1500, Floor: 1024, IsMax: true}}, "Top tier"},
		{"tiered, no counter", report.Row{Slug: "pair-extraordinaire", Kind: catalog.Earnable, Tiers: []int{1, 10, 24, 48},
			Status: report.Earned, Tier: 2}, "x2, next tier at 24"},
		{"tiered, no counter, top", report.Row{Slug: "pair-extraordinaire", Kind: catalog.Earnable, Tiers: []int{1, 10, 24, 48},
			Status: report.Earned, Tier: 4}, "Top tier"},
		{"tiered, not started", report.Row{Slug: "finishing-touch", Kind: catalog.Earnable, Tiers: []int{1, 4, 16, 32},
			Status: report.NotEarned}, "First tier at 1"},
		{"tiered, first threshold unknown", report.Row{Slug: "x", Kind: catalog.Earnable, Tiers: []int{0, 16, 128, 0},
			Status: report.NotEarned}, "First tier: threshold unknown"},
		{"next threshold unknown", report.Row{Slug: "x", Kind: catalog.Earnable, Tiers: []int{0, 16, 128, 0},
			Status: report.Earned, Tier: 3}, "x3, next tier unknown"},
		{"single tier, earned", report.Row{Slug: "quickdraw", Kind: catalog.Earnable, Tiers: []int{1}, Status: report.Earned,
			Tier: 1, Unlocked: "2020-02-02"}, "Since 2020"},
		{"single tier, earned, no date", report.Row{Slug: "quickdraw", Kind: catalog.Earnable, Tiers: []int{1},
			Status: report.Earned, Tier: 1}, "Earned"},
		{"single tier, not earned", report.Row{Slug: "quickdraw", Kind: catalog.Earnable, Tiers: []int{1}, Status: report.NotEarned,
			HowToEarn: "Close an issue or pull request within 5 minutes of opening it"},
			"Close an issue or pull request within 5 minutes of opening it"},
	}
	for _, tc := range cases {
		tc.row.Name = "Badge"
		if got := newCard(tc.row).Line; got != tc.want {
			t.Errorf("%s: line = %q, want %q", tc.name, got, tc.want)
		}
	}
}
