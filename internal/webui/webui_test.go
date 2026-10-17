package webui

import (
	"strings"
	"testing"
)

func TestRender_usesTheVercelDomainAndRootPaths(t *testing.T) {
	b, err := Render(Page{Path: "404.html", URL: SiteURL + "404.html", Title: "T", Description: "D",
		Heading: "H", Content: "404.html", NoIndex: true})
	if err != nil {
		t.Fatal(err)
	}
	html := string(b)
	for _, want := range []string{
		`<link rel="stylesheet" href="/style.css">`,
		`<script src="/theme.js"></script>`,
		`<button type="button" class="theme-btn" data-theme-toggle hidden>`,
		`<script src="/app.js" defer data-ga="G-3EEKKGL110"></script>`,
		`<meta name="robots" content="noindex">`,
		`id="consent" hidden`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %s", want)
		}
	}
	if strings.Contains(html, "nmloc.github.io") || strings.Contains(html, "/gh-achievements/") {
		t.Error("old GitHub Pages paths remain")
	}
}

func TestRender_noInlineScripts(t *testing.T) {
	// The CSP forbids inline scripts; only JSON-LD data blocks are allowed inline.
	b, _ := Render(Page{Path: "404.html", URL: SiteURL, Title: "T", Description: "D", Heading: "H", Content: "404.html"})
	for _, chunk := range strings.Split(string(b), "<script")[1:] {
		head := chunk[:strings.Index(chunk, ">")]
		if !strings.Contains(head, "src=") && !strings.Contains(head, `type="application/ld+json"`) {
			t.Fatalf("inline script: <script%s>", head)
		}
	}
}

func TestRel(t *testing.T) {
	if got := Rel(SiteURL + "achievements/"); got != "/achievements/" {
		t.Fatalf("Rel = %q", got)
	}
}

func TestInitials(t *testing.T) {
	for name, want := range map[string]string{
		"Heart On Your Sleeve": "HO", "YOLO": "Y", "open sourcerer": "OS", "": "", "Édith Piaf": "ÉP",
	} {
		if got := Initials(name); got != want {
			t.Errorf("Initials(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestStyle_openCardAndLongLinkRules(t *testing.T) {
	// Layout rules checked in a real browser (open card on desktop, 375px mobile overflow).
	// Pinned here so a stylesheet edit cannot drop them silently.
	css, err := Static("style.css")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		".badge-card[open] summary .badge-img { grid-row: 1 / span 3; }",
		".badge-card[open] summary > :not(.badge-img) { grid-column: 2; }",
		".badge-card[open] summary .badge-img { grid-row: auto; }",
		".badge-card[open] summary > :not(.badge-img) { grid-column: auto; }",
		".badge-detail { padding: 0 16px 16px; overflow-wrap: anywhere; }",
		"grid-template-columns: repeat(2, minmax(0, 1fr));",
	} {
		if !strings.Contains(string(css), want) {
			t.Errorf("style.css lacks %q", want)
		}
	}
}

// The dark tokens are written twice: for an explicit choice and for the system setting.
// They must never drift apart.
func TestStyle_darkTokensMatchInBothBlocks(t *testing.T) {
	css, err := Static("style.css")
	if err != nil {
		t.Fatal(err)
	}
	block := func(selector string) string {
		s := string(css)
		i := strings.Index(s, selector+" {")
		if i < 0 {
			t.Fatalf("style.css has no %s block", selector)
		}
		rest := s[i+len(selector)+2:]
		return strings.Join(strings.Fields(rest[:strings.Index(rest, "}")]), " ")
	}
	explicit, system := block(`:root[data-theme="dark"]`), block(`:root:not([data-theme="light"])`)
	if explicit != system {
		t.Fatalf("dark tokens differ:\n explicit: %s\n system:   %s", explicit, system)
	}
	for _, want := range []string{"--bg: #0d1c26", "--gold: #e2b53a", "color-scheme: dark"} {
		if !strings.Contains(explicit, want) {
			t.Errorf("dark tokens lack %s", want)
		}
	}
	for _, want := range []string{"@media (prefers-color-scheme: dark)", "font-display: swap",
		`url("/fonts/bricolage-grotesque.woff2")`, "@media (prefers-reduced-motion: no-preference)"} {
		if !strings.Contains(string(css), want) {
			t.Errorf("style.css lacks %s", want)
		}
	}
}

func TestThemeJS_appliesOnlyAStoredLightOrDark(t *testing.T) {
	js, err := Static("theme.js")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`localStorage.getItem("gha:theme")`, `t === "light" || t === "dark"`,
		`setAttribute("data-theme", t)`, `classList.add("js")`} {
		if !strings.Contains(string(js), want) {
			t.Errorf("theme.js lacks %s", want)
		}
	}
}
