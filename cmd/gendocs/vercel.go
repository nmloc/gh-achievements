package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nmloc/gh-achievements/internal/web"
	"github.com/nmloc/gh-achievements/internal/webui"
)

func vercelJSON() ([]byte, error) {
	headers := []map[string]string{}
	for _, h := range web.SecurityHeaders() {
		headers = append(headers, map[string]string{"key": h.Key, "value": h.Value})
	}
	// The sign-in callback carries a one-time code in the URL fragment. GitHub redirects
	// without a trailing slash, so both spellings need the rule. They come after the
	// global rule so their Referrer-Policy wins.
	callback := []map[string]string{
		{"key": "Referrer-Policy", "value": "no-referrer"},
		{"key": "Cache-Control", "value": "no-store"},
		{"key": "X-Robots-Tag", "value": "noindex"},
	}
	cfg := map[string]any{
		"$schema":         "https://openapi.vercel.sh/vercel.json",
		"outputDirectory": "site",
		"functions":       map[string]any{"api/**/*.go": map[string]any{"maxDuration": 60}},
		"rewrites": []map[string]string{
			{"source": "/result", "destination": "/api/result"},
			{"source": "/auth/login", "destination": "/api/login"},
		},
		"headers": []map[string]any{
			{"source": "/(.*)", "headers": headers},
			{"source": "/auth/callback", "headers": callback},
			{"source": "/auth/callback/(.*)", "headers": callback},
		},
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	return append(b, '\n'), err
}

// redirectPages replaces the GitHub Pages site with pages that send every old URL to
// the same path on Vercel, so links already out there keep working.
func redirectPages() map[string][]byte {
	paths := append([]string{"", "achievements/", "methodology/"}, sitemapPaths()...)
	out := map[string][]byte{}
	for _, p := range paths {
		to := webui.SiteURL + p
		file := "pages/" + p + "index.html"
		out[file] = []byte(fmt.Sprintf(`<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<title>Moved to gh-achievements.vercel.app</title>
<link rel="canonical" href="%[1]s">
<meta http-equiv="refresh" content="0; url=%[1]s">
<meta name="robots" content="noindex">
</head><body><p>gh-achievements moved to <a href="%[1]s">%[1]s</a>.</p></body></html>
`, to))
	}
	out["pages/404.html"] = out["pages/index.html"]
	out["pages/.nojekyll"] = []byte{}
	return out
}

func sitemapPaths() []string {
	var ps []string
	for _, u := range sitemapURLs() {
		if p := strings.TrimPrefix(u, webui.SiteURL); p != "" && p != "achievements/" && p != "methodology/" {
			ps = append(ps, p)
		}
	}
	return ps
}
