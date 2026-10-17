// Package official reads what GitHub itself shows on a profile: earned badges, their
// tiers, unlock dates and milestone history, and the profile highlights.
//
// None of this is in GitHub's API. It comes from the public profile page and from the
// /users/<user>/achievements/<slug>/detail fragment GitHub loads into the badge dialog.
// Both are HTML, so every parser here is covered by fixtures saved from real profiles
// and by a weekly canary against live pages.
package official

import (
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Milestone is one row of a badge's history, e.g. "2nd pull request merged".
type Milestone struct {
	N     int    `json:"n,omitempty"`    // the ordinal GitHub counted to (2nd → 2), 0 if none
	PR    string `json:"pr,omitempty"`   // "owner/repo#123", empty when private ("inaccessible")
	Repo  string `json:"repo,omitempty"` // "owner/repo" for repository milestones (Starstruck)
	Label string `json:"label"`
}

// Detail is GitHub's own record of one earned badge.
type Detail struct {
	Slug       string      `json:"slug"`
	Name       string      `json:"name"`
	Tier       int         `json:"tier"`
	Unlocked   string      `json:"unlocked,omitempty"` // first unlock date, YYYY-MM-DD
	Milestones []Milestone `json:"milestones,omitempty"`
}

// Profile is what the profile page shows.
type Profile struct {
	// Hidden is true when the user turned off "Show Achievements on my profile".
	Hidden     bool
	Earned     []string // badge slugs, in profile order
	Highlights []string
}

// ErrUserNotFound means GitHub has no such user.
var ErrUserNotFound = errors.New("user not found")

var (
	reDetailLink = regexp.MustCompile(`/users/[^/"]+/achievements/([a-z0-9-]+)/detail`)
	reHighlights = regexp.MustCompile(`(?s)>Highlights</h2>(.*?)</ul>`)
	reListItem   = regexp.MustCompile(`(?s)<li[^>]*>(.*?)</li>`)
	reName       = regexp.MustCompile(`<h3>([^<]+)</h3>`)
	reTierCount  = regexp.MustCompile(`tier-count="(\d+)"`)
	reUnlocked   = regexp.MustCompile(`datetime="(\d{4}-\d{2}-\d{2})`)
	reHistoryRow = regexp.MustCompile(`(?s)achievement-history-tier.*?TimelineItem-body[^>]*>(.*?)</div>`)
	rePRLink     = regexp.MustCompile(`href="/([^/"]+/[^/"]+)/pull/(\d+)"`)
	reRepoLink   = regexp.MustCompile(`href="/([^/"]+/[^/"]+)"`)
	reOrdinal    = regexp.MustCompile(`\b(\d+)(?:st|nd|rd|th)\b`)
	reTags       = regexp.MustCompile(`<[^>]+>`)
	reSpace      = regexp.MustCompile(`\s+`)
)

// Text strips tags and collapses whitespace.
func Text(fragment string) string {
	s := reTags.ReplaceAllString(fragment, " ")
	return strings.TrimSpace(reSpace.ReplaceAllString(html.UnescapeString(s), " "))
}

// ParseEarned returns the earned badge slugs on an achievements tab, without duplicates.
func ParseEarned(page string) []string {
	seen := map[string]bool{}
	var slugs []string
	for _, m := range reDetailLink.FindAllStringSubmatch(page, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			slugs = append(slugs, m[1])
		}
	}
	return slugs
}

// ParseHighlights returns the labels in the profile's Highlights list.
func ParseHighlights(page string) []string {
	block := reHighlights.FindStringSubmatch(page)
	if block == nil {
		return nil
	}
	var labels []string
	for _, li := range reListItem.FindAllStringSubmatch(block[1], -1) {
		if label := Text(li[1]); label != "" {
			labels = append(labels, label)
		}
	}
	return labels
}

var reLeadingCount = regexp.MustCompile(`^\d+\s+`)

// normalizeHighlight drops a leading count and a plural "s":
// "33 security advisory credits" → "security advisory credit".
func normalizeHighlight(label string) string {
	s := strings.ToLower(reLeadingCount.ReplaceAllString(strings.TrimSpace(label), ""))
	return strings.TrimSuffix(s, "s")
}

// HasHighlight matches a highlight by its whole label, ignoring a leading count and a
// plural. A substring match would let "Pro" match "Developer Program Member".
func HasHighlight(highlights []string, name string) bool {
	want := normalizeHighlight(name)
	for _, h := range highlights {
		if normalizeHighlight(h) == want {
			return true
		}
	}
	return false
}

// ParseDetail reads one badge's detail fragment.
func ParseDetail(slug, fragment string) Detail {
	d := Detail{Slug: slug, Name: slug, Tier: 1}
	if m := reName.FindStringSubmatch(fragment); m != nil {
		d.Name = strings.TrimSpace(html.UnescapeString(m[1]))
	}
	if m := reTierCount.FindStringSubmatch(fragment); m != nil {
		d.Tier, _ = strconv.Atoi(m[1])
	}
	if m := reUnlocked.FindStringSubmatch(fragment); m != nil {
		d.Unlocked = m[1]
	}
	for _, row := range reHistoryRow.FindAllStringSubmatch(fragment, -1) {
		body := row[1]
		label := Text(body)
		ms := Milestone{Label: label}
		// The last ordinal: a repository name such as "me/1st-app" comes first.
		if all := reOrdinal.FindAllStringSubmatch(label, -1); all != nil {
			ms.N, _ = strconv.Atoi(all[len(all)-1][1])
		}
		if m := rePRLink.FindStringSubmatch(body); m != nil {
			ms.PR = m[1] + "#" + m[2]
		} else if m := reRepoLink.FindStringSubmatch(body); m != nil {
			ms.Repo = m[1]
		}
		d.Milestones = append(d.Milestones, ms)
	}
	return d
}

// Client fetches official pages from github.com.
type Client struct {
	HTTP *http.Client
	Base string // https://github.com
}

// NewClient returns a client for github.com.
func NewClient() *Client {
	return &Client{HTTP: &http.Client{Timeout: 30 * time.Second}, Base: "https://github.com"}
}

// get returns the page body, or "", false on 404.
func (c *Client) get(path string, xhr bool) (string, bool, error) {
	req, err := http.NewRequest(http.MethodGet, c.Base+path, nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("User-Agent", "gh-achievements (+https://github.com/nmloc/gh-achievements)")
	if xhr {
		// The detail fragment answers 406 without it.
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("GET %s: HTTP %d", path, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	return string(body), true, err
}

// Profile reads the achievements tab. The tab answers 404 both for a missing user and
// for one who hides achievements, so the profile page tells the two apart.
func (c *Client) Profile(user string) (Profile, error) {
	user = url.PathEscape(user)
	tab, found, err := c.get("/"+user+"?tab=achievements", false)
	if err != nil {
		return Profile{}, err
	}
	if found {
		return Profile{Earned: ParseEarned(tab), Highlights: ParseHighlights(tab)}, nil
	}
	main, found, err := c.get("/"+user, false)
	if err != nil {
		return Profile{}, err
	}
	if !found {
		return Profile{}, ErrUserNotFound
	}
	return Profile{Hidden: true, Highlights: ParseHighlights(main)}, nil
}

// Detail reads GitHub's record of one earned badge. It returns nil when the badge is
// not earned, since GitHub serves the fragment only for earned badges.
func (c *Client) Detail(user, slug string) (*Detail, error) {
	frag, found, err := c.get("/users/"+url.PathEscape(user)+"/achievements/"+url.PathEscape(slug)+"/detail", true)
	if err != nil || !found {
		return nil, err
	}
	d := ParseDetail(slug, frag)
	return &d, nil
}
