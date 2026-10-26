package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/nmloc/gh-achievements/internal/webui"
)

// Every page in the sitemap has a Markdown twin, so Accept: text/markdown never 404s on a real page.
func TestMarkdownPages_coverTheSitemap(t *testing.T) {
	md := markdownPages()
	for _, u := range sitemapURLs() {
		p := strings.TrimPrefix(u, webui.SiteURL)
		body, ok := md[p]
		if !ok {
			t.Errorf("%s has no Markdown twin", u)
			continue
		}
		if !strings.HasPrefix(body, "# ") || !strings.Contains(body, webui.SiteURL+"llms.txt") {
			t.Errorf("%s: twin lacks a heading or the llms.txt link", u)
		}
	}
}

var tags = regexp.MustCompile(`<[^>]+>`)

// About, contact and privacy are real pages: each says enough to be worth reading
// (over 500 characters of its own text), in the sitemap, linked from every footer.
func TestTrustPages_haveRealContentAndAreLinked(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	sm := sitemap()
	for _, tp := range trustPages {
		html := string(pages["site/"+tp.Slug+"/index.html"])
		main := html[strings.Index(html, "<main>"):strings.Index(html, "</main>")]
		if text := strings.Join(strings.Fields(tags.ReplaceAllString(main, " ")), " "); len(text) < 500 {
			t.Errorf("%s: %d characters of text, want at least 500", tp.Slug, len(text))
		}
		if !strings.Contains(html, "<h1>"+tp.Heading+"</h1>") || !strings.Contains(html, `<link rel="canonical" href="`+webui.SiteURL+tp.Slug+`/">`) {
			t.Errorf("%s: no heading or canonical link", tp.Slug)
		}
		if !strings.Contains(sm, "<loc>"+webui.SiteURL+tp.Slug+"/</loc>") {
			t.Errorf("%s not in the sitemap", tp.Slug)
		}
		if !strings.Contains(string(pages["site/index.html"]), `<a href="/`+tp.Slug+`/">`) {
			t.Errorf("footer does not link %s", tp.Slug)
		}
	}
}

// llms.txt tells an agent which jobs gh-achievements is for, and how to call it.
func TestLlmsTxt_saysWhenAndHowToUseIt(t *testing.T) {
	txt := llmsTxt()
	for _, want := range []string{"## When to use gh-achievements", "## How an agent should call it",
		"gh achievements <username> --json", webui.SiteURL + "achievements.json", "Accept: text/markdown",
		webui.SiteURL + "about/", webui.SiteURL + "contact/", webui.SiteURL + "privacy/"} {
		if !strings.Contains(txt, want) {
			t.Errorf("llms.txt lacks %q", want)
		}
	}
	// llmstxt.org: an H1, then a blockquote summary, before any H2.
	if !strings.HasPrefix(txt, "# gh-achievements\n\n> ") {
		t.Error("llms.txt must open with the H1 and the summary blockquote")
	}
}

func TestMarkdownNotFound_pointsSomewhereUseful(t *testing.T) {
	body := markdownNotFound()
	if len(body) < 20 || !strings.Contains(body, webui.SiteURL+"sitemap.xml") || !strings.Contains(body, webui.SiteURL+"llms.txt") {
		t.Errorf("404 body = %q", body)
	}
}
