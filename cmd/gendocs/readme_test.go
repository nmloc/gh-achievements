package main

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/nmloc/gh-achievements/internal/catalog"
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
