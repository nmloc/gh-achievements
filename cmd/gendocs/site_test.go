package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/nmloc/gh-achievements/internal/catalog"
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
		if strings.Contains(string(pages[path]), ` placeholder"`) {
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

// theme.js must run before the stylesheet so the first paint is in the chosen theme,
// and the CSP forbids inline scripts on every page (JSON-LD data blocks excepted).
func TestSitePages_themeLoadsFirstAndNoInlineScript(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	for path, body := range pages {
		if !strings.HasSuffix(path, ".html") {
			continue
		}
		html := string(body)
		theme := strings.Index(html, `<script src="/theme.js"></script>`)
		css := strings.Index(html, `<link rel="stylesheet" href="/style.css">`)
		if theme < 0 || css < 0 || theme > css {
			t.Errorf("%s: theme.js (at %d) must load before the stylesheet (at %d)", path, theme, css)
		}
		for _, chunk := range strings.Split(html, "<script")[1:] {
			head := chunk[:strings.Index(chunk, ">")]
			if !strings.Contains(head, "src=") && !strings.Contains(head, `type="application/ld+json"`) {
				t.Errorf("%s: inline script <script%s>", path, head)
			}
		}
	}
	if len(pages["site/fonts/bricolage-grotesque.woff2"]) < 10000 || len(pages["site/theme.js"]) == 0 ||
		!strings.Contains(string(pages["site/fonts/OFL.txt"]), "SIL OPEN FONT LICENSE") {
		t.Error("theme.js, the font or its licence is not copied into site/")
	}
}

func TestSitePages_homeIsATrophyCase(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	home := string(pages["site/index.html"])
	for _, want := range []string{`<h1>Your GitHub trophy case</h1>`, `<form action="/result" method="get" id="check-form">`,
		`<a href="/auth/login" data-signin>Sign in with GitHub</a> to include private work.`,
		`<ul class="vitrine shelf home-shelf"`, `data-star="home"`, `data-copy="gh extension install nmloc/gh-achievements"`,
		`"@type":"FAQPage"`, `"@type":"SoftwareApplication"`} {
		if !strings.Contains(home, want) {
			t.Errorf("home lacks %s", want)
		}
	}
	for _, a := range catalog.Achievements {
		link := `<a class="medal" href="achievements/` + a.Slug + `/">`
		if got := strings.Contains(home, link); got != (a.Kind == catalog.Earnable) {
			t.Errorf("%s on the home shelf = %v, want %v", a.Slug, got, a.Kind == catalog.Earnable)
		}
	}
	if strings.Index(home, `data-star="home"`) < strings.Index(home, `<dl class="faq">`) {
		t.Error("the home star banner belongs under the FAQ")
	}
	if strings.Contains(home, "demo.svg") {
		t.Error("the home page no longer shows the terminal demo")
	}
}

func TestSitePages_everyPageHasOneH1(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	for path, body := range pages {
		if strings.HasSuffix(path, ".html") {
			if n := strings.Count(string(body), "<h1"); n != 1 {
				t.Errorf("%s has %d <h1>", path, n)
			}
		}
	}
}

func TestSitePages_listGroupsBadgesByKind(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	list := string(pages["site/achievements/index.html"])
	last := -1
	for _, h := range []string{"<h2>Earnable</h2>", "<h2>Switched off by GitHub</h2>", "<h2>Retired</h2>", "<h2>GitHub staff only</h2>"} {
		i := strings.Index(list, h)
		if i < 0 {
			t.Errorf("%s missing", h)
			continue
		}
		if i < last {
			t.Errorf("%s out of order", h)
		}
		last = i
	}
	for _, a := range catalog.Achievements {
		if !strings.Contains(list, `<h3><a href="`+a.Slug+`/">`+a.Name+`</a></h3>`) {
			t.Errorf("list lacks a medal for %s", a.Slug)
		}
	}
	if !strings.Contains(list, `<ul class="medal-grid">`) || strings.Contains(list, `<table class="badges">`) {
		t.Error("the list is a medal grid now, not a table")
	}
}

func TestSitePages_badgePageShowsEveryTierMedal(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	ps := string(pages["site/achievements/pull-shark/index.html"])
	for i := 1; i <= 4; i++ {
		if !strings.Contains(ps, `<img class="medal-img" src="`+catalog.Image("pull-shark", i)+`"`) {
			t.Errorf("pull shark page lacks the x%d medal", i)
		}
	}
	for _, want := range []string{`<div class="vitrine badge-top">`, `<ul class="tier-strip" aria-label="Tiers">`, "x4 gold"} {
		if !strings.Contains(ps, want) {
			t.Errorf("pull shark page lacks %s", want)
		}
	}
	if yolo := string(pages["site/achievements/yolo/index.html"]); strings.Contains(yolo, "tier-strip") {
		t.Error("a single-tier badge has no tier strip")
	}
	if m := string(pages["site/methodology/index.html"]); !strings.Contains(m, `<div class="prose">`) {
		t.Error("methodology is not in a readable column")
	}
	if cb := string(pages["site/auth/callback/index.html"]); !strings.Contains(cb, `<div id="status" class="notice checking" role="status" aria-live="polite">`) {
		t.Error("callback status card lost its id, role or style")
	}
}
