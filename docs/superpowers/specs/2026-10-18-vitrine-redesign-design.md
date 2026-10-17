# Vitrine redesign: design

Date: 2026-10-18 · Status: direction and results layout approved by Loc in brainstorming

## Goal

Make gh-achievements.vercel.app simple, easy to understand and good-looking for people, not
only for search engines and AI crawlers, with a dark mode the visitor can switch. Every SEO and
GEO property the site has today stays: titles, meta descriptions, canonical URLs, JSON-LD,
sitemap, llms.txt, achievements.json, the redirect tree.

## Direction: Vitrine

Your badges are medals in a lit trophy cabinet. Earned ones stand on a shelf under a spotlight,
the ones still to collect wait in the shade. The only accent colours are the medals' own metals.
One memorable thing: the spotlit shelf of real GitHub medals. Everything around it stays quiet.

### Tokens

| Token | Light | Dark | Use |
|---|---|---|---|
| `--bg` | `#eef2f4` | `#0d1c26` | page |
| `--surface` | `#ffffff` | `#122633` | header, cards, inputs |
| `--ink` | `#142430` | `#eaf1f5` | text |
| `--muted` | `#5b6e7a` | `#93a9b7` | secondary text |
| `--line` | `#cfd9df` | `#25414f` | borders |
| `--shelf` / `--edge` | `#dfe6ea` / `#c3ced5` | `#16303f` / `#08131a` | shelf surface and front edge |
| `--spot` | `rgba(255,255,255,.95)` | `rgba(64,110,140,.55)` | spotlight at the top of a shelf |
| `--gold` / `--silver` / `--bronze` | `#c9971f` / `#8a96a0` / `#b5754a` | `#e2b53a` / `#c7d0d6` / `#d39468` | tier pills, progress, CTA |

Text and UI colours meet WCAG AA in both themes (checked in the browser during build).

### Type

Bricolage Grotesque (SIL OFL), self-hosted as one variable woff2 under `/fonts/`, with
`font-display: swap` and a system-ui fallback. Weights 400, 500, 700, 800. Body 17px / 1.6,
headings on a modular scale (1.25). Sentence case everywhere; no all-caps labels.

## Theme

- Default follows `prefers-color-scheme`. A header button switches Light ↔ Dark; the choice is
  stored in `localStorage` (`gha:theme`) and wins over the system setting until changed.
- To avoid a flash of the wrong theme, `/theme.js` (tiny, external, so the CSP still forbids
  inline scripts) runs in `<head>` before the stylesheet and sets `data-theme` on `<html>`.
- CSS: `:root` holds light tokens; `:root[data-theme="dark"]` and
  `@media (prefers-color-scheme: dark) { :root:not([data-theme="light"]) }` hold dark ones.
- The button says what it will do ("Dark" / "Light"), has an `aria-label`, and is keyboard reachable.

## Pages

**Header (all pages).** Brand, then nav: Check yours, All achievements, How it works, GitHub,
theme button. Collapses to brand + theme button + a "Menu" disclosure under 640px.

**Home.** Hero: headline "Your GitHub trophy case", one line of explanation, the username field
with Check, and "Sign in with GitHub to include private work". Under it, the spotlit shelf of
the eight earnable badges at their top tier, each linking to its page. Then three short
"what it shows" points, the CLI install with a Copy button, the FAQ (JSON-LD unchanged).

**/check.** The hero's choice card on its own page (username or sign-in), same copy as today.

**/result** (approved layout).
- Header: avatar, @login, "3 of 8 badges you can still earn, 12 tiers", the closest next unlock
  on the right; the data-scope line under it ("Showing public data. Sign in with GitHub to include
  private repositories." or "Includes private work, so this page is not shareable.").
- Shelf "On display": every earned badge (any kind), tier art, tier pill in the tier's metal,
  progress bar, and the line in the new format.
- Shelf "Still to collect": earnable badges not earned, greyed art, the new line or the how-to
  for single-tier badges.
- "The archive": a closed disclosure with badges that cannot be earned and are not earned.
- Highlights row; actions: Copy share link (gold CTA) and Check another profile; private results
  show Refresh instead of the share link.
- Tapping a badge opens its details in place (native `<details>`), as today.

**Progress line format (changed).** Show where the count is against the next threshold:
`72 / 128 to x3` (count / next threshold, then the tier it unlocks); with a unit when the badge
has one: `0 / 2 answers to x1`. Top tier: `Top tier` (plus the count when there is one:
`Top tier, 252,192 stars`). Single-tier earned: `Since 2020`. Tiered badge without a counter:
`x2, next tier at 24`. Hidden profiles keep their counted numbers with the same format.

**All achievements.** A medal grid grouped into Earnable, Switched off by GitHub, Retired, and
GitHub staff only; each medal shows its name, how to earn it and its tiers, and links to its page.
The heading structure and text stay crawlable.

**Badge page.** A large medal hero with the four tier medals in a row (default, bronze, silver,
gold) above the existing facts, tier table, FAQ and "Check your progress".

**How it works** (methodology). A single readable column (max 68ch).

**Sign-in callback.** The "Checking…" card in the same style; the result replaces it.

## Motion

One moment only: on the home page the shelf's spotlight fades in once on load. Disabled under
`prefers-reduced-motion`. Opening a badge's details animates its height. Nothing else moves.

## Constraints kept

CSP unchanged except `font-src 'self'` being covered by `default-src 'self'`; no inline scripts
or styles; no build step; no new Go dependencies; GA rules unchanged; every number keeps its
source label; html/template escaping everywhere.

## Testing

- Go tests and goldens updated for the new markup and the line format (view tests assert
  `72 / 128 to x3`, `0 / 2 answers to x1`, `Top tier`, `Since 2020`).
- A test that every page loads `/theme.js` before the stylesheet and has no inline script.
- Browser check (controller): light and dark on desktop and 375px for home, result, list and a
  badge page; contrast of text tokens; keyboard focus visible; the theme survives a reload;
  no horizontal scroll; no console or CSP errors.

## Out of scope

New features, copy beyond what the layout needs, changes to the CLI output.
