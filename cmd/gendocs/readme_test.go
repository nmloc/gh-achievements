package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/fixtures"
	"github.com/nmloc/gh-achievements/internal/render"
	"github.com/nmloc/gh-achievements/internal/report"
	"github.com/nmloc/gh-achievements/internal/webui"
)

func TestReadmeBadges_showEveryEarnableBadgeAtItsTopTier(t *testing.T) {
	got := readmeBadges()
	earnable := 0
	for _, a := range catalog.Achievements {
		if a.Kind != catalog.Earnable {
			if strings.Contains(got, `alt="`+a.Name+`"`) {
				t.Errorf("%s is not earnable but is in the hero row", a.Name)
			}
			continue
		}
		earnable++
		img := catalog.Image(a.Slug, len(a.Tiers))
		link := `<a href="` + webui.SiteURL + "achievements/" + a.Slug + `/">`
		if !strings.Contains(got, link+`<img src="`+img+`"`) || !strings.Contains(got, `alt="`+a.Name+`"`) {
			t.Errorf("%s missing from the hero row (want %s around %s)", a.Name, link, img)
		}
	}
	if n := strings.Count(got, "<img "); n != earnable {
		t.Fatalf("hero row has %d images, want %d", n, earnable)
	}
}

func TestReadmeTable_groupsBadgesAndShowsEveryTierMedal(t *testing.T) {
	table := readmeTable()
	last := -1
	for _, h := range []string{"### Earnable", "### Switched off by GitHub", "### Retired", "### GitHub staff only"} {
		i := strings.Index(table, h)
		if i < 0 {
			t.Errorf("%s missing", h)
			continue
		}
		if i < last {
			t.Errorf("%s out of order", h)
		}
		last = i
	}
	row := func(slug string) string {
		for _, line := range strings.Split(table, "\n") {
			if strings.Contains(line, "achievements/"+slug+"/)") {
				return line
			}
		}
		t.Fatalf("no row for %s", slug)
		return ""
	}
	ps := row("pull-shark")
	for i, want := range []string{"2", "16", "128", "1024"} {
		if !strings.Contains(ps, `<img src="`+catalog.Image("pull-shark", i+1)+`" width="48" height="48" alt="Pull Shark `+metals[i]+`">`) {
			t.Errorf("pull shark lacks its %s medal", metals[i])
		}
		if !strings.Contains(ps, `<td align="center">`+want+`</td>`) {
			t.Errorf("pull shark lacks the amount %s", want)
		}
	}
	if !strings.Contains(ps, "<th>Default</th><th>Bronze</th><th>Silver</th><th>Gold</th>") {
		t.Errorf("pull shark tier headings: %s", ps)
	}
	if !strings.Contains(ps, "<sub>gh achievements: count, checked against GitHub's milestones</sub>") {
		t.Errorf("pull shark lacks what gh achievements shows: %s", ps)
	}
	if heart := row("heart-on-your-sleeve"); strings.Count(heart, `<td align="center">?</td>`) != 2 {
		t.Errorf("heart on your sleeve shows unknown amounts as ?: %s", heart)
	}
	if yolo := row("yolo"); !strings.Contains(yolo, "<th>Default</th></tr>") || strings.Contains(yolo, "Bronze") ||
		!strings.Contains(yolo, `<td align="center">1</td>`) {
		t.Errorf("a single-tier badge has one medal and the amount 1: %s", yolo)
	}
	for _, a := range catalog.Achievements {
		if line := row(a.Slug); strings.Count(line, "|") != 5 {
			t.Errorf("%s row must stay one markdown row of 4 cells: %s", a.Slug, line)
		}
	}
	if !strings.Contains(table, "Profile highlights it also reports:") || !strings.Contains(table, "_Catalog last verified against live profiles on ") {
		t.Error("highlights and verified lines lost")
	}
}

func TestCheckButtonSVG_isWellFormedAndSaysWhatItDoes(t *testing.T) {
	svg := checkButtonSVG()
	d := xml.NewDecoder(strings.NewReader(string(svg)))
	for {
		if _, err := d.Token(); err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("button SVG is not well-formed XML: %v", err)
		}
	}
	for _, want := range []string{"Check your GitHub achievements", "No install", "private work", `role="img"`, "<title"} {
		if !strings.Contains(string(svg), want) {
			t.Errorf("button SVG lacks %q", want)
		}
	}
	// A fixed length keeps the arrow off the edge when a wider fallback font is used.
	if !strings.Contains(string(svg), `textLength="400" lengthAdjust="spacingAndGlyphs"`) {
		t.Error("the button title must have a fixed textLength so wider fonts cannot overflow")
	}
	if strings.Contains(string(svg), "<image") || strings.Contains(string(svg), "href=") {
		t.Error("README SVGs must be self-contained: GitHub does not load external resources inside them")
	}
}

func TestTerminalSVG_usesTheSitePaletteAndTheNewLayout(t *testing.T) {
	demo, err := report.Build("torvalds", fixtures.Official{}, fixtures.TorvaldsDemo())
	if err != nil {
		t.Fatal(err)
	}
	var colored bytes.Buffer
	render.Text(&colored, demo, render.Options{Color: true})
	svg := string(terminalSVG(colored.String(), ""))
	for _, want := range []string{`fill="#0d1c26"`, `fill="#e2b53a"`, `fill="#d39468"`, `fill="#93a9b7"`,
		"On display", "Still to collect", "72\u00a0/\u00a0128\u00a0to\u00a0x3", "The\u00a0archive"} {
		if !strings.Contains(svg, want) {
			t.Errorf("demo SVG lacks %q", want)
		}
	}
	for _, gone := range []string{"#0d1117", "#3fb950", "official\u00a0=\u00a0GitHub", "official = GitHub"} {
		if strings.Contains(svg, gone) {
			t.Errorf("demo SVG still has %q", gone)
		}
	}
}

func TestSVGs_matchTheSite(t *testing.T) {
	for name, svg := range map[string]string{"social": string(socialSVG()), "button": string(checkButtonSVG())} {
		for _, want := range []string{"#e2b53a"} {
			if !strings.Contains(svg, want) {
				t.Errorf("%s SVG lacks the site gold %s", name, want)
			}
		}
		for _, gone := range []string{"#2ea043", "#1a7f37", "#0d1117", "#fae57e"} {
			if strings.Contains(svg, gone) {
				t.Errorf("%s SVG still uses the old colour %s", name, gone)
			}
		}
	}
	if !strings.Contains(string(socialSVG()), "72\u00a0/\u00a0128\u00a0to\u00a0x3") { // the SVG keeps columns with no-break spaces
		t.Error("the social card should show the new progress wording")
	}
}

func TestReadme_showsTheSiteAndBothWaysToCheck(t *testing.T) {
	b, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	readme := string(b)
	counts := map[catalog.Kind]int{}
	for _, a := range catalog.Achievements {
		counts[a.Kind]++
	}
	scope := fmt.Sprintf("All %d known achievements (%d earnable, %d switched off by GitHub, %d retired, %d for GitHub staff) and %d profile highlights",
		len(catalog.Achievements), counts[catalog.Earnable], counts[catalog.Disabled], counts[catalog.Retired], counts[catalog.Internal], len(catalog.Highlights))
	if !strings.Contains(readme, scope) {
		t.Errorf("README lacks the sentence computed from the catalog: %q", scope)
	}
	for _, want := range []string{
		`<source media="(prefers-color-scheme: dark)" srcset="docs/screenshot-dark.png">`,
		`<img src="docs/screenshot-light.png"`,
		"## Two ways to check", "## What it shows", "gh achievements --all",
		"(docs/demo.svg)", "docs/check-button.svg",
		"https://github.com/nmloc/gh-achievements/stargazers",
		"is a [GitHub CLI](https://cli.github.com/) extension", "**No stored token.**",
		"until an admin approves gh-achievements",
	} {
		if !strings.Contains(readme, want) {
			t.Errorf("README lacks %q", want)
		}
	}
	for _, png := range []string{"../../docs/screenshot-light.png", "../../docs/screenshot-dark.png"} {
		if fi, err := os.Stat(png); err != nil || fi.Size() < 20_000 {
			t.Errorf("%s missing or too small", png)
		}
	}
}
