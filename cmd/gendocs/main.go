// Command gendocs builds everything readers see from internal/catalog: the README
// achievements table, the GitHub Pages site under site/, achievements.json, llms.txt,
// the sitemap and the README demo image. One source, so nothing drifts.
//
//	go run ./cmd/gendocs          # write
//	go run ./cmd/gendocs -check   # fail if anything is out of date (CI)
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/fixtures"
	"github.com/nmloc/gh-achievements/internal/render"
	"github.com/nmloc/gh-achievements/internal/report"
	"github.com/nmloc/gh-achievements/internal/webui"
)

const (
	tagline  = "Check your GitHub achievements progress: every badge, your tier and how far to the next one, verified against GitHub's own data, no token needed."
	readmeIn = "README.md"
)

func main() {
	check := flag.Bool("check", false, "fail if generated files are out of date")
	flag.Parse()

	files, err := generate()
	if err != nil {
		fail(err)
	}
	var stale []string
	for _, path := range sortedKeys(files) {
		want := files[path]
		got, err := os.ReadFile(path)
		if err == nil && bytes.Equal(got, want) {
			continue
		}
		if *check {
			stale = append(stale, path)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fail(err)
		}
		if err := os.WriteFile(path, want, 0o644); err != nil {
			fail(err)
		}
		fmt.Println("wrote", path)
	}
	if len(stale) > 0 {
		fail(fmt.Errorf("out of date, run `go run ./cmd/gendocs`:\n  %s", strings.Join(stale, "\n  ")))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gendocs:", err)
	os.Exit(1)
}

func sortedKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func generate() (map[string][]byte, error) {
	files := map[string][]byte{}

	demo, err := report.Build("torvalds", fixtures.Official{}, fixtures.TorvaldsDemo())
	if err != nil {
		return nil, fmt.Errorf("demo report: %w", err)
	}
	var colored bytes.Buffer
	render.Text(&colored, demo, render.Options{Color: true})
	svg := terminalSVG(colored.String(), "DISABLED")
	files["docs/demo.svg"] = svg
	// social.png is rendered from this by scripts/social-png.sh; PNG bytes depend on
	// the renderer version, so only the SVG is checked.
	files["docs/social.svg"] = socialSVG()

	readme, err := os.ReadFile(readmeIn)
	if err != nil {
		return nil, err
	}
	updated, err := replaceBetween(string(readme), "<!-- catalog:start -->", "<!-- catalog:end -->", readmeTable())
	if err != nil {
		return nil, fmt.Errorf("README: %w", err)
	}
	updated, err = replaceBetween(updated, "<!-- badges:start -->", "<!-- badges:end -->", readmeBadges())
	if err != nil {
		return nil, fmt.Errorf("README: %w", err)
	}
	files[readmeIn] = []byte(updated)
	files["docs/check-button.svg"] = checkButtonSVG()

	pages, err := sitePages()
	if err != nil {
		return nil, err
	}
	for path, body := range pages {
		files[path] = body
	}

	data, err := catalogJSON()
	if err != nil {
		return nil, err
	}
	files["site/achievements.json"] = data
	files["site/llms.txt"] = []byte(llmsTxt())
	files["site/llms-full.txt"] = []byte(llmsFull())
	for path, body := range verificationFiles() {
		files[path] = body
	}
	files["site/robots.txt"] = []byte("User-agent: *\nAllow: /\n\nSitemap: " + webui.SiteURL + "sitemap.xml\n")
	files["site/sitemap.xml"] = []byte(sitemap())
	files["site/.nojekyll"] = []byte{}

	vj, err := vercelJSON()
	if err != nil {
		return nil, err
	}
	files["vercel.json"] = vj
	for path, body := range redirectPages() {
		files[path] = body
	}
	return files, nil
}

// googleSiteVerification is the file Google Search Console asks for to prove the site is ours.
// It is public by design: Google fetches it from the site root.
const googleSiteVerification = "google398fb29c0f7ca94a.html"

// verificationFiles are the ownership proofs search engines fetch from the site root.
func verificationFiles() map[string][]byte {
	return map[string][]byte{
		"site/" + googleSiteVerification: []byte("google-site-verification: " + googleSiteVerification),
	}
}

func replaceBetween(s, start, end, body string) (string, error) {
	i, j := strings.Index(s, start), strings.Index(s, end)
	if i < 0 || j < i {
		return "", fmt.Errorf("markers %s … %s not found", start, end)
	}
	return s[:i+len(start)] + "\n" + body + s[j:], nil
}

// ---------- shared wording ----------

func tierText(a catalog.Achievement) string {
	if !a.IsTiered() {
		return "one tier"
	}
	parts := make([]string, len(a.Tiers))
	for i, t := range a.Tiers {
		if t == 0 {
			parts[i] = "?"
		} else {
			parts[i] = fmt.Sprint(t)
		}
	}
	return strings.Join(parts, " · ")
}

func statusText(k catalog.Kind) string {
	switch k {
	case catalog.Earnable:
		return "Earnable"
	case catalog.Disabled:
		return "Disabled (tested, not earnable now)"
	case catalog.Retired:
		return "Retired (no longer earnable)"
	case catalog.Internal:
		return "Internal (GitHub staff only)"
	}
	return string(k)
}

func progressText(a catalog.Achievement) string {
	if a.Kind != catalog.Earnable {
		return "Earned or not"
	}
	switch a.Counter {
	case catalog.CounterMergedPRs:
		return "Count, checked against GitHub's milestones"
	case catalog.CounterStars:
		return "Stars of the repository GitHub names"
	case catalog.CounterAnswers:
		return "Count of accepted answers"
	}
	if a.IsTiered() {
		return "Official tier (no public counting rule)"
	}
	return "Earned or not (official)"
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

var tierNames = []string{"x1 · base", "x2 · bronze", "x3 · silver", "x4 · gold"}

// ---------- README ----------

// readmeBadges is the hero row: GitHub's own top-tier art for every earnable badge,
// each linking to its page. GitHub serves README images through its camo proxy.
func readmeBadges() string {
	var b strings.Builder
	b.WriteString("<p align=\"center\">\n")
	for _, a := range catalog.Achievements {
		if a.Kind != catalog.Earnable {
			continue
		}
		fmt.Fprintf(&b, "  <a href=\"%sachievements/%s/\"><img src=\"%s\" width=\"72\" height=\"72\" alt=\"%s\" title=\"%s\"></a>\n",
			webui.SiteURL, a.Slug, catalog.Image(a.Slug, len(a.Tiers)), a.Name, a.Name)
	}
	b.WriteString("</p>\n")
	return b.String()
}

// readmeTable groups the catalog like the site's list and, as GitHub's own badge dialog does,
// shows every tier as its medal above the amount it needs.
func readmeTable() string {
	var b strings.Builder
	for _, g := range kindGroups {
		fmt.Fprintf(&b, "### %s\n\n%s\n\n", g.Heading, g.Note)
		b.WriteString("| Badge | Achievement | How to earn it | Tiers |\n|:-:|---|---|:-:|\n")
		for _, a := range catalog.Achievements {
			if a.Kind != g.Kind {
				continue
			}
			art := ""
			if img := catalog.Image(a.Slug, len(a.Tiers)); img != "" {
				art = fmt.Sprintf(`<img src="%s" width="40" height="40" alt="%s badge">`, img, a.Name)
			}
			fmt.Fprintf(&b, "| %s | [%s](%sachievements/%s/) | %s<br><sub>gh achievements: %s</sub> | %s |\n",
				art, a.Name, webui.SiteURL, a.Slug, a.HowToEarn, lowerFirst(progressText(a)), tierTable(a))
		}
		b.WriteString("\n")
	}
	b.WriteString("Profile highlights it also reports: ")
	names := make([]string, len(catalog.Highlights))
	for i, h := range catalog.Highlights {
		names[i] = h.Name
	}
	b.WriteString(strings.Join(names, ", ") + ".\n\n")
	fmt.Fprintf(&b, "_Catalog last verified against live profiles on %s._\n", catalog.LastVerified)
	return b.String()
}

// tierTable is one inline HTML table: a column per tier, its medal above the amount needed.
// It stays on one line so the markdown row around it does not break.
func tierTable(a catalog.Achievement) string {
	n := min(max(len(a.Tiers), 1), len(metalTitles))
	var head, art, need strings.Builder
	for i := 0; i < n; i++ {
		amount := "1"
		if i < len(a.Tiers) {
			amount = "?"
			if a.Tiers[i] > 0 {
				amount = fmt.Sprint(a.Tiers[i])
			}
		}
		head.WriteString("<th>" + metalTitles[i] + "</th>")
		fmt.Fprintf(&art, `<td align="center"><img src="%s" width="48" height="48" alt="%s %s"></td>`,
			catalog.Image(a.Slug, i+1), a.Name, metals[i])
		need.WriteString(`<td align="center">` + amount + "</td>")
	}
	return "<table><tr>" + head.String() + "</tr><tr>" + art.String() + "</tr><tr>" + need.String() + "</tr></table>"
}

// metalTitles head the tier columns, as on GitHub's own badge art.
var metalTitles = []string{"Default", "Bronze", "Silver", "Gold"}

// ---------- machine-readable ----------

func catalogJSON() ([]byte, error) {
	doc := map[string]interface{}{
		"name":         "gh-achievements catalog",
		"description":  "Every known GitHub profile achievement, its tiers and whether it can be earned.",
		"url":          webui.SiteURL + "achievements.json",
		"source":       webui.RepoURL,
		"lastVerified": catalog.LastVerified,
		"license":      "MIT",
		"achievements": catalog.Achievements,
		"highlights":   catalog.Highlights,
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

func llmsTxt() string {
	var b strings.Builder
	b.WriteString("# gh-achievements\n\n")
	b.WriteString("> " + tagline + "\n\n")
	b.WriteString("gh-achievements is an open source GitHub CLI extension (MIT). Install with `" + webui.Install + "`, run `gh achievements [username]`. ")
	b.WriteString("It reads tiers, unlock dates and milestone history from GitHub's own profile pages and counts progress with the user's existing gh login, so no personal access token is created. The web checker signs in with GitHub on a gh-achievements server on Vercel, uses the token for one check and revokes it. ")
	b.WriteString("A web checker at " + webui.SiteURL + "check/ shows the same report without installing anything. ")
	fmt.Fprintf(&b, "It knows %d achievements (%s) and %d profile highlights. Catalog last verified %s. It is not affiliated with GitHub.\n\n",
		len(catalog.Achievements), kindSummary(), len(catalog.Highlights), catalog.LastVerified)
	b.WriteString("## Achievements\n\n")
	for _, a := range catalog.Achievements {
		fmt.Fprintf(&b, "- [%s](%sachievements/%s/): %s. Tiers: %s. %s.\n", a.Name, webui.SiteURL, a.Slug, a.HowToEarn, tierText(a), statusText(a.Kind))
	}
	b.WriteString("\n## Docs\n\n")
	b.WriteString("- [All GitHub achievements](" + webui.SiteURL + "achievements/): the full list with tiers\n")
	b.WriteString("- [How accurate is it](" + webui.SiteURL + "methodology/): data sources, calibration, known limits\n")
	b.WriteString("- [Source code](" + webui.RepoURL + ")\n")
	b.WriteString("\n## Optional\n\n")
	b.WriteString("- [achievements.json](" + webui.SiteURL + "achievements.json): the catalog as JSON\n")
	b.WriteString("- [llms-full.txt](" + webui.SiteURL + "llms-full.txt): every page as plain text\n")
	return b.String()
}

func kindSummary() string {
	counts := map[catalog.Kind]int{}
	for _, a := range catalog.Achievements {
		counts[a.Kind]++
	}
	return fmt.Sprintf("%d earnable, %d disabled, %d retired, %d internal",
		counts[catalog.Earnable], counts[catalog.Disabled], counts[catalog.Retired], counts[catalog.Internal])
}

func llmsFull() string {
	var b strings.Builder
	b.WriteString(llmsTxt())
	for _, a := range catalog.Achievements {
		fmt.Fprintf(&b, "\n---\n\n# %s (GitHub achievement)\n\nURL: %sachievements/%s/\nStatus: %s\nHow to earn: %s\nTiers: %s\n\n%s\n",
			a.Name, webui.SiteURL, a.Slug, statusText(a.Kind), a.HowToEarn, tierText(a), a.Details)
		for _, qa := range a.FAQ {
			fmt.Fprintf(&b, "\nQ: %s\nA: %s\n", qa.Q, qa.A)
		}
	}
	b.WriteString("\n---\n\n# How accurate is gh-achievements\n\n")
	for _, p := range methodology {
		fmt.Fprintf(&b, "## %s\n\n%s\n\n", p.Heading, p.Text)
	}
	b.WriteString("# FAQ\n")
	for _, qa := range siteFAQ() {
		fmt.Fprintf(&b, "\nQ: %s\nA: %s\n", qa.Q, qa.A)
	}
	return b.String()
}

func sitemapURLs() []string {
	urls := []string{webui.SiteURL, webui.SiteURL + "achievements/", webui.SiteURL + "methodology/", webui.SiteURL + "check/"}
	for _, a := range catalog.Achievements {
		urls = append(urls, webui.SiteURL+"achievements/"+a.Slug+"/")
	}
	return urls
}

func sitemap() string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, u := range sitemapURLs() {
		fmt.Fprintf(&b, "  <url><loc>%s</loc><lastmod>%s</lastmod></url>\n", u, catalog.LastVerified)
	}
	b.WriteString("</urlset>\n")
	return b.String()
}
