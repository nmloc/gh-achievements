// web/markdown.go
package web

import (
	"embed"
	"net/http"
	"regexp"
	"strings"

	"github.com/nmloc/gh-achievements/internal/webui"
)

// markdownFS holds the Markdown twin of every static page. cmd/gendocs writes it from the
// same catalog as the HTML, so the two cannot drift: markdown/<path>index.md answers /<path>.
//
//go:embed markdown
var markdownFS embed.FS

// MarkdownRoute is the vercel.json route source (PCRE) for paths that answer
// Accept: text/markdown: every page path, but not files (a dot in the path) and not the
// functions, whose HTML has no Markdown twin. $1 is the path without its leading slash.
// WantsMarkdown applies the same rule in Go; cmd/devserver and the tests rely on it.
const MarkdownRoute = `^/(?!api/|auth/|result(?:/|$))([^.]*)$`

// MarkdownAccept is the vercel.json `has` value matched against the Accept header.
const MarkdownAccept = `.*text/markdown.*`

var functionPath = regexp.MustCompile(`^/(?:api/|auth/|result(?:/|$))`)

// WantsMarkdown reports whether a request should get the Markdown twin instead of HTML.
func WantsMarkdown(r *http.Request) bool {
	p := r.URL.Path
	return strings.Contains(r.Header.Get("Accept"), "text/markdown") &&
		strings.HasPrefix(p, "/") && !strings.Contains(p, ".") && !functionPath.MatchString(p)
}

// Markdown serves GET /api/markdown?path=<path>, which vercel.json routes every page
// request with Accept: text/markdown to. An unknown path is a 404 with a Markdown body
// that points at the sitemap and llms.txt.
func Markdown(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	w.Header().Set("Vary", "Accept")
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Cache-Control", "public, s-maxage=3600, stale-while-revalidate=86400")
	body, status := markdownFor(r.URL.Query().Get("path")), http.StatusOK
	if body == nil { // a page added later should not stay a cached 404 for long
		body, status = notFoundMarkdown(), http.StatusNotFound
		w.Header().Set("Cache-Control", "public, s-maxage=60")
	}
	w.WriteHeader(status)
	if r.Method == http.MethodGet {
		w.Write(body)
	}
}

// markdownFor returns the twin of a page path such as "", "achievements" or
// "achievements/pull-shark/", or nil when there is none.
func markdownFor(path string) []byte {
	p := strings.Trim(path, "/")
	if strings.Contains(p, ".") || strings.Contains(p, "//") {
		return nil
	}
	file := "markdown/index.md"
	if p != "" {
		file = "markdown/" + p + "/index.md"
	}
	b, err := markdownFS.ReadFile(file)
	if err != nil {
		return nil
	}
	return b
}

func notFoundMarkdown() []byte {
	b, err := markdownFS.ReadFile("markdown/404.md")
	if err != nil { // only if gendocs was never run; still a useful body
		return []byte("# Page not found\n\nSee " + webui.SiteURL + "llms.txt for every page.\n")
	}
	return b
}
