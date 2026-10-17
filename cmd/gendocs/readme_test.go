package main

import (
	"bytes"
	"encoding/xml"
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

func TestReadmeTable_hasTheBadgeArt(t *testing.T) {
	table := readmeTable()
	if !strings.HasPrefix(table, "| Badge | Achievement |") {
		t.Fatalf("table header = %q", strings.SplitN(table, "\n", 2)[0])
	}
	for _, a := range catalog.Achievements {
		img := catalog.Image(a.Slug, len(a.Tiers))
		row := ""
		for _, line := range strings.Split(table, "\n") {
			if strings.Contains(line, "achievements/"+a.Slug+"/)") {
				row = line
			}
		}
		switch {
		case row == "":
			t.Errorf("no row for %s", a.Slug)
		case img == "":
			t.Errorf("%s has no art at all", a.Slug)
		case img != "" && !strings.HasPrefix(row, `| <img src="`+img+`" width="40" height="40" alt="`):
			t.Errorf("%s row lacks its art: %s", a.Slug, row)
		}
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
