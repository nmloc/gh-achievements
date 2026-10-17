//go:build canary

// The canary runs the parsers against live github.com pages (weekly in CI). Fixtures
// prove the parsers on yesterday's HTML; this proves them on today's, and it notices
// badges GitHub adds before anyone files an issue.
//
//	go test -tags canary ./internal/official -run Canary -v
package official

import (
	"net/http"
	"os"
	"testing"

	"github.com/nmloc/gh-achievements/internal/catalog"
)

func live(t *testing.T) *Client {
	t.Helper()
	if os.Getenv("CANARY_SIMULATE_FAILURE") == "true" {
		t.Fatal("simulated canary failure (workflow_dispatch input)")
	}
	return NewClient()
}

func TestCanary_officialPagesStillParse(t *testing.T) {
	c := live(t)

	p, err := c.Profile("torvalds")
	if err != nil {
		t.Fatal(err)
	}
	for _, slug := range []string{"pull-shark", "starstruck", "mars-2020-contributor"} {
		if !contains(p.Earned, slug) {
			t.Errorf("torvalds: %s missing from %v: the achievements tab changed", slug, p.Earned)
		}
	}
	ss, err := c.Detail("torvalds", "starstruck")
	if err != nil || ss == nil || ss.Tier != 4 || len(ss.Milestones) != 4 || ss.Milestones[3].Repo != "torvalds/linux" {
		t.Errorf("torvalds starstruck detail no longer parses: %+v, %v", ss, err)
	}
	ps, err := c.Detail("torvalds", "pull-shark")
	if err != nil || ps == nil || ps.Unlocked == "" || len(ps.Milestones) < 2 || ps.Milestones[1].N != 16 || ps.Milestones[1].PR == "" {
		t.Errorf("torvalds pull-shark detail no longer parses: %+v, %v", ps, err)
	}

	lj, err := c.Profile("ljharb")
	if err != nil {
		t.Fatal(err)
	}
	if !HasHighlight(lj.Highlights, "Pro") || !HasHighlight(lj.Highlights, "Security advisory credit") {
		t.Errorf("ljharb highlights no longer parse: %q", lj.Highlights)
	}

	hidden, err := c.Profile("sindresorhus")
	if err != nil || !hidden.Hidden {
		t.Errorf("hidden profile detection broke: %+v, %v", hidden, err)
	}
}

func TestCanary_noUnknownBadges(t *testing.T) {
	c := live(t)
	known := map[string]bool{}
	for _, a := range catalog.Achievements {
		known[a.Slug] = true
	}
	// Users who hold many badges; a new slug here means GitHub added an achievement.
	for _, user := range []string{"torvalds", "ljharb", "Schweinepriester", "brannon", "dylanatsmith", "timrogers", "kentcdodds"} {
		p, err := c.Profile(user)
		if err != nil {
			t.Fatalf("%s: %v", user, err)
		}
		for _, slug := range p.Earned {
			if !known[slug] {
				t.Errorf("new achievement on @%s: %q is not in internal/catalog", user, slug)
			}
		}
	}
}

func TestCanary_badgeImagesStillLoad(t *testing.T) {
	live(t)
	for slug, urls := range catalog.Images {
		for _, u := range urls {
			resp, err := http.Head(u)
			if err != nil {
				t.Errorf("%s: %v", slug, err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("%s: %s answered %d: GitHub renamed the artwork", slug, u, resp.StatusCode)
			}
		}
	}
}
