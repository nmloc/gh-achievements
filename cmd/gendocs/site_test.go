package main

import (
	"strings"
	"testing"
)

// Badges with no GitHub artwork (disabled ones) get a neutral placeholder, not nothing,
// so the badge page has a hero and the table rows stay aligned.
func TestSitePlaceholders_badgesWithoutArtGetAPlaceholder(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}

	heart := string(pages["site/achievements/heart-on-your-sleeve/index.html"])
	if !strings.Contains(heart, `class="badge-hero placeholder"`) || !strings.Contains(heart, ">HO</span>") {
		t.Errorf("heart-on-your-sleeve page has no HO placeholder hero")
	}
	open := string(pages["site/achievements/open-sourcerer/index.html"])
	if !strings.Contains(open, `class="badge-hero placeholder"`) || !strings.Contains(open, ">OS</span>") {
		t.Errorf("open-sourcerer page has no OS placeholder hero")
	}

	for _, path := range []string{"site/achievements/index.html", "site/index.html"} {
		page := string(pages[path])
		if n := strings.Count(page, `class="badge-icon placeholder"`); n != 2 {
			t.Errorf("%s has %d placeholder icons, want 2 (heart-on-your-sleeve, open-sourcerer)", path, n)
		}
	}
}

func TestSitePages_checkAndCallback(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	check := string(pages["site/check/index.html"])
	for _, want := range []string{`href="/auth/login" data-signin`, `<form action="/result" method="get" id="check-form">`,
		`name="u"`, `pattern="[A-Za-z0-9][A-Za-z0-9-]{0,38}"`} {
		if !strings.Contains(check, want) {
			t.Errorf("check page lacks %s", want)
		}
	}
	cb := string(pages["site/auth/callback/index.html"])
	for _, want := range []string{`<meta name="robots" content="noindex">`, `<meta name="referrer" content="no-referrer">`,
		`<script src="/callback.js" defer>`, `id="status"`, `id="result"`,
		`id="error-denied"`, `id="error-state"`, `id="error-exchange"`, `id="error-timeout"`, `id="error-github"`, `id="error-internal"`, `id="error-network"`} {
		if !strings.Contains(cb, want) {
			t.Errorf("callback page lacks %s", want)
		}
	}
	if len(pages["site/app.js"]) < 500 || len(pages["site/callback.js"]) < 300 {
		t.Error("scripts not copied or still stubs")
	}
}
