package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serveMarkdown(method, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, "/api/markdown?path="+path, nil)
	r.Header.Set("Accept", "text/markdown")
	Markdown(w, r)
	return w
}

func TestMarkdown_servesEveryPageTwin(t *testing.T) {
	for path, heading := range map[string]string{
		"":                          "# gh-achievements",
		"achievements":              "# All GitHub achievements",
		"achievements/":             "# All GitHub achievements",
		"achievements/pull-shark/":  "# Pull Shark (GitHub achievement)",
		"/achievements/pull-shark/": "# Pull Shark (GitHub achievement)",
		"methodology/":              "# How accurate is gh-achievements?",
		"about/":                    "# About gh-achievements",
		"contact":                   "# Contact",
		"privacy/":                  "# Privacy",
	} {
		w := serveMarkdown("GET", path)
		if w.Code != 200 || !strings.HasPrefix(w.Body.String(), heading) {
			t.Errorf("%q: status %d, body starts %.40q, want %q", path, w.Code, w.Body.String(), heading)
		}
		if ct := w.Header().Get("Content-Type"); ct != "text/markdown; charset=utf-8" {
			t.Errorf("%q: Content-Type = %q", path, ct)
		}
		if v := w.Header().Get("Vary"); v != "Accept" {
			t.Errorf("%q: Vary = %q", path, v)
		}
		if w.Header().Get("Content-Security-Policy") == "" {
			t.Errorf("%q: no security headers", path)
		}
	}
}

// An agent that asks for Markdown and hits a dead link gets a 404 it can read: a Markdown
// body that says what happened and where to look instead.
func TestMarkdown_unknownPathIsA404WithAMarkdownBody(t *testing.T) {
	for _, path := range []string{"__ora-404-probe-79rn61ru", "achievements/nope/", "../web/markdown.go", "achievements/../about"} {
		w := serveMarkdown("GET", path)
		body := w.Body.String()
		if w.Code != 404 || w.Header().Get("Content-Type") != "text/markdown; charset=utf-8" || w.Header().Get("Vary") != "Accept" {
			t.Errorf("%q: status %d, headers %v", path, w.Code, w.Header())
		}
		if !strings.HasPrefix(body, "# Page not found") || !strings.Contains(body, "https://gh-achievements.vercel.app/llms.txt") ||
			!strings.Contains(body, "https://gh-achievements.vercel.app/sitemap.xml") {
			t.Errorf("%q: body %q", path, body)
		}
	}
}

func TestMarkdown_headHasNoBodyAndPostIsRefused(t *testing.T) {
	if w := serveMarkdown("HEAD", ""); w.Code != 200 || w.Body.Len() != 0 || w.Header().Get("Content-Type") != "text/markdown; charset=utf-8" {
		t.Errorf("HEAD: status %d, %d bytes", w.Code, w.Body.Len())
	}
	if w := serveMarkdown("POST", ""); w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != "GET, HEAD" {
		t.Errorf("POST: status %d", w.Code)
	}
}

// WantsMarkdown is MarkdownRoute in Go: pages yes, files and functions no, and only when asked.
func TestWantsMarkdown(t *testing.T) {
	for _, tc := range []struct {
		path, accept string
		want         bool
	}{
		{"/", "text/markdown", true},
		{"/achievements/pull-shark/", "text/markdown, text/html;q=0.9", true},
		{"/__ora-404-probe-79rn61ru", "text/markdown", true},
		{"/results", "text/markdown", true},
		{"/", "text/html,application/xhtml+xml,*/*;q=0.8", false},
		{"/", "", false},
		{"/llms.txt", "text/markdown", false},
		{"/achievements.json", "text/markdown", false},
		{"/result", "text/markdown", false},
		{"/result/", "text/markdown", false},
		{"/api/markdown", "text/markdown", false},
		{"/auth/callback/", "text/markdown", false},
	} {
		r := httptest.NewRequest("GET", tc.path, nil)
		if tc.accept != "" {
			r.Header.Set("Accept", tc.accept)
		}
		if got := WantsMarkdown(r); got != tc.want {
			t.Errorf("WantsMarkdown(%s, %q) = %v, want %v", tc.path, tc.accept, got, tc.want)
		}
	}
}
