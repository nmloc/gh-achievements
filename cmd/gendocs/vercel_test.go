package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nmloc/gh-achievements/internal/web"
)

func TestVercelJSON_matchesTheFunctionsHeadersAndRoutes(t *testing.T) {
	b, err := vercelJSON()
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		OutputDirectory string
		Functions       map[string]struct{ MaxDuration int }
		Rewrites        []struct{ Source, Destination string }
		Headers         []struct {
			Source  string
			Headers []struct{ Key, Value string }
		}
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.OutputDirectory != "site" || cfg.Functions["api/**/*.go"].MaxDuration != 60 {
		t.Fatalf("config = %+v", cfg)
	}
	routes := map[string]string{}
	for _, r := range cfg.Rewrites {
		routes[r.Source] = r.Destination
	}
	if routes["/result"] != "/api/result" || routes["/auth/login"] != "/api/login" {
		t.Fatalf("rewrites = %v", routes)
	}
	got := map[string]string{}
	for _, h := range cfg.Headers[0].Headers {
		got[h.Key] = h.Value
	}
	for _, h := range web.SecurityHeaders() {
		if got[h.Key] != h.Value {
			t.Errorf("vercel.json %s differs from web.SecurityHeaders", h.Key)
		}
	}
}

// GitHub redirects to /auth/callback without a trailing slash, so both spellings need
// the no-referrer rule, and both must come after the global rule so their value wins.
func TestVercelJSON_callbackRulesCoverBothSpellingsAndComeAfterTheGlobalRule(t *testing.T) {
	b, err := vercelJSON()
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Headers []struct {
			Source  string
			Headers []struct{ Key, Value string }
		}
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	index := map[string]int{}
	for i, h := range cfg.Headers {
		index[h.Source] = i
	}
	global, ok := index["/(.*)"]
	if !ok {
		t.Fatal("no global header rule")
	}
	want := map[string]string{"Referrer-Policy": "no-referrer", "Cache-Control": "no-store", "X-Robots-Tag": "noindex"}
	for _, src := range []string{"/auth/callback", "/auth/callback/(.*)"} {
		i, ok := index[src]
		if !ok {
			t.Fatalf("no header rule for %s", src)
		}
		if i <= global {
			t.Errorf("%s (rule %d) must come after the global rule (%d)", src, i, global)
		}
		got := map[string]string{}
		for _, h := range cfg.Headers[i].Headers {
			got[h.Key] = h.Value
		}
		for k, v := range want {
			if got[k] != v {
				t.Errorf("%s: %s = %q, want %q", src, k, got[k], v)
			}
		}
	}
}

func TestRedirectPages_pointEveryOldURLAtVercel(t *testing.T) {
	pages := redirectPages()
	page := string(pages["pages/achievements/pull-shark/index.html"])
	want := "https://gh-achievements.vercel.app/achievements/pull-shark/"
	if !strings.Contains(page, `<link rel="canonical" href="`+want+`">`) || !strings.Contains(page, `content="0; url=`+want+`"`) {
		t.Fatalf("redirect page = %s", page)
	}
	if _, ok := pages["pages/index.html"]; !ok {
		t.Fatal("no root redirect")
	}
}
