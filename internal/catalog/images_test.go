package catalog

import (
	"strings"
	"testing"
)

func TestImage_everyBadgeWithArtHasEveryTier(t *testing.T) {
	for _, a := range Achievements {
		if a.Kind == Disabled {
			if Image(a.Slug, 1) != "" {
				t.Errorf("%s: disabled badges never shipped art", a.Slug)
			}
			continue
		}
		for tier := 0; tier <= len(a.Tiers); tier++ {
			u := Image(a.Slug, tier)
			if !strings.HasPrefix(u, "https://github.githubassets.com/assets/"+a.Slug+"-") {
				t.Errorf("%s tier %d: %q", a.Slug, tier, u)
			}
		}
	}
}

func TestImage_tierPicksTheMatchingArt(t *testing.T) {
	cases := map[int]string{0: "-default-", 1: "-default-", 2: "-bronze-", 3: "-silver-", 4: "-gold-", 9: "-gold-"}
	for tier, want := range cases {
		if u := Image("pull-shark", tier); !strings.Contains(u, want) {
			t.Errorf("tier %d → %s, want %s", tier, u, want)
		}
	}
	if Image("quickdraw", 3) != Image("quickdraw", 1) {
		t.Error("single-tier badges have one image")
	}
	if Image("no-such-badge", 1) != "" {
		t.Error("unknown badge got an image")
	}
}
