package official

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestParseEarned_listsEachBadgeOnce(t *testing.T) {
	got := ParseEarned(loadFixture(t, "brannon", "tab"))
	want := []string{"pair-extraordinaire", "starstruck", "proxima-staffshipper", "proxima-pioneer",
		"quickdraw", "yolo", "arctic-code-vault-contributor", "pull-shark"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseEarned = %v, want %v", got, want)
	}
}

func TestParseEarned_findsEveryKindOfBadge(t *testing.T) {
	cases := map[string]string{
		"torvalds":     "mars-2020-contributor",
		"brannon":      "proxima-pioneer",
		"timrogers":    "proxima-staffuser",
		"dylanatsmith": "finishing-touch",
		"ljharb":       "galaxy-brain",
	}
	for user, slug := range cases {
		if got := ParseEarned(loadFixture(t, user, "tab")); !contains(got, slug) {
			t.Errorf("%s: %v has no %s", user, got, slug)
		}
	}
}

func TestParseDetail_pullSharkMilestones(t *testing.T) {
	d := ParseDetail("pull-shark", loadFixture(t, "brannon", "detail-pull-shark"))
	if d.Name != "Pull Shark" || d.Tier != 3 || d.Unlocked != "2012-04-26" {
		t.Fatalf("got %+v", d)
	}
	want := []Milestone{
		{N: 2, PR: "DapperLib/Dapper#32"},
		{N: 16},  // private: "inaccessible"
		{N: 128}, // private
	}
	if len(d.Milestones) != len(want) {
		t.Fatalf("milestones = %+v", d.Milestones)
	}
	for i, m := range d.Milestones {
		if m.N != want[i].N || m.PR != want[i].PR || m.Repo != "" {
			t.Errorf("milestone %d = %+v, want %+v", i, m, want[i])
		}
	}
}

func TestParseDetail_singleTierBadgeIsTierOne(t *testing.T) {
	// Regression: the MVP read tiers off the achievements tab and gave a
	// single-tier badge its neighbour's "x3" label.
	for _, c := range [][2]string{{"brannon", "quickdraw"}, {"brannon", "yolo"}, {"Schweinepriester", "public-sponsor"}} {
		if d := ParseDetail(c[1], loadFixture(t, c[0], "detail-"+c[1])); d.Tier != 1 {
			t.Errorf("%s/%s tier = %d, want 1", c[0], c[1], d.Tier)
		}
	}
}

func TestParseDetail_starstruckNamesRepositoriesNotOrdinals(t *testing.T) {
	// "⭐️ 16 stars" carries no ordinal; each row names a repository, and the
	// repositories can change between tiers and belong to organisations.
	d := ParseDetail("starstruck", loadFixture(t, "ljharb", "detail-starstruck"))
	var repos []string
	for _, m := range d.Milestones {
		if m.N != 0 || m.PR != "" {
			t.Errorf("row %+v should be a repository row", m)
		}
		repos = append(repos, m.Repo)
	}
	want := []string{"es-shims/es6-shim", "ljharb/qs", "enzymejs/enzyme", "enzymejs/enzyme"}
	if !reflect.DeepEqual(repos, want) {
		t.Fatalf("repos = %v, want %v", repos, want)
	}
}

func TestParseDetail_finishingTouchHasFourTiers(t *testing.T) {
	d := ParseDetail("finishing-touch", loadFixture(t, "dylanatsmith", "detail-finishing-touch"))
	var ns []int
	for _, m := range d.Milestones {
		ns = append(ns, m.N)
	}
	if d.Tier != 4 || !reflect.DeepEqual(ns, []int{1, 4, 16, 32}) {
		t.Fatalf("tier %d, milestones %v", d.Tier, ns)
	}
}

func TestParseDetail_pairMilestonesNameOtherPeoplesPRs(t *testing.T) {
	d := ParseDetail("pair-extraordinaire", loadFixture(t, "torvalds", "detail-pair-extraordinaire"))
	if d.Tier != 4 || d.Milestones[0].PR != "samodostal/.dotfiles#4" || d.Milestones[0].Repo != "" {
		t.Fatalf("got %+v", d)
	}
}

func TestParseHighlights(t *testing.T) {
	cases := map[string][]string{
		"Schweinepriester": {"Developer Program Member"},
		"kentcdodds":       {"GitHub Star", "Pro"},
		"ljharb":           {"Developer Program Member", "Pro", "33 security advisory credits"},
		"torvalds":         nil,
	}
	for user, want := range cases {
		if got := ParseHighlights(loadFixture(t, user, "tab")); !reflect.DeepEqual(got, want) {
			t.Errorf("%s highlights = %q, want %q", user, got, want)
		}
	}
}

func TestHasHighlight_matchesWholeLabelOnly(t *testing.T) {
	// Regression: a substring match let "Pro" match "Developer Program Member".
	if HasHighlight([]string{"Developer Program Member"}, "Pro") {
		t.Error(`"Pro" matched "Developer Program Member"`)
	}
	if !HasHighlight([]string{"Pro"}, "Pro") {
		t.Error(`"Pro" did not match "Pro"`)
	}
	// The advisory highlight carries a count and a plural.
	if !HasHighlight([]string{"33 security advisory credits"}, "Security advisory credit") {
		t.Error("counted, plural advisory highlight not matched")
	}
}

// fakeGitHub serves fixtures the way github.com does.
func fakeGitHub(t *testing.T, routes map[string]string, requireXHR bool) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		if requireXHR && strings.HasSuffix(r.URL.Path, "/detail") && r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
			w.WriteHeader(http.StatusNotAcceptable) // what github.com answers
			return
		}
		body, ok := routes[path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return &Client{HTTP: srv.Client(), Base: srv.URL}
}

func TestProfile_hiddenAchievementsAreNotAMissingUser(t *testing.T) {
	// Regression: the tab 404s for a hidden profile too; the MVP called that "not found".
	c := fakeGitHub(t, map[string]string{"/sindresorhus": loadFixture(t, "sindresorhus", "main")}, false)
	p, err := c.Profile("sindresorhus")
	if err != nil || !p.Hidden || p.Earned != nil {
		t.Fatalf("Profile = %+v, %v; want hidden", p, err)
	}
}

func TestProfile_missingUser(t *testing.T) {
	c := fakeGitHub(t, map[string]string{}, false)
	if _, err := c.Profile("nobody"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("err = %v, want ErrUserNotFound", err)
	}
}

func TestDetail_sendsXHRHeaderAndReturnsNilWhenNotEarned(t *testing.T) {
	c := fakeGitHub(t, map[string]string{
		"/users/brannon/achievements/yolo/detail": loadFixture(t, "brannon", "detail-yolo"),
	}, true)
	d, err := c.Detail("brannon", "yolo")
	if err != nil || d == nil || d.Tier != 1 {
		t.Fatalf("Detail(yolo) = %+v, %v", d, err)
	}
	d, err = c.Detail("brannon", "galaxy-brain")
	if err != nil || d != nil {
		t.Fatalf("Detail(unearned) = %+v, %v; want nil, nil", d, err)
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestParseDetail_ordinalIsTheLastOneInTheRow(t *testing.T) {
	// Regression: a repository named like "1st-app" made the 16th milestone read as 1st.
	frag := `<h3>Pull Shark</h3><div class="TimelineItem achievement-history-tier"><div class="TimelineItem-body x">` +
		`<a class="Link" href="/me/1st-app/pull/5">me/1st-app#5</a> · <span>16th pull request merged</span></div></div>`
	d := ParseDetail("pull-shark", frag)
	if len(d.Milestones) != 1 || d.Milestones[0].N != 16 || d.Milestones[0].PR != "me/1st-app#5" {
		t.Fatalf("got %+v", d.Milestones)
	}
}

func TestIsLogin(t *testing.T) {
	// Regression: "torvalds?x" became /torvalds?x?tab=achievements and printed a wrong
	// report for torvalds under another name.
	valid := []string{"torvalds", "nmloc", "a", "Schweinepriester", "dependabot-preview", "a-b-c", "x1234567890123456789012345678901234567"}
	invalid := []string{"", "torvalds?x", "../etc", "a b", "-lead", "trail-", "dou--ble", "a/b", "x12345678901234567890123456789012345678901"}
	for _, s := range valid {
		if !IsLogin(s) {
			t.Errorf("IsLogin(%q) = false", s)
		}
	}
	for _, s := range invalid {
		if IsLogin(s) {
			t.Errorf("IsLogin(%q) = true", s)
		}
	}
}
