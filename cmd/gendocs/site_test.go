package main

import (
	"strings"
	"testing"
)

// Badges with no GitHub artwork (disabled ones) get a neutral placeholder, not nothing,
// so the badge page has a hero and the table rows stay aligned.
func TestSitePlaceholders_badgesWithoutArtGetAPlaceholder(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}

	heart := string(pages["site/achievements/heart-on-your-sleeve/index.html"])
	if !strings.Contains(heart, `class="badge-hero placeholder"`) || !strings.Contains(heart, ">HO</span>") {
		t.Errorf("heart-on-your-sleeve page has no HO placeholder hero")
	}
	open := string(pages["site/achievements/open-sourcerer/index.html"])
	if !strings.Contains(open, `class="badge-hero placeholder"`) || !strings.Contains(open, ">OS</span>") {
		t.Errorf("open-sourcerer page has no OS placeholder hero")
	}

	for _, path := range []string{"site/achievements/index.html", "site/index.html"} {
		page := string(pages[path])
		if n := strings.Count(page, `class="badge-icon placeholder"`); n != 2 {
			t.Errorf("%s has %d placeholder icons, want 2 (heart-on-your-sleeve, open-sourcerer)", path, n)
		}
	}
}
