// Package webui holds what every page shares: site constants, the layout template,
// CSS and scripts. cmd/gendocs renders static pages with it; web renders
// results with it, so the two never drift.
package webui

import (
	"bytes"
	"embed"
	"html/template"
	"strings"
	"unicode"
)

const (
	SiteURL         = "https://gh-achievements.vercel.app/"
	BasePath        = "/"
	OldSiteURL      = "https://nmloc.github.io/gh-achievements/"
	RepoURL         = "https://github.com/nmloc/gh-achievements"
	Install         = "gh extension install nmloc/gh-achievements"
	GAMeasurementID = "G-3EEKKGL110"
)

//go:embed templates/*.html static/*
var fs embed.FS

// Crumb is one breadcrumb entry.
type Crumb struct{ Name, URL string }

// Page is one rendered page. Content names the template file that defines "content".
type Page struct {
	Path, URL, Title, Description, Heading, Content string
	NoIndex, NoReferrer                             bool
	Crumbs                                          []Crumb
	JSONLD                                          []template.JS
	Scripts                                         []string // extra script paths, e.g. "/callback.js"
	Data                                            any
}

// Rel keeps internal links root-relative, so a local preview stays local.
func Rel(u string) string { return strings.Replace(u, SiteURL, BasePath, 1) }

// Initials is the first letter of the first two words, upper-case: "Open Sourcerer" → "OS".
func Initials(name string) string {
	words := strings.Fields(name)
	var out []rune
	for _, w := range words[:min(2, len(words))] {
		out = append(out, unicode.ToUpper([]rune(w)[0]))
	}
	return string(out)
}

func parse(files ...string) (*template.Template, error) {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = "templates/" + f
	}
	return template.New(files[0]).Funcs(template.FuncMap{"rel": Rel, "repoURL": func() string { return RepoURL }}).ParseFS(fs, paths...)
}

// Render renders a full page: layout plus p.Content.
func Render(p Page) ([]byte, error) {
	// result.html defines only "result" and "card", never "content", so every page can
	// embed a result without one page's content overriding another's. parts.html holds
	// snippets several pages share.
	files := []string{"layout.html", p.Content, "parts.html", "result.html"}
	t, err := parse(files...)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "layout", map[string]any{
		"P": p, "SiteURL": SiteURL, "RepoURL": RepoURL, "Base": BasePath, "GAID": GAMeasurementID,
	})
	return buf.Bytes(), err
}

// RenderFragment executes one named template from result.html or parts.html without the layout.
func RenderFragment(name string, data any) ([]byte, error) {
	t, err := parse("result.html", "parts.html")
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, name, data)
	return buf.Bytes(), err
}

// Static returns a file from static/ (style.css, app.js, callback.js).
func Static(name string) ([]byte, error) { return fs.ReadFile("static/" + name) }
