// Package fixtures serves the real github.com pages saved under
// internal/official/testdata, for tests and for the README demo.
package fixtures

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/nmloc/gh-achievements/internal/counters"
	"github.com/nmloc/gh-achievements/internal/official"
)

// Dir is the fixture directory, found relative to this file so it works from any package.
func Dir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "official", "testdata")
}

// Read returns a fixture page, or false when there is none.
func Read(user, name string) (string, bool, error) {
	f, err := os.Open(filepath.Join(Dir(), user, name+".html.gz"))
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return "", false, fmt.Errorf("%s/%s: %w", user, name, err)
	}
	b, err := io.ReadAll(zr)
	return string(b), true, err
}

// Official answers like github.com, from fixtures.
type Official struct{}

func (Official) Profile(user string) (official.Profile, error) {
	tab, ok, err := Read(user, "tab")
	if err != nil {
		return official.Profile{}, err
	}
	if ok {
		return official.Profile{Earned: official.ParseEarned(tab), Highlights: official.ParseHighlights(tab)}, nil
	}
	main, ok, err := Read(user, "main")
	if err != nil {
		return official.Profile{}, err
	}
	if ok {
		return official.Profile{Hidden: true, Highlights: official.ParseHighlights(main)}, nil
	}
	return official.Profile{}, official.ErrUserNotFound
}

func (Official) Detail(user, slug string) (*official.Detail, error) {
	frag, ok, err := Read(user, "detail-"+slug)
	if err != nil || !ok {
		return nil, err
	}
	d := official.ParseDetail(slug, frag)
	return &d, nil
}

// Counts is a counters.Source with fixed answers.
type Counts struct {
	Authed     bool
	PRs        counters.MergedPRs
	Owned      int
	OwnedRepo  string
	RepoStar   map[string]int
	RepoErr    map[string]error // RepoStars fails for these
	FactsValue *counters.Facts
}

func (c *Counts) IsAuthenticated() bool                           { return c.Authed }
func (c *Counts) MergedPRs(string) (counters.MergedPRs, error)    { return c.PRs, nil }
func (c *Counts) MaxOwnedStars(string) (int, string, error)       { return c.Owned, c.OwnedRepo, nil }
func (c *Counts) RepoStars(repo string) (int, error)              { return c.RepoStar[repo], c.RepoErr[repo] }
func (c *Counts) Facts(string, []string) (*counters.Facts, error) { return c.FactsValue, nil }

// TorvaldsDemo is the README demo: @torvalds's real official pages, with counts fixed
// at the values the live CLI returned on the catalog's verification date.
func TorvaldsDemo() *Counts {
	refs := make([]string, 72)
	for i := range refs {
		refs[i] = fmt.Sprintf("demo/repo#%d", i+1)
	}
	refs[1] = "subsurface/libdc#1"
	refs[15] = "subsurface/subsurface#1506"
	return &Counts{
		Authed: true,
		PRs:    counters.MergedPRs{Total: 72, Refs: refs},
		Owned:  252192, OwnedRepo: "torvalds/linux",
		RepoStar:   map[string]int{"torvalds/linux": 252192},
		FactsValue: &counters.Facts{Flags: map[string]bool{}},
	}
}
