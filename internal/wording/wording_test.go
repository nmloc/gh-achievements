package wording

import (
	"strings"
	"testing"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/fixtures"
	"github.com/nmloc/gh-achievements/internal/progress"
	"github.com/nmloc/gh-achievements/internal/report"
)

func TestLine_countAgainstTheNextThreshold(t *testing.T) {
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
		{"switched off", report.Row{Slug: "x", Kind: catalog.Disabled, Tiers: []int{0, 16, 128, 0}, Status: report.NotEarned}, "Switched off by GitHub"},
		{"retired, earned", report.Row{Slug: "x", Kind: catalog.Retired, Tiers: []int{1}, Status: report.Earned, Tier: 1, Unlocked: "Jan 2021"}, "Since 2021"},
		{"hidden, no count", report.Row{Slug: "quickdraw", Kind: catalog.Earnable, Tiers: []int{1}, Status: report.Unknown}, "Hidden on this profile"},
	}
	for _, tc := range cases {
		tc.row.Name = "Badge"
		if got := Line(tc.row); got != tc.want {
			t.Errorf("%s: line = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func demo(t *testing.T) *report.Report {
	t.Helper()
	r, err := report.Build("torvalds", fixtures.Official{}, fixtures.TorvaldsDemo())
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestSummarize_matchesTheResultPage(t *testing.T) {
	s := Summarize(demo(t))
	if s.Earned != 3 || s.EarnableTotal != 8 || s.TierTotal != 12 {
		t.Fatalf("summary = %+v", s)
	}
	if s.Closest == nil || s.Closest.Slug != "galaxy-brain" || Line(*s.Closest) != "0 / 2 answers to x1" {
		t.Fatalf("closest = %+v", s.Closest)
	}
}

func TestShelfOf(t *testing.T) {
	got := map[Shelf][]string{}
	for _, row := range demo(t).Achievements {
		got[ShelfOf(row)] = append(got[ShelfOf(row)], row.Slug)
	}
	for shelf, want := range map[Shelf]string{
		Display: "pull-shark,starstruck,pair-extraordinaire,arctic-code-vault-contributor,mars-2020-contributor",
		Collect: "galaxy-brain,quickdraw,yolo,public-sponsor,finishing-touch",
		Archive: "heart-on-your-sleeve,open-sourcerer,proxima-pioneer,proxima-staffshipper,proxima-staffuser",
	} {
		if strings.Join(got[shelf], ",") != want {
			t.Errorf("shelf %d = %v", shelf, got[shelf])
		}
	}
}
