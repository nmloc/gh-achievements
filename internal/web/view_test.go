package web

import (
	"strings"
	"testing"

	"github.com/nmloc/gh-achievements/internal/counters"
	"github.com/nmloc/gh-achievements/internal/fixtures"
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
	if v.Closest == nil || v.Closest.Slug != "galaxy-brain" || !strings.Contains(v.Closest.Line, "2 to x1") {
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
		ps.Progress == nil || ps.Progress.Value != 56 || ps.Progress.Max != 112 || ps.Line != "56 to x3" {
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
	if heart := card(t, v.Other, "heart-on-your-sleeve"); heart.Image != "" || heart.Initials != "HO" {
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
