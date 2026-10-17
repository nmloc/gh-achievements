package main

import (
	"regexp"
	"strings"
	"testing"
)

// Every badge shows real art: the two GitHub deleted come from the pinned community copy,
// so no page renders an empty placeholder.
func TestSitePages_everyBadgeHasArt(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	for _, slug := range []string{"heart-on-your-sleeve", "open-sourcerer"} {
		page := string(pages["site/achievements/"+slug+"/index.html"])
		if !strings.Contains(page, `<img class="badge-hero" src="https://raw.githubusercontent.com/Schweinepriester/`) {
			t.Errorf("%s page lacks its art", slug)
		}
	}
	for _, path := range []string{"site/achievements/index.html", "site/index.html"} {
		if strings.Contains(string(pages[path]), "placeholder") {
			t.Errorf("%s still renders an empty placeholder", path)
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
		`name="u"`, `pattern="[A-Za-z0-9][A-Za-z0-9\-]{0,38}"`,
		`<button type="button" class="btn small" data-copy="gh extension install nmloc/gh-achievements">Copy</button>`} {
		if !strings.Contains(check, want) {
			t.Errorf("check page lacks %s", want)
		}
	}
	cb := string(pages["site/auth/callback/index.html"])
	for _, want := range []string{`<meta name="robots" content="noindex">`, `<meta name="referrer" content="no-referrer">`,
		`<script src="/callback.js" defer>`, `id="status"`, `id="result"`,
		`id="error-denied"`, `id="error-state"`, `id="error-exchange"`, `id="error-timeout"`, `id="error-github"`, `id="error-internal"`, `id="error-network"`, `<noscript>`} {
		if !strings.Contains(cb, want) {
			t.Errorf("callback page lacks %s", want)
		}
	}
	ref := strings.Index(cb, `<meta name="referrer"`)
	css := strings.Index(cb, `<link rel="stylesheet"`)
	if ref < 0 || css < 0 || ref > css {
		t.Errorf("referrer meta (at %d) must come before the stylesheet link (at %d)", ref, css)
	}
	if len(pages["site/app.js"]) < 500 || len(pages["site/callback.js"]) < 300 {
		t.Error("scripts not copied or still stubs")
	}
}

// Withdrawing consent must be possible from every page.
func TestSitePages_everyPageCanResetConsent(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for path, body := range pages {
		if !strings.HasSuffix(path, ".html") {
			continue
		}
		n++
		if !strings.Contains(string(body), "data-consent-reset") {
			t.Errorf("%s has no data-consent-reset control", path)
		}
	}
	if n == 0 {
		t.Fatal("no html pages generated")
	}
}

// GA reads document.title for every hit and /result's title names the user, so app.js must
// override page_title in both the config call and the global set call.
func TestAppJS_sanitisesPageTitleForEveryHit(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	js := string(pages["site/app.js"])
	def := regexp.MustCompile(`(?s)var sanitized = \{[^}]*page_title[^}]*page_location[^}]*\}|var sanitized = \{[^}]*page_location[^}]*page_title[^}]*\}`)
	if !def.MatchString(js) {
		t.Fatal("app.js lacks one sanitized object holding page_title and page_location")
	}
	for _, call := range []string{`gtag("config", GA, sanitized)`, `gtag("set", sanitized)`} {
		if !strings.Contains(js, call) {
			t.Errorf("app.js lacks %s", call)
		}
	}
	for _, prefix := range []string{`"/result"`, `"/auth/callback"`} {
		if !strings.Contains(js, prefix) {
			t.Errorf("app.js does not special-case %s", prefix)
		}
	}
}
