# Vitrine Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restyle every page of gh-achievements.vercel.app as a lit trophy cabinet ("Vitrine") with a light/dark theme switch, a star-the-repo call to action and the new `count / target to xN` progress lines, without losing any SEO/GEO property.

**Architecture:** The site is server-rendered Go `html/template` pages: `cmd/gendocs` renders the static pages into `site/`, and `web/` renders `/result` and the signed-in result fragment with the same templates from `internal/webui`. All design lives in `internal/webui/static/style.css` plus three small scripts (`theme.js` in `<head>`, `app.js` deferred, `callback.js` on the sign-in page). Card wording comes from `web/view.go`.

**Tech Stack:** Go 1.2x standard library (`html/template`, `embed`), vanilla JS (ES5 style, no build step), CSS custom properties, a self-hosted Bricolage Grotesque variable woff2.

**Spec:** `docs/superpowers/specs/2026-10-18-vitrine-redesign-design.md`

## Global Constraints

- Tokens, verbatim from the spec. Light: `--bg #eef2f4`, `--surface #ffffff`, `--ink #142430`, `--muted #5b6e7a`, `--line #cfd9df`, `--shelf #dfe6ea`, `--edge #c3ced5`, `--spot rgba(255,255,255,.95)`, `--gold #c9971f`, `--silver #8a96a0`, `--bronze #b5754a`. Dark: `--bg #0d1c26`, `--surface #122633`, `--ink #eaf1f5`, `--muted #93a9b7`, `--line #25414f`, `--shelf #16303f`, `--edge #08131a`, `--spot rgba(64,110,140,.55)`, `--gold #e2b53a`, `--silver #c7d0d6`, `--bronze #d39468`.
- Font: Bricolage Grotesque (SIL OFL), one variable woff2 under `/fonts/`, `font-display: swap`, system-ui fallback. Body 17px / 1.6, headings on a 1.25 modular scale. Sentence case everywhere; no all-caps labels.
- Theme: default follows `prefers-color-scheme`; the header button stores `gha:theme` in `localStorage`; `/theme.js` runs in `<head>` **before** the stylesheet and sets `data-theme` on `<html>`. CSS: `:root` = light, `:root[data-theme="dark"]` and `@media (prefers-color-scheme: dark) { :root:not([data-theme="light"]) }` = dark. The button says what it will do ("Dark" / "Light") and has an `aria-label`.
- CSP unchanged except `connect-src` gains `https://api.github.com`. No inline scripts or styles. No build step. No new Go dependencies. `html/template` escaping everywhere.
- GA rules unchanged: consent-gated, no usernames or repository names ever reach GA. The only new event is `star_click {place: header|result|home}`.
- Every SEO/GEO property stays: titles, meta descriptions, canonical URLs, JSON-LD, sitemap, llms.txt, achievements.json, the redirect tree.
- Every number keeps its source label (official / calibrated / counted).
- Out of scope: new features, copy beyond what the layout needs, changes to the CLI output.
- Commits: author `nmloc <nmloc312@gmail.com>` (already the repo's git config), Conventional Commits, **no** `Co-Authored-By` trailer and no AI attribution. Never push.
- After any change to templates, static files or `cmd/gendocs`, regenerate and commit generated files: `go run ./cmd/gendocs` (CI runs `go run ./cmd/gendocs -check`), and `go test ./web -run TestRenderResult_golden -update` when the result markup changes. `gofmt -l .` must print nothing.

## Rulings on what the spec leaves open

1. **Units in the progress line.** The spec shows `72 / 128 to x3` (Pull Shark, no unit) and `0 / 2 answers to x1` (Galaxy Brain, unit). Only badges whose name does not say what is counted get a short unit: `starstruck` → `stars`, `galaxy-brain` → `answers`. The top-tier line adds the count only for those: `Top tier, 252,192 stars`; Pull Shark at the top is just `Top tier`.
2. **Other lines.** Tiered badge without a counter at x0: `First tier at 1` (`First tier: threshold unknown` when GitHub's threshold is unknown); above x0 with an unknown threshold: `x3, next tier unknown`. An earnable single-tier badge not yet earned shows its how-to (`Close an issue or pull request within 5 minutes of opening it`). All lines are sentence case: `Hidden on this profile`, `Switched off by GitHub`, `No longer earnable`, `GitHub staff only`, `Earned`.
3. **Unlock date.** `Since 2020` keeps only the year, so the card's detail gains "First unlocked 2020-02-02."
4. **Quick wins** are dropped: the "Still to collect" shelf shows the how-to of every single-tier badge, which is what quick wins listed.
5. **Result page order:** header, scope line, hidden notice, On display, Still to collect, The archive, Highlights, star banner, actions, source legend.
6. **One `<h1>` per page.** The layout prints `<h1>` only when `Page.Heading` is set. The home page sets none and its hero holds `<h1>Your GitHub trophy case</h1>`; the `/result` page sets none and the result header's `@login` is the `<h1>` (it stays `<h2>` in the signed-in fragment, which sits under the callback page's `<h1>`).
7. **Star pill text.** `★ Star on GitHub` before the count loads; `★ Star 128` after (the " on GitHub" part is hidden). Under 640px " on GitHub" is always hidden.
8. **Menu under 640px** is a button (`aria-expanded`, `aria-controls="site-nav"`), not `<details>`: a closed `<details>` cannot show its content on desktop. `theme.js` adds `class="js"` to `<html>`, and the nav only collapses when that class is present, so without JavaScript the nav simply stays open.
9. **style.css is rewritten once, in Task 2**, including the rules for classes the page tasks (4–6) introduce. Pages look unfinished between Task 2 and Task 6; nothing ships mid-plan, and the controller checks every page in the browser in Task 7.
10. **Home page** drops the demo SVG and the full achievements table (both live elsewhere: README and `/achievements/`), so `site/demo.svg` is no longer generated and is deleted.
11. **Font file:** `@fontsource-variable/bricolage-grotesque@5.2.8`, `files/bricolage-grotesque-latin-wght-normal.woff2` (41,344 bytes, weights 200–800, Latin), plus the package's `LICENSE` (4,314 bytes, SIL OFL 1.1) shipped next to it as `OFL.txt`. Downloading needs Loc's go-ahead; the controller records it in the ledger before dispatching Task 2.

## File structure

| File | Responsibility | Tasks |
|---|---|---|
| `web/view.go` | Card wording and the result page model (shelves) | 1, 4 |
| `web/render.go` | Full result page vs fragment | 4 |
| `internal/webui/webui.go` | Template parsing (`parts.html`, `repoURL` func) | 3 |
| `internal/webui/templates/layout.html` | Head order, header (brand, nav, star pill, theme, menu), optional `<h1>` | 2, 3, 4 |
| `internal/webui/templates/parts.html` (new) | Shared snippets: `star-banner`, `checkform` | 3, 5 |
| `internal/webui/templates/result.html` | Result header, shelves, archive, card | 4 |
| `internal/webui/templates/home.html`, `check.html` | Hero, home shelf, choice card | 5 |
| `internal/webui/templates/list.html`, `badge.html`, `methodology.html`, `callback.html` | Medal grid, badge hero + tier strip, prose column, checking card | 6 |
| `internal/webui/static/style.css` | All visual design | 2 |
| `internal/webui/static/theme.js` (new) | Apply a stored theme before paint | 2 |
| `internal/webui/static/app.js` | Theme button, menu, star count, `star_click` | 2, 3 |
| `internal/webui/static/callback.js` | Fill the star count in the injected fragment | 4 |
| `internal/webui/static/fonts/` (new) | `bricolage-grotesque.woff2`, `OFL.txt` | 2 |
| `web/headers.go` | CSP `connect-src` | 3 |
| `cmd/gendocs/site.go`, `main.go` | Copy new static files, home/list data, drop `site/demo.svg` | 2, 5, 6 |

---

### Task 1: Progress lines show count / target

**Files:**
- Modify: `web/view.go` (`newCard`, `fixedLine`, `nextTierLine`; add `countLine`, `topTierLine`, `sinceLine`, `group`, `shortUnit`, `yearPattern`)
- Test: `web/view_test.go`
- Regenerate: `web/testdata/result-public.html`, `web/testdata/result-private.html`

**Interfaces:**
- Consumes: `report.Row`, `progress.Standing{Tier, Count, Floor, Target, Remaining, IsMax, IsAhead}` (unchanged).
- Produces: `Card.Line` strings in the new format (later tasks only print them). `nextTierLine(tier, threshold int) string` now takes the **current** tier.

- [ ] **Step 1: Write the failing test**

Add to `web/view_test.go`:

```go
func TestCard_lineShowsTheCountAgainstTheNextThreshold(t *testing.T) {
	pullShark := []int{2, 16, 128, 1024}
	stars := []int{16, 128, 512, 4096}
	cases := []struct {
		name string
		row  report.Row
		want string
	}{
		{"count / target", report.Row{Slug: "pull-shark", Kind: catalog.Earnable, Tiers: pullShark, Status: report.Earned, Tier: 2,
			Standing: &progress.Standing{Tier: 2, Count: 72, Floor: 16, Target: 128, Remaining: 56}}, "72 / 128 to x3"},
		{"with a unit", report.Row{Slug: "galaxy-brain", Kind: catalog.Earnable, Tiers: []int{2, 8, 16, 32}, Status: report.NotEarned,
			Standing: &progress.Standing{Tier: 0, Count: 0, Floor: 0, Target: 2, Remaining: 2}}, "0 / 2 answers to x1"},
		{"thousands", report.Row{Slug: "starstruck", Kind: catalog.Earnable, Tiers: stars, Status: report.Earned, Tier: 3,
			Standing: &progress.Standing{Tier: 3, Count: 1200, Floor: 512, Target: 4096, Remaining: 2896}}, "1,200 / 4,096 stars to x4"},
		{"top tier with a unit", report.Row{Slug: "starstruck", Kind: catalog.Earnable, Tiers: stars, Status: report.Earned, Tier: 4,
			Standing: &progress.Standing{Tier: 4, Count: 252192, Floor: 4096, IsMax: true}}, "Top tier, 252,192 stars"},
		{"top tier without a unit", report.Row{Slug: "pull-shark", Kind: catalog.Earnable, Tiers: pullShark, Status: report.Earned, Tier: 4,
			Standing: &progress.Standing{Tier: 4, Count: 1500, Floor: 1024, IsMax: true}}, "Top tier"},
		{"tiered, no counter", report.Row{Slug: "pair-extraordinaire", Kind: catalog.Earnable, Tiers: []int{1, 10, 24, 48},
			Status: report.Earned, Tier: 2}, "x2, next tier at 24"},
		{"tiered, no counter, top", report.Row{Slug: "pair-extraordinaire", Kind: catalog.Earnable, Tiers: []int{1, 10, 24, 48},
			Status: report.Earned, Tier: 4}, "Top tier"},
		{"tiered, not started", report.Row{Slug: "finishing-touch", Kind: catalog.Earnable, Tiers: []int{1, 4, 16, 32},
			Status: report.NotEarned}, "First tier at 1"},
		{"tiered, first threshold unknown", report.Row{Slug: "x", Kind: catalog.Earnable, Tiers: []int{0, 16, 128, 0},
			Status: report.NotEarned}, "First tier: threshold unknown"},
		{"next threshold unknown", report.Row{Slug: "x", Kind: catalog.Earnable, Tiers: []int{0, 16, 128, 0},
			Status: report.Earned, Tier: 3}, "x3, next tier unknown"},
		{"single tier, earned", report.Row{Slug: "quickdraw", Kind: catalog.Earnable, Tiers: []int{1}, Status: report.Earned,
			Tier: 1, Unlocked: "2020-02-02"}, "Since 2020"},
		{"single tier, earned, no date", report.Row{Slug: "quickdraw", Kind: catalog.Earnable, Tiers: []int{1},
			Status: report.Earned, Tier: 1}, "Earned"},
		{"single tier, not earned", report.Row{Slug: "quickdraw", Kind: catalog.Earnable, Tiers: []int{1}, Status: report.NotEarned,
			HowToEarn: "Close an issue or pull request within 5 minutes of opening it"},
			"Close an issue or pull request within 5 minutes of opening it"},
	}
	for _, tc := range cases {
		tc.row.Name = "Badge"
		if got := newCard(tc.row).Line; got != tc.want {
			t.Errorf("%s: line = %q, want %q", tc.name, got, tc.want)
		}
	}
}
```

Update the existing assertions in the same file:

- `TestResultView_closestNextUnlockIsTheSmallestRemaining`: replace `!strings.Contains(v.Closest.Line, "2 to x1")` with `v.Closest.Line != "0 / 2 answers to x1"`.
- `TestResultView_cards`: replace `ps.Line != "56 to x3"` with `ps.Line != "72 / 128 to x3"`.
- `TestCard_nonEarnableKindsNeverShowProgress`: `heart.Line != "Switched off by GitHub"`; the map becomes `catalog.Disabled: "Switched off by GitHub"`, `catalog.Retired: "No longer earnable"`, `catalog.Internal: "GitHub staff only"`; the earned retired row with `Unlocked: "Jan 2021"` wants `"Since 2021"`, and with `Unlocked: ""` wants `"Earned"`.
- `TestResultView_hiddenProfileKeepsCountedClosest`: both `"hidden on this profile"` literals become `"Hidden on this profile"`.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./web -run 'TestCard_|TestResultView_'`
Expected: FAIL, e.g. `count / target: line = "56 to x3", want "72 / 128 to x3"`.

- [ ] **Step 3: Implement**

In `web/view.go` add `"regexp"` and `"strconv"` to the imports, and below `sourceMeaning`:

```go
// shortUnit names what a count counts, for badges whose name does not already say it.
var shortUnit = map[string]string{"starstruck": "stars", "galaxy-brain": "answers"}

// yearPattern finds the year in an unlock date such as "2020-02-02" or "Jan 2021".
var yearPattern = regexp.MustCompile(`\b(19|20)\d{2}\b`)
```

Replace the `switch s := row.Standing; { … }` block in `newCard` with:

```go
	switch s := row.Standing; {
	case row.Kind != catalog.Earnable:
		c.Line = fixedLine(row.Kind, c.Earned, row.Unlocked)
	case s != nil && s.IsMax:
		c.Line = topTierLine(row.Slug, s.Count)
		c.Progress = &Meter{Value: 1, Max: 1, Label: row.Name + ": top tier"}
	case s != nil:
		c.Line = countLine(row.Slug, s.Count, s.Target, s.Tier+1)
		span := max(s.Target-s.Floor, 1)
		c.Progress = &Meter{Value: min(max(s.Count-s.Floor, 0), span), Max: span,
			Label: fmt.Sprintf("%s: %d of %d", row.Name, s.Count, s.Target)}
		if s.IsAhead {
			c.Notes = append(slices.Clone(c.Notes), aheadNote(row.Tier, s.Tier))
		}
	case row.Status == report.Unknown:
		c.Line = "Hidden on this profile"
	case len(row.Tiers) > 1 && row.Tier < len(row.Tiers):
		c.Line = nextTierLine(row.Tier, row.Tiers[row.Tier])
	case len(row.Tiers) > 1:
		c.Line = "Top tier"
	case c.Earned:
		c.Line = sinceLine(row.Unlocked)
	default:
		c.Line = row.HowToEarn
	}
```

Replace `fixedLine` and `nextTierLine`, and add the helpers:

```go
// fixedLine is the status of a badge that can no longer be earned: no progress, no threshold.
func fixedLine(kind catalog.Kind, earned bool, unlocked string) string {
	switch {
	case earned:
		return sinceLine(unlocked)
	case kind == catalog.Disabled:
		return "Switched off by GitHub"
	case kind == catalog.Retired:
		return "No longer earnable"
	case kind == catalog.Internal:
		return "GitHub staff only"
	}
	return "Not yet"
}

// countLine is "count / next threshold to xN": where the count stands against the next tier.
func countLine(slug string, count, target, nextTier int) string {
	nums := group(count) + " / " + group(target)
	if unit := shortUnit[slug]; unit != "" {
		nums += " " + unit
	}
	return fmt.Sprintf("%s to x%d", nums, nextTier)
}

// topTierLine adds the count where it means something on its own ("252,192 stars").
func topTierLine(slug string, count int) string {
	if unit := shortUnit[slug]; unit != "" && count > 0 {
		return "Top tier, " + group(count) + " " + unit
	}
	return "Top tier"
}

// nextTierLine is for a tiered badge with no counter: the tier GitHub shows, then the next threshold.
func nextTierLine(tier, threshold int) string {
	switch {
	case tier == 0 && threshold > 0:
		return fmt.Sprintf("First tier at %d", threshold)
	case tier == 0:
		return "First tier: threshold unknown"
	case threshold > 0:
		return fmt.Sprintf("x%d, next tier at %d", tier, threshold)
	}
	return fmt.Sprintf("x%d, next tier unknown", tier)
}

// sinceLine is the year a badge was first unlocked, or "Earned" when GitHub gives no date.
func sinceLine(unlocked string) string {
	if y := yearPattern.FindString(unlocked); y != "" {
		return "Since " + y
	}
	if unlocked != "" {
		return "Since " + unlocked
	}
	return "Earned"
}

// group writes n with thousands separators: 252192 → "252,192". Counts are never negative.
func group(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
```

- [ ] **Step 4: Run the tests to verify they pass, then refresh the goldens**

Run: `go test ./web -run 'TestCard_|TestResultView_'` → PASS.
Run: `go test ./web -run TestRenderResult_golden -update && go test ./...` → PASS.
Check: `grep -c "72 / 128 to x3" web/testdata/result-public.html` prints at least `1`.

- [ ] **Step 5: Commit**

```bash
gofmt -l . && git add web/view.go web/view_test.go web/testdata && git commit -m "feat(web): show progress as count / next threshold"
```

---

### Task 2: Theme, font and the new stylesheet

**Files:**
- Create: `internal/webui/static/theme.js`, `internal/webui/static/fonts/bricolage-grotesque.woff2`, `internal/webui/static/fonts/OFL.txt`
- Replace: `internal/webui/static/style.css`
- Modify: `internal/webui/templates/layout.html` (head order, header wrapper, theme button), `internal/webui/static/app.js` (theme button), `cmd/gendocs/site.go` (copy list)
- Test: `internal/webui/webui_test.go`, `cmd/gendocs/site_test.go`

**Interfaces:**
- Produces, used by Tasks 3–6 (CSS class contract): `.top-in`, `.brand`, `#site-nav`, `.theme-btn`, `.menu-btn`, `.star`, `.star-long`, `.star-count`, `.btn` (`.primary`, `.gold`, `.small`, `.big`), `.notice` (`.error`), `.vitrine`, `.shelf`, `.home-shelf`, `.medal`, `.medal-img`, `.medal-name`, `.badge-card`, `.badge-img`, `.badge-name`, `.badge-line`, `.badge-detail`, `.tier .t1–.t4`, `.locked`, `.result-head`, `.avatar`, `.result-user`, `.result-sum`, `.next`, `.scope`, `.shelf-head`, `.hint`, `.archive`, `.highlights` (`.on`, `.off`), `.actions`, `.star-banner`, `.star-cta`, `.hero`, `.signin`, `.points`, `.install`, `.choice`, `.or`, `#check-form`, `.medal-grid`, `.badge-top`, `.badge-hero`, `.tier-strip`, `.prose`, `.facts`, `.faq`, `.crumbs`, `.pager`, `.checking`, `.steps`, `.linklike`, `.visually-hidden`, `#consent`.
- Produces: `<html class="js">` once `theme.js` ran; `localStorage["gha:theme"]` ∈ `light|dark`; the header `<button type="button" class="theme-btn" data-theme-toggle hidden>`.

- [ ] **Step 1: Get the font (only after Loc's go-ahead is in the ledger)**

```bash
mkdir -p internal/webui/static/fonts
curl -fsSL -o internal/webui/static/fonts/bricolage-grotesque.woff2 https://cdn.jsdelivr.net/npm/@fontsource-variable/bricolage-grotesque@5.2.8/files/bricolage-grotesque-latin-wght-normal.woff2
curl -fsSL -o internal/webui/static/fonts/OFL.txt https://cdn.jsdelivr.net/npm/@fontsource-variable/bricolage-grotesque@5.2.8/LICENSE
wc -c internal/webui/static/fonts/*
grep -c "SIL OPEN FONT LICENSE" internal/webui/static/fonts/OFL.txt
```

Expected: `41344 …woff2`, `4314 …OFL.txt`, and `1` (or more). `file internal/webui/static/fonts/bricolage-grotesque.woff2` reports `Web Open Font Format (Version 2)`.

- [ ] **Step 2: Write the failing tests**

In `internal/webui/webui_test.go`, add `<script src="/theme.js"></script>` and `<button type="button" class="theme-btn" data-theme-toggle hidden>` to the `want` list of `TestRender_usesTheVercelDomainAndRootPaths`, and add:

```go
// The dark tokens are written twice: for an explicit choice and for the system setting.
// They must never drift apart.
func TestStyle_darkTokensMatchInBothBlocks(t *testing.T) {
	css, err := Static("style.css")
	if err != nil {
		t.Fatal(err)
	}
	block := func(selector string) string {
		s := string(css)
		i := strings.Index(s, selector+" {")
		if i < 0 {
			t.Fatalf("style.css has no %s block", selector)
		}
		rest := s[i+len(selector)+2:]
		return strings.Join(strings.Fields(rest[:strings.Index(rest, "}")]), " ")
	}
	explicit, system := block(`:root[data-theme="dark"]`), block(`:root:not([data-theme="light"])`)
	if explicit != system {
		t.Fatalf("dark tokens differ:\n explicit: %s\n system:   %s", explicit, system)
	}
	for _, want := range []string{"--bg: #0d1c26", "--gold: #e2b53a", "color-scheme: dark"} {
		if !strings.Contains(explicit, want) {
			t.Errorf("dark tokens lack %s", want)
		}
	}
	for _, want := range []string{"@media (prefers-color-scheme: dark)", "font-display: swap",
		`url("/fonts/bricolage-grotesque.woff2")`, "@media (prefers-reduced-motion: no-preference)"} {
		if !strings.Contains(string(css), want) {
			t.Errorf("style.css lacks %s", want)
		}
	}
}

func TestThemeJS_appliesOnlyAStoredLightOrDark(t *testing.T) {
	js, err := Static("theme.js")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`localStorage.getItem("gha:theme")`, `t === "light" || t === "dark"`,
		`setAttribute("data-theme", t)`, `classList.add("js")`} {
		if !strings.Contains(string(js), want) {
			t.Errorf("theme.js lacks %s", want)
		}
	}
}
```

In `cmd/gendocs/site_test.go` add:

```go
// theme.js must run before the stylesheet so the first paint is in the chosen theme,
// and the CSP forbids inline scripts on every page (JSON-LD data blocks excepted).
func TestSitePages_themeLoadsFirstAndNoInlineScript(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	for path, body := range pages {
		if !strings.HasSuffix(path, ".html") {
			continue
		}
		html := string(body)
		theme := strings.Index(html, `<script src="/theme.js"></script>`)
		css := strings.Index(html, `<link rel="stylesheet" href="/style.css">`)
		if theme < 0 || css < 0 || theme > css {
			t.Errorf("%s: theme.js (at %d) must load before the stylesheet (at %d)", path, theme, css)
		}
		for _, chunk := range strings.Split(html, "<script")[1:] {
			head := chunk[:strings.Index(chunk, ">")]
			if !strings.Contains(head, "src=") && !strings.Contains(head, `type="application/ld+json"`) {
				t.Errorf("%s: inline script <script%s>", path, head)
			}
		}
	}
	if len(pages["site/fonts/bricolage-grotesque.woff2"]) < 10000 || len(pages["site/theme.js"]) == 0 ||
		!strings.Contains(string(pages["site/fonts/OFL.txt"]), "SIL OPEN FONT LICENSE") {
		t.Error("theme.js, the font or its licence is not copied into site/")
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./internal/webui ./cmd/gendocs`
Expected: FAIL (`missing <script src="/theme.js"></script>`, `style.css has no :root[data-theme="dark"] block`, `open static/theme.js: file does not exist`).

- [ ] **Step 4: Create `internal/webui/static/theme.js`**

```js
// internal/webui/static/theme.js
// Runs in <head> before the stylesheet, so the first paint already uses the visitor's
// chosen theme. app.js owns the toggle button; this only applies a stored choice.
(function () {
  "use strict";
  var root = document.documentElement;
  root.classList.add("js");
  try {
    var t = localStorage.getItem("gha:theme");
    if (t === "light" || t === "dark") root.setAttribute("data-theme", t);
  } catch (e) { /* storage blocked: follow the system setting */ }
})();
```

- [ ] **Step 5: Replace `internal/webui/static/style.css` with exactly this**

```css
/* internal/webui/static/style.css
   Vitrine: badges are medals in a lit trophy cabinet. The only accent colours are the
   medals' own metals; everything else stays quiet. */

@font-face {
  font-family: "Bricolage Grotesque";
  src: url("/fonts/bricolage-grotesque.woff2") format("woff2");
  font-weight: 200 800;
  font-display: swap;
}

/* Light tokens. The dark set is written twice, for an explicit choice and for the system
   setting; webui_test checks the two copies match. */
:root {
  --bg: #eef2f4; --surface: #ffffff; --ink: #142430; --muted: #5b6e7a; --line: #cfd9df;
  --shelf: #dfe6ea; --edge: #c3ced5; --spot: rgba(255, 255, 255, .95);
  --gold: #c9971f; --silver: #8a96a0; --bronze: #b5754a; --metal-ink: #1d1600;
  --code-bg: #142430; --code-ink: #eaf1f5; --danger: #b3261e; --focus: #142430;
  color-scheme: light;
}
:root[data-theme="dark"] {
  --bg: #0d1c26; --surface: #122633; --ink: #eaf1f5; --muted: #93a9b7; --line: #25414f;
  --shelf: #16303f; --edge: #08131a; --spot: rgba(64, 110, 140, .55);
  --gold: #e2b53a; --silver: #c7d0d6; --bronze: #d39468; --metal-ink: #1d1600;
  --code-bg: #08131a; --code-ink: #eaf1f5; --danger: #f2b8b5; --focus: #e2b53a;
  color-scheme: dark;
}
@media (prefers-color-scheme: dark) {
  :root:not([data-theme="light"]) {
    --bg: #0d1c26; --surface: #122633; --ink: #eaf1f5; --muted: #93a9b7; --line: #25414f;
    --shelf: #16303f; --edge: #08131a; --spot: rgba(64, 110, 140, .55);
    --gold: #e2b53a; --silver: #c7d0d6; --bronze: #d39468; --metal-ink: #1d1600;
    --code-bg: #08131a; --code-ink: #eaf1f5; --danger: #f2b8b5; --focus: #e2b53a;
    color-scheme: dark;
  }
}

/* Base: 17px body, headings on a 1.25 scale (rem = 17px). */
* { box-sizing: border-box; }
[hidden] { display: none !important; }
html { font-size: 106.25%; -webkit-text-size-adjust: 100%; }
body {
  margin: 0; background: var(--bg); color: var(--ink);
  font-family: "Bricolage Grotesque", system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
  font-size: 1rem; line-height: 1.6;
}
a { color: inherit; text-decoration-color: var(--gold); text-decoration-thickness: 2px; text-underline-offset: 3px; }
a:hover { text-decoration-color: currentColor; }
:focus-visible { outline: 3px solid var(--focus); outline-offset: 2px; border-radius: 4px; }
h1, h2, h3 { line-height: 1.15; letter-spacing: -.015em; }
h1 { font-size: 2.441rem; font-weight: 800; margin: .5em 0 .35em; }
h2 { font-size: 1.563rem; font-weight: 700; margin: 2em 0 .5em; }
h3 { font-size: 1.25rem; font-weight: 700; margin: 1.5em 0 .4em; }
p { margin: 0 0 1em; }
.lede { font-size: 1.25rem; line-height: 1.5; color: var(--muted); max-width: 60ch; }
.meta, .hint { color: var(--muted); font-size: .875rem; }
.hint { margin: 0 0 12px; }
.prose { max-width: 68ch; }
code { font: .88em ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; background: var(--shelf); padding: .1em .35em; border-radius: 4px; overflow-wrap: anywhere; }
pre { background: var(--code-bg); color: var(--code-ink); padding: 14px 16px; border-radius: 10px; overflow-x: auto; line-height: 1.5; }
pre code { background: none; padding: 0; color: inherit; overflow-wrap: normal; }
.visually-hidden { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }

/* Header: brand, nav, the gold star pill (the header's only gold), theme and menu buttons. */
.top { background: var(--surface); border-bottom: 1px solid var(--line); }
.top-in { display: flex; align-items: center; gap: 10px 20px; max-width: 1040px; margin: 0 auto; padding: 12px 16px; }
.brand { margin-right: auto; font-weight: 800; font-size: 1.1rem; text-decoration: none; white-space: nowrap; }
.top nav { display: flex; gap: 18px; font-size: .94rem; }
.top nav a { color: var(--muted); text-decoration: none; }
.top nav a:hover { color: var(--ink); }
.theme-btn, .menu-btn {
  font: inherit; font-size: .875rem; font-weight: 600; color: var(--ink); background: var(--bg);
  border: 1px solid var(--line); border-radius: 999px; padding: 5px 12px; cursor: pointer;
}
.menu-btn { display: none; }
.star {
  display: inline-flex; align-items: center; gap: 6px; padding: 5px 12px; border-radius: 999px;
  background: var(--gold); color: var(--metal-ink); font-size: .875rem; font-weight: 700;
  text-decoration: none; white-space: nowrap; box-shadow: inset 0 -2px 0 rgba(0, 0, 0, .15);
}
.star:hover, .star-banner:hover { filter: brightness(1.06); }
.star-count { padding: 0 7px; border-radius: 999px; background: rgba(255, 255, 255, .5); font-variant-numeric: tabular-nums; }

main { max-width: 1040px; margin: 0 auto; padding: 8px 16px 0; }
footer { max-width: 1040px; margin: 64px auto 0; padding: 20px 16px 36px; border-top: 1px solid var(--line); color: var(--muted); font-size: .875rem; }

/* Controls */
.btn {
  display: inline-flex; align-items: center; gap: 8px; padding: 9px 16px; border-radius: 10px;
  border: 1px solid var(--line); background: var(--surface); color: var(--ink);
  font: inherit; font-size: .94rem; font-weight: 700; text-decoration: none; cursor: pointer;
}
.btn:hover { border-color: var(--muted); }
.btn.primary { background: var(--ink); border-color: var(--ink); color: var(--surface); }
.btn.gold { background: var(--gold); border-color: var(--gold); color: var(--metal-ink); }
.btn.small { padding: 4px 10px; font-size: .82rem; border-radius: 8px; }
.btn.big { padding: 12px 20px; font-size: 1.05rem; }
.linklike { background: none; border: 0; padding: 0; font: inherit; color: inherit; text-decoration: underline; text-decoration-color: var(--gold); cursor: pointer; }
#check-form { display: flex; gap: 8px; max-width: 460px; }
#check-form input {
  flex: 1; min-width: 0; padding: 10px 12px; border-radius: 10px; border: 1px solid var(--line);
  background: var(--surface); color: var(--ink); font: inherit;
}
.notice { background: var(--surface); border: 1px solid var(--line); border-radius: 12px; padding: 12px 16px; }
.notice.error { border-color: var(--danger); }
.notice p:last-child { margin-bottom: 0; }

/* The vitrine: a shelf lit from above. ::before carries the spotlight so it can fade in. */
.vitrine {
  position: relative; isolation: isolate; border-radius: 14px;
  background: linear-gradient(transparent 40%, var(--shelf)); box-shadow: inset 0 -6px 0 var(--edge);
}
.vitrine::before {
  content: ""; position: absolute; inset: 0; z-index: -1; border-radius: inherit; pointer-events: none;
  background: radial-gradient(80% 70% at 50% 0%, var(--spot), transparent 72%);
}
.shelf {
  list-style: none; margin: 0; padding: 22px 12px 0;
  display: grid; grid-template-columns: repeat(auto-fill, minmax(168px, 1fr)); gap: 8px;
}
.medal { display: grid; justify-items: center; gap: 6px; padding: 0 6px 20px; text-align: center; text-decoration: none; }
.medal-name { font-weight: 700; }
.medal-img, .badge-img, .badge-hero { filter: drop-shadow(0 6px 6px rgba(0, 0, 0, .18)); }
.medal .medal-img { width: 88px; height: 88px; }
.medal-img.placeholder, .badge-img.placeholder, .badge-hero.placeholder {
  display: grid; place-items: center; border-radius: 50%; background: var(--line); color: var(--muted); font-weight: 700; filter: none;
}

/* Result cards: a medal on the shelf; opening one shows its details across the row. */
.badge-card { border-radius: 12px; }
.badge-card summary { list-style: none; cursor: pointer; padding: 6px 6px 18px; display: grid; justify-items: center; gap: 6px; text-align: center; border-radius: 12px; }
.badge-card summary::-webkit-details-marker { display: none; }
.badge-card[open] { grid-column: 1 / -1; background: var(--surface); border: 1px solid var(--line); margin-bottom: 14px; }
.badge-card[open] summary { justify-items: start; text-align: left; grid-template-columns: 96px 1fr; align-items: center; column-gap: 16px; padding: 14px; }
.badge-card[open] summary .badge-img { grid-row: 1 / span 3; }
.badge-card[open] summary > :not(.badge-img) { grid-column: 2; }
.badge-img { width: 96px; height: 96px; }
.badge-img.placeholder { font-size: 1.6rem; }
.locked .badge-img { filter: grayscale(1); opacity: .45; }
.badge-name { font-weight: 700; }
.badge-line { color: var(--muted); font-size: .875rem; }
.badge-detail { padding: 0 16px 16px; overflow-wrap: anywhere; }
.badge-card progress { width: 100%; max-width: 150px; }
.badge-card[open] progress { max-width: 280px; }
progress {
  -webkit-appearance: none; appearance: none; height: 6px; border: 0; border-radius: 3px; overflow: hidden;
  background: color-mix(in srgb, var(--ink) 14%, transparent); color: var(--gold);
}
progress::-webkit-progress-bar { background: color-mix(in srgb, var(--ink) 14%, transparent); border-radius: 3px; }
progress::-webkit-progress-value { background: var(--gold); border-radius: 3px; }
progress::-moz-progress-bar { background: var(--gold); border-radius: 3px; }
.tier {
  display: inline-block; margin-left: 6px; padding: 3px 7px; border-radius: 999px; vertical-align: 2px;
  font-size: .75rem; font-weight: 700; line-height: 1; color: var(--ink); background: var(--surface); box-shadow: inset 0 0 0 1px var(--line);
}
.tier.t2, .tier.t3, .tier.t4 { color: var(--metal-ink); box-shadow: none; }
.tier.t2 { background: var(--bronze); }
.tier.t3 { background: var(--silver); }
.tier.t4 { background: var(--gold); }

/* Result page */
.result-head { display: flex; flex-wrap: wrap; align-items: center; gap: 16px; margin: 24px 0 8px; }
.avatar { width: 72px; height: 72px; border-radius: 50%; box-shadow: 0 0 0 3px var(--surface), 0 0 0 4px var(--line); }
.result-user { margin: 0; font-size: 2.441rem; font-weight: 800; }
.result-sum { margin: 2px 0 0; color: var(--muted); }
.next { margin: 0 0 0 auto; padding: 10px 14px; border-radius: 12px; background: var(--surface); border: 1px solid var(--line); color: var(--muted); font-size: .875rem; }
.next strong { display: block; color: var(--ink); font-size: 1rem; }
.scope { color: var(--muted); font-size: .94rem; }
.shelf-head { margin: 1.8em 0 .2em; }
.archive { margin-top: 24px; padding: 2px 14px; border: 1px dashed var(--line); border-radius: 12px; }
.archive > summary { cursor: pointer; padding: 10px 0; color: var(--muted); font-weight: 600; }
.archive .shelf { margin-bottom: 14px; }
.highlights { list-style: none; padding: 0; display: flex; flex-wrap: wrap; gap: 8px; }
.highlights li { padding: 3px 12px; border-radius: 999px; border: 1px solid var(--line); background: var(--surface); font-size: .875rem; }
.highlights .off { color: var(--muted); background: transparent; border-style: dashed; }
.actions { display: flex; flex-wrap: wrap; gap: 10px; margin: 20px 0 10px; }

/* Star banner: the page's one loud moment after the shelves (result) or the FAQ (home). */
.star-banner {
  display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 12px 20px;
  margin: 32px 0 8px; padding: 18px 22px; border-radius: 14px; background: var(--gold); color: var(--metal-ink);
  text-decoration: none; box-shadow: inset 0 -4px 0 rgba(0, 0, 0, .14);
}
.star-banner strong { font-size: 1.25rem; font-weight: 800; }
.star-cta { display: inline-flex; align-items: center; gap: 8px; padding: 8px 16px; border-radius: 999px; background: var(--metal-ink); color: var(--gold); font-weight: 700; }
.star-cta .star-count { background: rgba(255, 255, 255, .16); }

/* Home */
.hero { padding: 32px 0 8px; }
.hero h1 { font-size: 3.052rem; letter-spacing: -.025em; margin: 0 0 .25em; max-width: 14ch; }
.hero .lede { margin-bottom: 20px; }
.signin { margin: 12px 0 0; color: var(--muted); font-size: .94rem; }
.home-shelf { margin-top: 36px; }
.points { list-style: none; padding: 0; margin: 0; display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: 8px 32px; }
.points h3 { margin: .6em 0 .2em; font-size: 1.1rem; }
.points p { margin: 0; color: var(--muted); }
.install { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; }
.install pre { flex: 1 1 320px; margin: 0; }

/* Check page */
.choice { max-width: 540px; padding: 24px; border-radius: 16px; background: var(--surface); border: 1px solid var(--line); }
.choice .or { margin: 20px 0 8px; color: var(--muted); }

/* All achievements */
.medal-grid { list-style: none; padding: 0; margin: 0; display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 24px 28px; }
.medal-grid li { display: grid; grid-template-columns: 64px 1fr; gap: 14px; align-items: start; }
.medal-grid .medal-img { width: 64px; height: 64px; }
.medal-grid h3 { margin: 4px 0 2px; font-size: 1.1rem; }
.medal-grid p { margin: 0; }

/* Badge page */
.badge-top { display: flex; flex-wrap: wrap; align-items: center; gap: 20px 36px; margin: 8px 0 28px; padding: 24px 24px 30px; }
.badge-hero { width: 160px; height: 160px; }
.badge-hero.placeholder { font-size: 2.6rem; }
.tier-strip { list-style: none; padding: 0; margin: 0; display: flex; flex-wrap: wrap; gap: 16px; }
.tier-strip li { display: grid; justify-items: center; gap: 4px; color: var(--muted); font-size: .82rem; }
.tier-strip .medal-img { width: 64px; height: 64px; }
table { width: 100%; max-width: 520px; border-collapse: collapse; font-size: .94rem; }
th, td { text-align: left; padding: 8px 10px; border-bottom: 1px solid var(--line); vertical-align: top; }
thead th { color: var(--muted); font-weight: 600; }
.facts { display: grid; grid-template-columns: max-content 1fr; gap: 8px 20px; max-width: 760px; }
.facts dt { font-weight: 700; }
.facts dd { margin: 0; }
.faq dt { font-weight: 700; margin-top: 18px; }
.faq dd { margin: 4px 0 0; color: var(--muted); max-width: 68ch; }
.crumbs ol { list-style: none; display: flex; flex-wrap: wrap; gap: 6px; padding: 0; margin: 12px 0 0; font-size: .875rem; color: var(--muted); }
.crumbs a { text-decoration: none; }
.pager { display: flex; justify-content: space-between; flex-wrap: wrap; gap: 12px; margin: 40px 0 8px; }

/* Sign-in callback */
.checking { max-width: 560px; padding: 20px 22px; border-radius: 14px; }
.steps { margin: 6px 0 12px; color: var(--muted); }

#consent {
  position: fixed; left: 16px; right: 16px; bottom: 16px; z-index: 10; max-width: 560px; margin: 0 auto;
  padding: 14px 16px; border-radius: 14px; background: var(--surface); border: 1px solid var(--line);
  box-shadow: 0 10px 30px rgba(0, 0, 0, .25);
}
#consent p { margin: 0 0 10px; }

@media (max-width: 640px) {
  .top-in { flex-wrap: wrap; }
  .js .menu-btn { display: inline-block; }
  .js .top nav:not([data-open]) { display: none; }
  .top nav { order: 5; flex-basis: 100%; flex-direction: column; gap: 0; padding: 4px 0; }
  .top nav a { padding: 8px 0; }
  .star-long { display: none; }
}
@media (max-width: 600px) {
  h1, .result-user { font-size: 1.953rem; }
  .hero h1 { font-size: 2.441rem; }
  .shelf { grid-template-columns: repeat(2, minmax(0, 1fr)); padding: 18px 6px 0; }
  .next { margin-left: 0; }
  .facts { grid-template-columns: 1fr; }
  .facts dd { margin-bottom: 8px; }
  .badge-hero { width: 120px; height: 120px; }
  .badge-card[open] summary { grid-template-columns: 1fr; justify-items: center; text-align: center; }
  .badge-card[open] summary .badge-img { grid-row: auto; }
  .badge-card[open] summary > :not(.badge-img) { grid-column: auto; }
}

/* Motion: the home spotlight fades in once; an opened card grows to its height. */
@media (prefers-reduced-motion: no-preference) {
  :root { interpolate-size: allow-keywords; }
  .home-shelf::before { animation: spot-in 1.4s ease-out both; }
  @keyframes spot-in { from { opacity: 0; } to { opacity: 1; } }
  .badge-card::details-content { height: 0; overflow: clip; transition: height .25s ease, content-visibility .25s allow-discrete; }
  .badge-card[open]::details-content { height: auto; }
}
```

- [ ] **Step 6: Update `internal/webui/templates/layout.html`**

In `<head>`, replace the line `<link rel="stylesheet" href="{{.Base}}style.css">` with these three lines (theme first, then the font preload, then the stylesheet):

```html
<script src="{{.Base}}theme.js"></script>
<link rel="preload" href="{{.Base}}fonts/bricolage-grotesque.woff2" as="font" type="font/woff2" crossorigin>
<link rel="stylesheet" href="{{.Base}}style.css">
```

Replace the whole `<header class="top"> … </header>` block with:

```html
<header class="top">
  <div class="top-in">
    <a class="brand" href="{{.Base}}">🏆 gh-achievements</a>
    <nav id="site-nav" aria-label="Site">
      <a href="{{.Base}}check/">Check yours</a>
      <a href="{{.Base}}achievements/">All achievements</a>
      <a href="{{.Base}}methodology/">How it works</a>
      <a href="{{.RepoURL}}">GitHub</a>
    </nav>
    <button type="button" class="theme-btn" data-theme-toggle hidden>Dark</button>
  </div>
</header>
```

- [ ] **Step 7: Add the theme button to `internal/webui/static/app.js`**

After `var LOGIN = …;` add:

```js
  var THEME = "gha:theme";
  // The theme in effect: an explicit choice (set by theme.js or the button), else the system's.
  function currentTheme() {
    var t = document.documentElement.getAttribute("data-theme");
    if (t === "light" || t === "dark") return t;
    return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  }
  // The button names the theme it switches to.
  function labelTheme(btn) {
    var next = currentTheme() === "dark" ? "light" : "dark";
    btn.textContent = next === "dark" ? "Dark" : "Light";
    btn.setAttribute("aria-label", "Switch to " + next + " theme");
  }
```

In `ready()`, change the click selector to
`"[data-consent],[data-consent-reset],[data-signin],[data-share],[data-copy],[data-theme-toggle]"`
and add this branch before `} else if (t.dataset.copy) {`:

```js
      } else if (t.hasAttribute("data-theme-toggle")) {
        var next = currentTheme() === "dark" ? "light" : "dark";
        document.documentElement.setAttribute("data-theme", next);
        set(THEME, next);
        labelTheme(t);
```

At the start of `ready()` (after `var banner = …;`) add:

```js
    var themeBtn = document.querySelector("[data-theme-toggle]");
    if (themeBtn) {
      labelTheme(themeBtn);
      themeBtn.hidden = false;
      var mq = window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)");
      if (mq && mq.addEventListener) mq.addEventListener("change", function () { labelTheme(themeBtn); });
    }
```

In the image fallback, change `var SIZES = ["badge-img", "badge-icon", "badge-hero"];` to `var SIZES = ["badge-img", "badge-icon", "medal-img", "badge-hero"];` (`medal-img` is the class Tasks 5–6 use; Task 6 drops `badge-icon` once no template uses it).

- [ ] **Step 8: Copy the new static files in `cmd/gendocs/site.go`**

Replace `for _, name := range []string{"style.css", "app.js", "callback.js"} {` with:

```go
	for _, name := range []string{"style.css", "theme.js", "app.js", "callback.js", "fonts/bricolage-grotesque.woff2", "fonts/OFL.txt"} {
```

(`//go:embed templates/*.html static/*` already embeds `static/fonts/` recursively.)

- [ ] **Step 9: Run the tests, regenerate, check**

```bash
go test ./internal/webui ./cmd/gendocs
go test ./web -run TestRenderResult_golden -update
go run ./cmd/gendocs
go run ./cmd/gendocs -check && go test ./... && gofmt -l .
```

Expected: all PASS, `-check` silent, `gofmt` prints nothing. `TestStyle_openCardAndLongLinkRules` still passes (its six pinned rules are in the new stylesheet verbatim).

- [ ] **Step 10: Commit**

```bash
git add internal/webui cmd/gendocs/site.go cmd/gendocs/site_test.go site web/testdata
git commit -m "feat(site): vitrine stylesheet with a light/dark theme"
```

---

### Task 3: Header menu and the star call to action

**Files:**
- Create: `internal/webui/templates/parts.html`
- Modify: `internal/webui/webui.go` (parse `parts.html`, `repoURL` func), `internal/webui/templates/layout.html` (star pill, menu button), `internal/webui/static/app.js` (menu, star count, `star_click`), `web/headers.go` (CSP)
- Test: `internal/webui/webui_test.go`, `web/handlers_test.go`

**Interfaces:**
- Consumes: `.star`, `.star-long`, `.star-count`, `.menu-btn`, `#site-nav`, `.star-banner`, `.star-cta` (Task 2 CSS); `labelTheme` and the click handler (Task 2).
- Produces: template `{{template "star-banner" "<place>"}}` (place ∈ `result|home`), available in every page and in `RenderFragment`; template func `repoURL`; `window.gha.fillStars()` (fills every `[data-stars]` on the page from the cached count; Task 4 calls it after injecting the signed-in fragment); markup contract `data-star="header|result|home"` on star links, `data-stars` on count spans.

- [ ] **Step 1: Write the failing tests**

In `internal/webui/webui_test.go`, add to the `want` list of `TestRender_usesTheVercelDomainAndRootPaths`:

```go
		`<a class="star" href="https://github.com/nmloc/gh-achievements" rel="noopener" data-star="header">`,
		`<span class="star-long"> on GitHub</span>`,
		`<span class="star-count" data-stars hidden></span>`,
		`<button type="button" class="menu-btn" data-menu aria-expanded="false" aria-controls="site-nav">Menu</button>`,
		`<a href="/methodology/">How it works</a>`,
```

and add:

```go
func TestStarBanner_namesItsPlaceAndTheRepo(t *testing.T) {
	b, err := RenderFragment("star-banner", "home")
	if err != nil {
		t.Fatal(err)
	}
	html := string(b)
	for _, want := range []string{`<a class="star-banner" href="https://github.com/nmloc/gh-achievements" rel="noopener" data-star="home">`,
		"Found this useful? Star gh-achievements on GitHub", `data-stars hidden`} {
		if !strings.Contains(html, want) {
			t.Errorf("star banner lacks %s", want)
		}
	}
}

// The star count is fetched anonymously, without a referrer (the result URL names a user),
// cached for the visit, and only the place of a click reaches GA.
func TestAppJS_starCountAndClick(t *testing.T) {
	js, err := Static("app.js")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"https://api.github.com/repos/nmloc/gh-achievements"`, `credentials: "omit"`,
		`referrerPolicy: "no-referrer"`, `sessionStorage`, `track("star_click", { place: t.dataset.star })`,
		`fillStars: fillStars`, `[data-menu]`, `aria-expanded`} {
		if !strings.Contains(string(js), want) {
			t.Errorf("app.js lacks %s", want)
		}
	}
}
```

In `web/handlers_test.go`, add `"connect-src 'self' https://api.github.com"` to the `want` list of `TestSecurityHeaders_cspAllowsWhatThePagesLoad`.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/webui ./web`
Expected: FAIL (`missing <a class="star" …`, `template: no template "star-banner"`, `CSP lacks connect-src 'self' https://api.github.com`).

- [ ] **Step 3: Create `internal/webui/templates/parts.html`**

```html
{{define "star-banner"}}
<a class="star-banner" href="{{repoURL}}" rel="noopener" data-star="{{.}}">
  <strong>Found this useful? Star gh-achievements on GitHub</strong>
  <span class="star-cta"><span aria-hidden="true">★</span> Star <span class="star-count" data-stars hidden></span></span>
</a>
{{end}}
```

- [ ] **Step 4: Parse it everywhere in `internal/webui/webui.go`**

In `parse`, change the FuncMap to `template.FuncMap{"rel": Rel, "repoURL": func() string { return RepoURL }}`.
In `Render`, change `files := []string{"layout.html", p.Content, "result.html"}` to `files := []string{"layout.html", p.Content, "parts.html", "result.html"}` and extend the comment above it: `parts.html holds snippets several pages share.`
In `RenderFragment`, change `parse("result.html")` to `parse("result.html", "parts.html")`, and its doc comment to `// RenderFragment executes one named template from result.html or parts.html without the layout.`

- [ ] **Step 5: Header markup in `layout.html`**

Between `</nav>` and the theme button add:

```html
    <a class="star" href="{{.RepoURL}}" rel="noopener" data-star="header"><span aria-hidden="true">★</span> Star<span class="star-long"> on GitHub</span> <span class="star-count" data-stars hidden></span></a>
```

After the theme button add:

```html
    <button type="button" class="menu-btn" data-menu aria-expanded="false" aria-controls="site-nav">Menu</button>
```

- [ ] **Step 6: Menu, star count and `star_click` in `app.js`**

After the `labelTheme` function add:

```js
  // Star count for the repository, fetched once per visit. Anonymous and without a referrer:
  // /result URLs name a user. On any failure the buttons keep their plain label.
  var STARS = "gha:stars", REPO_API = "https://api.github.com/repos/nmloc/gh-achievements";
  var stars = null;
  function fillStars() {
    if (stars === null) return;
    var text = stars.toLocaleString("en-US");
    Array.prototype.forEach.call(document.querySelectorAll("[data-stars]"), function (el) {
      el.textContent = text;
      el.hidden = false;
    });
    Array.prototype.forEach.call(document.querySelectorAll(".star-long"), function (el) { el.hidden = true; });
  }
  function loadStars() {
    var cached = null;
    try { cached = sessionStorage.getItem(STARS); } catch (e) { /* storage blocked: fetch instead */ }
    if (cached !== null && /^\d+$/.test(cached)) {
      stars = Number(cached);
      fillStars();
      return;
    }
    if (!window.fetch || !document.querySelector("[data-stars]")) return;
    fetch(REPO_API, { credentials: "omit", referrerPolicy: "no-referrer" })
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (d) {
        if (!d || typeof d.stargazers_count !== "number") return;
        stars = d.stargazers_count;
        try { sessionStorage.setItem(STARS, String(stars)); } catch (e) { /* not cached: fetched again on the next page */ }
        fillStars();
      })
      .catch(function () { /* offline or rate-limited: the label stays "Star on GitHub" */ });
  }
```

Change `window.gha = { track: track, reportView: reportView };` to
`window.gha = { track: track, reportView: reportView, fillStars: fillStars };`
(move that line below the `fillStars` definition if it now comes first).

Change the click selector to
`"[data-consent],[data-consent-reset],[data-signin],[data-share],[data-copy],[data-theme-toggle],[data-menu],[data-star]"`
and add these branches before `} else if (t.dataset.copy) {`:

```js
      } else if (t.hasAttribute("data-menu")) {
        var nav = document.getElementById("site-nav");
        var open = t.getAttribute("aria-expanded") !== "true";
        t.setAttribute("aria-expanded", String(open));
        if (nav && open) nav.setAttribute("data-open", "");
        else if (nav) nav.removeAttribute("data-open");
      } else if (t.dataset.star) {
        track("star_click", { place: t.dataset.star });
```

At the end of `ready()`, after `reportView(…);`, add `loadStars();`.

- [ ] **Step 7: CSP in `web/headers.go`**

Change the `connect-src` line to:

```go
			"connect-src 'self' https://api.github.com https://*.google-analytics.com https://*.analytics.google.com https://www.googletagmanager.com; " +
```

- [ ] **Step 8: Run the tests, regenerate, check**

```bash
go test ./web -run TestRenderResult_golden -update
go run ./cmd/gendocs
go run ./cmd/gendocs -check && go test ./... && gofmt -l .
```

Expected: PASS; `vercel.json` now carries the new CSP (its test compares it to `web.SecurityHeaders`).

- [ ] **Step 9: Commit**

```bash
git add internal/webui web/headers.go web/handlers_test.go web/testdata site vercel.json
git commit -m "feat(site): star the repo from the header and a banner"
```

---

### Task 4: Result page as two shelves and an archive

**Files:**
- Modify: `web/view.go` (`ResultView`: `Display`, `Collect`, `Archive`, `Standalone`; drop `Earnable`, `Other`, `QuickWins`), `web/render.go` (`RenderResultPage`), `internal/webui/templates/result.html`, `internal/webui/templates/layout.html` (optional `<h1>`), `internal/webui/static/callback.js`
- Test: `web/view_test.go`, `web/render_test.go`; regenerate `web/testdata/*`

**Interfaces:**
- Consumes: `Card` and its `Line` (Task 1), `star-banner` template and `window.gha.fillStars` (Task 3), result CSS classes (Task 2).
- Produces: `ResultView.Display`, `.Collect`, `.Archive []Card`; `ResultView.Standalone bool`; layout prints `<h1>` only when `Page.Heading != ""` (Task 5 relies on this for the home page).

- [ ] **Step 1: Write the failing tests**

In `web/view_test.go`, replace `TestResultView_quickWinsAreUnearnedSingleTierBadges` and `TestResultView_hiddenProfileHasNoQuickWins` with:

```go
func slugs(cards []Card) string {
	var s []string
	for _, c := range cards {
		s = append(s, c.Slug)
	}
	return strings.Join(s, ",")
}

func TestResultView_shelves(t *testing.T) {
	v := demoView(t, false)
	if got := slugs(v.Display); got != "pull-shark,starstruck,pair-extraordinaire,arctic-code-vault-contributor,mars-2020-contributor" {
		t.Errorf("on display = %s", got)
	}
	if got := slugs(v.Collect); got != "galaxy-brain,quickdraw,yolo,public-sponsor,finishing-touch" {
		t.Errorf("still to collect = %s", got)
	}
	if got := slugs(v.Archive); got != "heart-on-your-sleeve,open-sourcerer,proxima-pioneer,proxima-staffshipper,proxima-staffuser" {
		t.Errorf("archive = %s", got)
	}
}

func TestResultView_hiddenProfileHasNothingOnDisplay(t *testing.T) {
	r, err := report.Build("sindresorhus", fixtures.Official{}, &fixtures.Counts{Authed: true,
		PRs: counters.MergedPRs{Total: 2241}, FactsValue: &counters.Facts{}})
	if err != nil {
		t.Fatal(err)
	}
	v := NewResultView(r, false)
	if !v.Hidden || len(v.Display) != 0 || len(v.Collect) != 8 {
		t.Fatalf("hidden %v, display %s, collect %s", v.Hidden, slugs(v.Display), slugs(v.Collect))
	}
}
```

In `TestResultView_cards`, change `card(t, v.Earnable, "pull-shark")` to `card(t, v.Display, "pull-shark")`, `card(t, v.Earnable, "galaxy-brain")` to `card(t, v.Collect, "galaxy-brain")`, `card(t, v.Other, "mars-2020-contributor")` to `card(t, v.Display, "mars-2020-contributor")`, and `card(t, v.Other, "heart-on-your-sleeve")` to `card(t, v.Archive, "heart-on-your-sleeve")`. In `TestCard_nonEarnableKindsNeverShowProgress`, `card(t, v.Other, "heart-on-your-sleeve")` becomes `card(t, v.Archive, "heart-on-your-sleeve")`.

In `web/render_test.go`, replace `TestRenderResult_contract` with:

```go
func TestRenderResult_contract(t *testing.T) {
	pub, _ := RenderResultPage(demoView(t, false))
	priv, _ := RenderResultFragment(demoView(t, true))
	p, s := string(pub), string(priv)
	for _, want := range []string{`data-report data-scope="public" data-earned="3"`, `data-share`, `href="/auth/login"`,
		`<details class="badge-card" data-badge="pull-shark">`, `<progress value="56" max="112"`, `Closest next unlock`,
		`https://github.githubassets.com/assets/pull-shark-bronze-`, `<h1 class="result-user">@torvalds</h1>`,
		`3 of 8 badges you can still earn, 12 tiers`, `Showing public data.`, `<h2 class="shelf-head">On display</h2>`,
		`<h2 class="shelf-head">Still to collect</h2>`, `The archive: 5 badges you can no longer earn`,
		`data-star="result"`, `>Copy share link</button>`, `First unlocked 2020-02-02.`} {
		if !strings.Contains(p, want) {
			t.Errorf("public page lacks %s", want)
		}
	}
	if n := strings.Count(p, "<h1"); n != 1 {
		t.Errorf("result page has %d <h1>, want 1", n)
	}
	if strings.Index(p, `data-star="result"`) > strings.Index(p, `data-share`) {
		t.Error("the star banner must come before the share and check-another actions")
	}
	if !strings.Contains(s, `data-scope="private"`) || strings.Contains(s, "data-share") || !strings.Contains(s, "not shareable") {
		t.Error("private fragment must say it is not shareable and offer no share button")
	}
	if strings.Contains(s, "<html") || strings.Contains(s, "<h1") || !strings.Contains(s, `<h2 class="result-user">@torvalds</h2>`) {
		t.Error("the private result is a fragment under the callback page's h1")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./web`
Expected: FAIL to compile (`v.Display undefined`).

- [ ] **Step 3: The model in `web/view.go`**

Replace the `ResultView` struct with:

```go
// ResultView is everything result.html shows.
type ResultView struct {
	User, Avatar                     string
	Private, Hidden, Authenticated   bool
	Standalone                       bool // the full /result page: @login is its <h1>
	Earned, EarnableTotal, TierTotal int
	Closest                          *Card
	Display                          []Card // "On display": every earned badge, any kind
	Collect                          []Card // "Still to collect": earnable, not earned
	Archive                          []Card // "The archive": cannot be earned, not earned
	Highlights                       []report.HighlightRow
	Verified                         string
}
```

Replace the loop in `NewResultView` with:

```go
	bestRemaining := 0
	for _, row := range r.Achievements {
		c := newCard(row)
		if row.Status == report.Earned {
			v.TierTotal += row.Tier
		}
		earnable := row.Kind == catalog.Earnable
		switch {
		case c.Earned:
			v.Display = append(v.Display, c)
		case earnable:
			v.Collect = append(v.Collect, c)
		default:
			v.Archive = append(v.Archive, c)
		}
		if !earnable {
			continue
		}
		v.EarnableTotal++
		if c.Earned {
			v.Earned++
		}
		if st := row.Standing; st != nil && !st.IsMax && st.Remaining > 0 &&
			(v.Closest == nil || st.Remaining < bestRemaining) {
			cc := c
			v.Closest = &cc
			bestRemaining = st.Remaining
		}
	}
	return v
```

- [ ] **Step 4: `web/render.go`**

In `RenderResultPage`, set the flag on a copy and drop the layout heading:

```go
// RenderResultPage is the full /result page. Its <h1> is the @login in the result header.
func RenderResultPage(v ResultView) ([]byte, error) {
	v.Standalone = true
	return webui.Render(webui.Page{
		Path: "result", URL: webui.SiteURL + "result", Title: "@" + v.User + "'s GitHub achievements · gh-achievements",
		Description: "Every GitHub achievement, @" + v.User + "'s tier and how far to the next one.",
		Content:     "result-page.html", NoIndex: true, Data: v,
	})
}
```

- [ ] **Step 5: `layout.html` prints `<h1>` only when there is a heading**

Replace `<h1>{{.P.Heading}}</h1>` with `{{with .P.Heading}}<h1>{{.}}</h1>{{end}}`.

- [ ] **Step 6: Replace `internal/webui/templates/result.html` with**

```html
{{define "result"}}
<section class="result" data-report data-scope="{{if .Private}}private{{else}}public{{end}}" data-earned="{{.Earned}}" data-user="{{.User}}">
  <header class="result-head">
    <img class="avatar" src="{{.Avatar}}" width="72" height="72" alt="">
    <div>
      {{if .Standalone}}<h1 class="result-user">@{{.User}}</h1>{{else}}<h2 class="result-user">@{{.User}}</h2>{{end}}
      <p class="result-sum">{{.Earned}} of {{.EarnableTotal}} badges you can still earn, {{.TierTotal}} tiers</p>
    </div>
    {{with .Closest}}<p class="next">Closest next unlock <strong>{{.Name}}: {{.Line}}</strong></p>{{end}}
  </header>
  {{if .Private}}
  <p class="scope">Includes private work, so this page is not shareable. Organizations that restrict third-party apps hide their private repositories until an admin approves gh-achievements.</p>
  {{else}}
  <p class="scope">Showing public data. <a href="/auth/login" data-signin>Sign in with GitHub</a> to include private repositories.</p>
  {{end}}
  {{if .Hidden}}<p class="notice">This profile hides its achievements, so there are no official tiers: live counts only.</p>{{end}}

  {{if .Display}}
  <h2 class="shelf-head">On display</h2>
  <p class="hint">Badges you have, with how far the next tier is.</p>
  <div class="vitrine shelf">{{range .Display}}{{template "card" .}}{{end}}</div>
  {{else if not .Hidden}}
  <h2 class="shelf-head">On display</h2>
  <p class="hint">Nothing here yet. Start with a badge from the shelf below.</p>
  {{end}}

  {{with .Collect}}
  <h2 class="shelf-head">Still to collect</h2>
  <p class="hint">Badges you can still earn. Open one to see how.</p>
  <div class="vitrine shelf">{{range .}}{{template "card" .}}{{end}}</div>
  {{end}}

  {{with .Archive}}
  <details class="archive">
    <summary>The archive: {{len .}} badges you can no longer earn (switched off by GitHub, retired or staff only)</summary>
    <div class="vitrine shelf">{{range .}}{{template "card" .}}{{end}}</div>
  </details>
  {{end}}

  <h2 class="shelf-head">Highlights</h2>
  <ul class="highlights">{{range .Highlights}}<li class="{{if .Earned}}on{{else}}off{{end}}">{{if .Earned}}✓{{else}}·{{end}} {{.Name}}</li>{{end}}</ul>

  {{template "star-banner" "result"}}

  <p class="actions">
    {{if .Private}}<a href="/auth/login" data-signin class="btn">Refresh</a>
    {{else}}<button type="button" class="btn gold" data-share>Copy share link</button>{{end}}
    <a href="/check/" class="btn">Check another profile</a>
  </p>
  <p class="meta">official = GitHub's own pages · calibrated = counted, checked against GitHub · counted = counted only. Want every private count? Use the CLI: <code>gh extension install nmloc/gh-achievements</code> <button type="button" class="btn small" data-copy="gh extension install nmloc/gh-achievements">Copy</button></p>
</section>
{{end}}

{{define "card"}}
<details class="badge-card{{if .Locked}} locked{{end}}" data-badge="{{.Slug}}">
  <summary>
    {{if .Image}}<img class="badge-img" src="{{.Image}}" width="96" height="96" alt="{{.ImageAlt}}" loading="lazy">{{else}}<span class="badge-img placeholder" aria-hidden="true">{{.Initials}}</span>{{end}}
    <span class="badge-name">{{.Name}}{{if .TierLabel}} <span class="tier t{{.Tier}}">{{.TierLabel}}</span>{{end}}</span>
    {{with .Progress}}<progress value="{{.Value}}" max="{{.Max}}" aria-label="{{.Label}}">{{.Value}} of {{.Max}}</progress>{{end}}
    <span class="badge-line">{{.Line}}</span>
  </summary>
  <div class="badge-detail">
    <p>{{.HowToEarn}}.{{if .TierText}} Tiers: {{.TierText}}.{{end}}{{if .Unlocked}} First unlocked {{.Unlocked}}.{{end}}</p>
    {{if .Source}}<p class="meta"><strong>{{.Source}}</strong>: {{.SourceMeaning}}</p>{{end}}
    {{if .Notes}}<ul>{{range .Notes}}<li>{{.}}</li>{{end}}</ul>{{end}}
    {{if .Milestones}}<ol class="milestones">{{range .Milestones}}<li>{{if .URL}}<a href="{{.URL}}">{{.Label}}</a>{{else}}{{.Label}}{{end}}</li>{{end}}</ol>{{end}}
    {{if .BadgeURL}}<p><a href="{{.BadgeURL}}">About {{.Name}}</a></p>{{end}}
  </div>
</details>
{{end}}
```

(The source legend keeps its middle dots: it is the existing legend, a list of label = meaning pairs, not decoration.)

- [ ] **Step 7: Fill the star count in the injected fragment (`callback.js`)**

In `render(html)`, inside `if (window.gha) { … }`, after `window.gha.reportView(…);` add:

```js
      window.gha.fillStars();
```

- [ ] **Step 8: Run the tests, regenerate, check**

```bash
go test ./web -run 'TestResultView_|TestCard_'
go test ./web -run TestRenderResult_golden -update
go run ./cmd/gendocs
go run ./cmd/gendocs -check && go test ./... && gofmt -l .
```

Expected: PASS. `grep -c 'class="vitrine shelf"' web/testdata/result-public.html` prints `3`.

- [ ] **Step 9: Commit**

```bash
git add web internal/webui site
git commit -m "feat(web): result page as shelves of medals"
```

---

### Task 5: Home hero, home shelf and the check page

**Files:**
- Modify: `internal/webui/templates/home.html`, `internal/webui/templates/check.html`, `internal/webui/templates/parts.html` (add `checkform`), `cmd/gendocs/site.go` (home page data and heading), `cmd/gendocs/main.go` (stop generating `site/demo.svg`)
- Delete: `site/demo.svg`
- Test: `cmd/gendocs/site_test.go`

**Interfaces:**
- Consumes: optional layout `<h1>` (Task 4), `star-banner` (Task 3), `.hero`, `.signin`, `.home-shelf`, `.medal*`, `.points`, `.install`, `.choice` (Task 2).
- Produces: template `{{template "checkform" "<prefilled username>"}}` in `parts.html`.

- [ ] **Step 1: Write the failing tests**

Add to `cmd/gendocs/site_test.go`:

```go
func TestSitePages_homeIsATrophyCase(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	home := string(pages["site/index.html"])
	for _, want := range []string{`<h1>Your GitHub trophy case</h1>`, `<form action="/result" method="get" id="check-form">`,
		`<a href="/auth/login" data-signin>Sign in with GitHub</a> to include private work.`,
		`<ul class="vitrine shelf home-shelf"`, `data-star="home"`, `data-copy="gh extension install nmloc/gh-achievements"`,
		`"@type":"FAQPage"`, `"@type":"SoftwareApplication"`} {
		if !strings.Contains(home, want) {
			t.Errorf("home lacks %s", want)
		}
	}
	for _, a := range catalog.Achievements {
		link := `<a class="medal" href="achievements/` + a.Slug + `/">`
		if got := strings.Contains(home, link); got != (a.Kind == catalog.Earnable) {
			t.Errorf("%s on the home shelf = %v, want %v", a.Slug, got, a.Kind == catalog.Earnable)
		}
	}
	if strings.Index(home, `data-star="home"`) < strings.Index(home, `<dl class="faq">`) {
		t.Error("the home star banner belongs under the FAQ")
	}
	if strings.Contains(home, "demo.svg") {
		t.Error("the home page no longer shows the terminal demo")
	}
}

func TestSitePages_everyPageHasOneH1(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	for path, body := range pages {
		if strings.HasSuffix(path, ".html") {
			if n := strings.Count(string(body), "<h1"); n != 1 {
				t.Errorf("%s has %d <h1>", path, n)
			}
		}
	}
}
```

Add `"github.com/nmloc/gh-achievements/internal/catalog"` to the test file's imports. (`site/demo.svg` is written by `generate()`, which reads `README.md` from the repo root and cannot run from the test's working directory; Step 6's `grep` and `gendocs -check` cover it.)

The home page now holds the username form, whose `placeholder="GitHub username"` attribute would trip `TestSitePages_everyBadgeHasArt`'s plain `"placeholder"` match. Make that check look for the placeholder **class** instead, in the existing loop over `site/achievements/index.html` and `site/index.html`:

```go
		if strings.Contains(string(pages[path]), ` placeholder"`) {
			t.Errorf("%s still renders an empty placeholder", path)
		}
```

Confirm the narrowed check can still fail: copy `list.html` aside, change its `{{if .Image}}` to `{{if false}}`, run `go test ./cmd/gendocs -run TestSitePages_everyBadgeHasArt` (red), then restore the copy with `cp` (never `git checkout`, which would also drop uncommitted work).

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./cmd/gendocs -run 'TestSitePages_'`
Expected: FAIL (`home lacks <h1>Your GitHub trophy case</h1>`).

- [ ] **Step 3: Add `checkform` to `parts.html`**

```html
{{define "checkform"}}
<form action="/result" method="get" id="check-form">
  <label for="u" class="visually-hidden">GitHub username</label>
  <input id="u" name="u" required autocomplete="username" autocapitalize="off" spellcheck="false"
    pattern="[A-Za-z0-9][A-Za-z0-9\-]{0,38}" placeholder="GitHub username" value="{{.}}">
  <button type="submit" class="btn primary">Check</button>
</form>
{{end}}
```

- [ ] **Step 4: Replace `internal/webui/templates/check.html` with**

```html
{{define "content"}}
{{$user := ""}}{{with .}}{{with .Message}}<p class="notice error" role="alert">{{.}}</p>{{end}}{{$user = .User}}{{end}}
<div class="choice">
  <p class="lede">Every GitHub achievement, your tier, and how far you are from the next one.</p>
  <a href="/auth/login" data-signin class="btn primary big">Sign in with GitHub</a>
  <p class="meta">Includes private repositories. GitHub asks for repository access; the token is used for one check, then revoked. Nothing is stored.</p>
  <p class="or">or check any public profile</p>
  {{template "checkform" $user}}
</div>
<p class="meta">Prefer the terminal? <code>{{"gh extension install nmloc/gh-achievements"}}</code> <button type="button" class="btn small" data-copy="gh extension install nmloc/gh-achievements">Copy</button> counts everything your gh login can see.</p>
{{end}}
```

- [ ] **Step 5: Replace `internal/webui/templates/home.html` with**

```html
{{define "content"}}
<section class="hero">
  <h1>Your GitHub trophy case</h1>
  <p class="lede">See every GitHub achievement, the tier you hold and how far the next one is, checked against GitHub's own data.</p>
  {{template "checkform" ""}}
  <p class="signin"><a href="/auth/login" data-signin>Sign in with GitHub</a> to include private work.</p>
</section>

<ul class="vitrine shelf home-shelf" aria-label="The {{len .Earnable}} badges you can earn, at their top tier">
{{range .Earnable}}<li><a class="medal" href="achievements/{{.Slug}}/"><img class="medal-img" src="{{.Image}}" width="88" height="88" alt=""><span class="medal-name">{{.Name}}</span></a></li>
{{end}}</ul>

<h2>What it shows</h2>
<ul class="points">
  <li><h3>Every badge, not just yours</h3><p>All {{.Count}} known achievements ({{.Summary}}) and {{.HighlightCount}} profile highlights.</p></li>
  <li><h3>How far to the next tier</h3><p>Your count against the next threshold, such as 72 / 128 merged pull requests for Pull Shark x3.</p></li>
  <li><h3>Checked against GitHub</h3><p>Tiers, unlock dates and milestones come from GitHub's own badge history. <a href="methodology/">How it works</a></p></li>
</ul>

<h2>Prefer the terminal?</h2>
<div class="install">
  <pre><code>{{.Install}}
gh achievements</code></pre>
  <button type="button" class="btn small" data-copy="{{.Install}}">Copy</button>
</div>
<p class="meta">Needs the <a href="https://cli.github.com/">GitHub CLI</a> and reuses your <code>gh auth login</code>, so it counts everything your login can see. Add a username to check anyone: <code>gh achievements torvalds</code>. The web checker uses your GitHub sign-in for one check, then revokes the token.</p>

<h2>FAQ</h2>
<dl class="faq">
{{range .FAQ}}<dt>{{.Q}}</dt><dd>{{.A}}</dd>
{{end}}</dl>

{{template "star-banner" "home"}}
{{end}}
```

- [ ] **Step 6: Home data in `cmd/gendocs/site.go` and the demo in `main.go`**

In `sitePages`, after the `views` loop, add:

```go
	var earnable []badgeView
	for _, v := range views {
		if v.Kind == catalog.Earnable {
			earnable = append(earnable, v)
		}
	}
```

In the home page entry, change `Heading: "GitHub Achievements Checker",` to `Heading: "", // the hero holds the <h1>` and replace its `Data` with:

```go
			Data: map[string]interface{}{"Earnable": earnable, "FAQ": siteFAQ(), "Install": webui.Install,
				"Summary": kindSummary(), "Count": len(catalog.Achievements), "HighlightCount": len(catalog.Highlights)},
```

In `cmd/gendocs/main.go`, delete the line `files["site/demo.svg"] = svg`. Then:

```bash
git rm site/demo.svg
grep -rn "demo.svg" site internal cmd | grep -v "docs/demo.svg"
```

Expected: the `grep` prints nothing (`docs/demo.svg`, used by the README, stays).

- [ ] **Step 7: Run the tests, regenerate, check**

```bash
go test ./cmd/gendocs
go test ./web -run TestRenderResult_golden -update
go run ./cmd/gendocs
go run ./cmd/gendocs -check && go test ./... && gofmt -l .
```

Expected: PASS (the error page rendered by `web.RenderErrorPage` still pre-fills the username: `TestRender…` tests in `web` cover it).

- [ ] **Step 8: Commit**

```bash
git add internal/webui cmd/gendocs site web/testdata
git commit -m "feat(site): home page as a trophy case"
```

---

### Task 6: All achievements, badge pages, How it works, sign-in

**Files:**
- Modify: `internal/webui/templates/list.html`, `badge.html`, `methodology.html`, `callback.html`; `cmd/gendocs/site.go` (`badgeGroup`, `listGroups`, `tierImage`, `badgeView.TierImages`); `CHANGELOG.md`
- Test: `cmd/gendocs/site_test.go`

**Interfaces:**
- Consumes: `.medal-grid`, `.medal-img`, `.vitrine`, `.badge-top`, `.badge-hero`, `.tier-strip`, `.prose`, `.checking`, `.steps` (Task 2).
- Produces: nothing later tasks use.

- [ ] **Step 1: Write the failing tests**

Add to `cmd/gendocs/site_test.go`:

```go
func TestSitePages_listGroupsBadgesByKind(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	list := string(pages["site/achievements/index.html"])
	last := -1
	for _, h := range []string{"<h2>Earnable</h2>", "<h2>Switched off by GitHub</h2>", "<h2>Retired</h2>", "<h2>GitHub staff only</h2>"} {
		i := strings.Index(list, h)
		if i < last {
			t.Errorf("%s missing or out of order", h)
		}
		last = i
	}
	for _, a := range catalog.Achievements {
		if !strings.Contains(list, `<h3><a href="`+a.Slug+`/">`+a.Name+`</a></h3>`) {
			t.Errorf("list lacks a medal for %s", a.Slug)
		}
	}
	if !strings.Contains(list, `<ul class="medal-grid">`) || strings.Contains(list, `<table class="badges">`) {
		t.Error("the list is a medal grid now, not a table")
	}
}

func TestSitePages_badgePageShowsEveryTierMedal(t *testing.T) {
	pages, err := sitePages()
	if err != nil {
		t.Fatal(err)
	}
	ps := string(pages["site/achievements/pull-shark/index.html"])
	for i := 1; i <= 4; i++ {
		if !strings.Contains(ps, `<img class="medal-img" src="`+catalog.Image("pull-shark", i)+`"`) {
			t.Errorf("pull shark page lacks the x%d medal", i)
		}
	}
	for _, want := range []string{`<div class="vitrine badge-top">`, `<ul class="tier-strip" aria-label="Tiers">`, "x4 gold"} {
		if !strings.Contains(ps, want) {
			t.Errorf("pull shark page lacks %s", want)
		}
	}
	if yolo := string(pages["site/achievements/yolo/index.html"]); strings.Contains(yolo, "tier-strip") {
		t.Error("a single-tier badge has no tier strip")
	}
	if m := string(pages["site/methodology/index.html"]); !strings.Contains(m, `<div class="prose">`) {
		t.Error("methodology is not in a readable column")
	}
	if cb := string(pages["site/auth/callback/index.html"]); !strings.Contains(cb, `<div id="status" class="notice checking" role="status" aria-live="polite">`) {
		t.Error("callback status card lost its id, role or style")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./cmd/gendocs -run 'TestSitePages_(list|badge)'`
Expected: FAIL (`<h2>Earnable</h2> missing or out of order`).

- [ ] **Step 3: Models in `cmd/gendocs/site.go`**

Add `TierImages []tierImage` to `badgeView` (after `Initials string`), and below `tierRow`:

```go
type tierImage struct{ Label, Image string }

// metals names GitHub's tier art: x1 base, x2 bronze, x3 silver, x4 gold.
var metals = []string{"base", "bronze", "silver", "gold"}

type badgeGroup struct {
	Heading, Note string
	Badges        []badgeView
}

// listGroups sorts the catalog into the list page's sections, keeping catalog order.
func listGroups(views []badgeView) []badgeGroup {
	groups := []badgeGroup{
		{Heading: "Earnable", Note: "You can still earn these."},
		{Heading: "Switched off by GitHub", Note: "GitHub tested these and switched them off; nobody can earn them now."},
		{Heading: "Retired", Note: "No longer awarded. Profiles that have them keep them."},
		{Heading: "GitHub staff only", Note: "Only GitHub staff could earn these."},
	}
	kinds := []catalog.Kind{catalog.Earnable, catalog.Disabled, catalog.Retired, catalog.Internal}
	for _, v := range views {
		for i, k := range kinds {
			if v.Kind == k {
				groups[i].Badges = append(groups[i].Badges, v)
			}
		}
	}
	return groups
}
```

In `view(a)`, inside `if a.IsTiered() { … }` after the `TierRows` loop, add:

```go
		for i := 0; i < min(len(a.Tiers), len(catalog.Images[a.Slug]), len(metals)); i++ {
			v.TierImages = append(v.TierImages, tierImage{
				Label: fmt.Sprintf("x%d %s", i+1, metals[i]), Image: catalog.Image(a.Slug, i+1)})
		}
```

In the list page entry, replace `"Badges": views,` with `"Groups": listGroups(views),`.

- [ ] **Step 4: Replace `internal/webui/templates/list.html` with**

```html
{{define "content"}}
<p class="lede">GitHub publishes no official list of profile achievements any more. These are all the known ones ({{.Summary}}), with their tier thresholds and whether you can still earn them.</p>
{{range .Groups}}
<h2>{{.Heading}}</h2>
<p class="hint">{{.Note}}</p>
<ul class="medal-grid">
{{range .Badges}}<li>{{if .Image}}<img class="medal-img" src="{{.Image}}" width="64" height="64" alt="" loading="lazy">{{else}}<span class="medal-img placeholder" aria-hidden="true">{{.Initials}}</span>{{end}}
  <div><h3><a href="{{.Slug}}/">{{.Name}}</a></h3><p>{{.HowToEarn}}.</p><p class="meta">{{if .IsTiered}}Tiers: {{.TierText}}{{else}}One tier{{end}}</p></div></li>
{{end}}</ul>
{{end}}

<h2>Profile highlights</h2>
<p>Highlights appear below the achievements on a profile. They are not tiered.</p>
<ul>
{{range .Highlights}}<li><strong>{{.Name}}</strong>: {{.HowToEarn}}.</li>
{{end}}</ul>

<h2>Check your own progress</h2>
<pre><code>{{.Install}}
gh achievements</code></pre>
<p class="meta">Verified against live profiles on <time datetime="{{.Verified}}">{{.Verified}}</time>. Also available as <a href="../achievements.json">achievements.json</a>.</p>
{{end}}
```

- [ ] **Step 5: `badge.html`**

Replace its first line after `{{define "content"}}{{with .B}}` (the `badge-hero` line) with:

```html
<div class="vitrine badge-top">
  {{if .Image}}<img class="badge-hero" src="{{.Image}}" width="160" height="160" alt="{{.Name}} GitHub achievement badge">{{else}}<span class="badge-hero placeholder" aria-hidden="true">{{.Initials}}</span>{{end}}
  {{with .TierImages}}<ul class="tier-strip" aria-label="Tiers">{{range .}}<li><img class="medal-img" src="{{.Image}}" width="64" height="64" alt="" loading="lazy">{{.Label}}</li>{{end}}</ul>{{end}}
</div>
```

and wrap `<p class="lede">{{.Details}}</p>` as `<p class="lede prose">{{.Details}}</p>`.

- [ ] **Step 6: `methodology.html` and `callback.html`**

`methodology.html` becomes:

```html
{{define "content"}}
<div class="prose">
{{range .Paras}}
<h2>{{.Heading}}</h2>
<p>{{.Text}}</p>
{{end}}
<p class="meta">Last verified <time datetime="{{.Verified}}">{{.Verified}}</time>. Found something off? <a href="https://github.com/nmloc/gh-achievements/issues/new?template=count-looks-wrong.yml">Report it</a>.</p>
</div>
{{end}}
```

In `callback.html`, change `<div id="status" class="notice" role="status" aria-live="polite">` to `<div id="status" class="notice checking" role="status" aria-live="polite">`. Nothing else changes there: the ids, `<noscript>` and error blocks are what `callback.js` and its test rely on.

In `app.js`, `var SIZES = ["badge-img", "badge-icon", "medal-img", "badge-hero"];` becomes `var SIZES = ["badge-img", "medal-img", "badge-hero"];`, and `grep -rn "badge-icon" internal cmd` prints nothing.

- [ ] **Step 7: CHANGELOG**

Insert above `## v0.2.0`:

```markdown
## Unreleased

- New look for the site: your badges stand on a lit shelf, with a light and a dark theme (follows your system, switch in the header).
- Progress reads as count against the next threshold, e.g. `72 / 128 to x3`.
- A star button for the repository in the header and after your results.
```

- [ ] **Step 8: Run the tests, regenerate, check**

```bash
go test ./cmd/gendocs
go run ./cmd/gendocs
go run ./cmd/gendocs -check && go test ./... && gofmt -l .
```

Expected: PASS, including `TestSitePages_everyBadgeHasArt` (the `badge-hero` markup and the absence of `placeholder` on the list and home pages are unchanged by design).

- [ ] **Step 9: Commit**

```bash
git add internal/webui cmd/gendocs site CHANGELOG.md
git commit -m "feat(site): medal grid, tier strips and a reading column"
```

---

### Task 7: Browser verification (controller, not a subagent)

**Files:** none unless a defect is found (then fix in the owning task's files, with a test where one can fail, and commit `fix(site): …`).

- [ ] **Step 1: Start the dev server** with `preview_start` (add a `.claude/launch.json` entry `{"name": "gh-achievements", "runtimeExecutable": "go", "runtimeArgs": ["run", "./cmd/devserver"], "port": 8787}` if it is missing). Pages: `/`, `/check/`, `/result?u=torvalds`, `/achievements/`, `/achievements/pull-shark/`, `/achievements/yolo/`, `/methodology/`.
- [ ] **Step 2: For each page, light and dark (`resize_window colorScheme`), desktop and 375px:** screenshot; `read_console_messages` shows no errors and no CSP violations; `document.documentElement.scrollWidth <= innerWidth` (no horizontal scroll).
- [ ] **Step 3: Theme:** click the header button, reload, the choice survives; with a stored `light` and a dark system, the page paints light with no flash; the label says the theme it will switch to.
- [ ] **Step 4: Contrast:** compute `--ink`/`--muted` on `--bg` and `--surface`, `--metal-ink` on `--gold`/`--silver`/`--bronze`, in both themes, with `getComputedStyle` + a WCAG ratio in `javascript_tool`; every text pair ≥ 4.5:1.
- [ ] **Step 5: Keyboard:** Tab through the header (nav, star, theme, menu at 375px), a card's `<summary>`, the archive and the actions; the focus ring is visible on each. At 375px the Menu button opens and closes the nav and flips `aria-expanded`.
- [ ] **Step 6: Star:** the count appears in the pill and the banner; with the request blocked, the pill reads "★ Star on GitHub" and nothing errors. With analytics allowed, a star click sends `star_click` with `place` only (`read_network_requests` on the GA collect hit).
- [ ] **Step 7: Signed-in fragment:** load `/auth/callback/` with a stubbed `/api/private` response (the golden `web/testdata/result-private.html`) via `javascript_tool` calling the page's render path, or run the real dev OAuth flow if Loc is at the keyboard; the fragment is styled and its banner shows the count.
- [ ] **Step 8: Report** to Loc with screenshots (light, dark, mobile) and ask before pushing.
