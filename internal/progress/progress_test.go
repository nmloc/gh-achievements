package progress

import (
	"testing"

	"github.com/nmloc/gh-achievements/internal/official"
)

var pullShark = []int{2, 16, 128, 1024}

func TestStand_betweenTiers(t *testing.T) {
	s := Stand(pullShark, 3, 155)
	if s.Tier != 3 || s.Count != 155 || s.Floor != 128 || s.Target != 1024 || s.Remaining != 869 || s.IsMax || s.IsAhead {
		t.Fatalf("got %+v", s)
	}
}

func TestStand_countBelowGrantedTierShowsTheFloor(t *testing.T) {
	// A tier GitHub granted is never lost: a lower count only means part is invisible.
	s := Stand(pullShark, 3, 100)
	if s.Count != 128 || s.Remaining != 896 {
		t.Fatalf("got %+v, want count raised to the x3 floor", s)
	}
}

func TestStand_hiddenProfileTakesTierFromCount(t *testing.T) {
	// Regression: with no official tier the MVP printed "2241/2 (-2239 to go)".
	s := Stand(pullShark, 0, 2241)
	if s.Tier != 4 || !s.IsMax || !s.IsAhead || s.Remaining != 0 {
		t.Fatalf("got %+v", s)
	}
}

func TestStand_neverNegative(t *testing.T) {
	for official := 0; official <= 4; official++ {
		for _, count := range []int{0, 1, 2, 15, 16, 17, 127, 128, 1023, 1024, 5000} {
			if s := Stand(pullShark, official, count); s.Remaining < 0 {
				t.Fatalf("Stand(%d, %d) = %+v", official, count, s)
			}
		}
	}
}

func TestTierFromCount_skipsUnknownThresholds(t *testing.T) {
	heart := []int{0, 16, 128, 0} // "?" thresholds are 0
	if got := TierFromCount(heart, 20); got != 2 {
		t.Fatalf("TierFromCount = %d, want 2 (past the known x2 threshold)", got)
	}
}

func TestCalibrate(t *testing.T) {
	refs := []string{"a/a#1", "b/b#2", "c/c#3", "d/d#4"} // oldest merge first
	cases := []struct {
		name       string
		milestones []official.Milestone
		want       *Calibration
	}{
		{"agrees", []official.Milestone{{N: 2, PR: "b/b#2"}}, &Calibration{At: 2, PR: "b/b#2", Offset: 0}},
		{"GitHub counts more", []official.Milestone{{N: 4, PR: "c/c#3"}}, &Calibration{At: 4, PR: "c/c#3", Offset: 1}},
		{"GitHub counts fewer", []official.Milestone{{N: 2, PR: "c/c#3"}}, &Calibration{At: 2, PR: "c/c#3", Offset: -1}},
		{"highest public milestone wins",
			[]official.Milestone{{N: 2, PR: "b/b#2"}, {N: 16}, {N: 4, PR: "d/d#4"}},
			&Calibration{At: 4, PR: "d/d#4", Offset: 0}},
		{"private milestones are ignored", []official.Milestone{{N: 16}}, nil},
		{"unknown PR is ignored", []official.Milestone{{N: 2, PR: "x/x#9"}}, nil},
	}
	for _, c := range cases {
		got := Calibrate(refs, c.milestones)
		if (got == nil) != (c.want == nil) || (got != nil && *got != *c.want) {
			t.Errorf("%s: Calibrate = %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestOrdinal(t *testing.T) {
	cases := map[int]string{
		1: "1st", 2: "2nd", 3: "3rd", 4: "4th", 11: "11th", 12: "12th", 13: "13th",
		21: "21st", 22: "22nd", 23: "23rd", 101: "101st", 111: "111th", 112: "112th", 1024: "1024th",
	}
	for n, want := range cases {
		if got := Ordinal(n); got != want {
			t.Errorf("Ordinal(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestStand_tierAboveTheCatalogDoesNotPanic(t *testing.T) {
	// Regression: tier-count="5" from a future GitHub page indexed past the tiers.
	s := Stand([]int{2, 8, 16, 32}, 5, 40)
	if s.Tier != 4 || !s.IsMax {
		t.Fatalf("got %+v", s)
	}
}
