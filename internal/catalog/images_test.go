package catalog

import (
	"strings"
	"testing"
)

func TestImage_everyBadgeWithArtHasEveryTier(t *testing.T) {
	for _, a := range Achievements {
		// GitHub deleted the art of the badges it switched off; a community copy,
		// pinned to one commit, stands in for it.
		prefix := "https://github.githubassets.com/assets/" + a.Slug + "-"
		if a.Kind == Disabled {
			prefix = communityArt
		}
		for tier := 0; tier <= len(a.Tiers); tier++ {
			u := Image(a.Slug, tier)
			if !strings.HasPrefix(u, prefix) || !strings.Contains(u, a.Slug) {
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

func TestImage_communityArtIsPinnedToOneCommit(t *testing.T) {
	// A branch name would let the art change or vanish under us; a commit cannot.
	if !strings.Contains(communityArt, "/549ba7291ccb68efcff5daf8af8c2c0ac6817b6f/") {
		t.Fatalf("community art is not pinned to a commit: %s", communityArt)
	}
	if got := Image("heart-on-your-sleeve", 3); !strings.HasSuffix(got, "/tiers/heart-on-your-sleeve-silver.png") {
		t.Fatalf("heart x3 → %s", got)
	}
}
