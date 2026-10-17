# CLI and README in the Vitrine style: design

Date: 2026-10-18 · Status: approved by Loc in chat ("ok, star chỉ hiện một lần, screenshot ok, làm luôn")

## Goal

The terminal output and the README read the way the redesigned site does: the same shelves, the
same wording and the same metals, so a reader who knows one knows the other.

## CLI output (`gh achievements torvalds`)

```
  🏆 @torvalds
  3 of 8 badges you can still earn, 12 tiers
  Closest next unlock: Galaxy Brain, 0 / 2 answers to x1

  On display
  ● Pull Shark                     x2  ████████░░░░  72 / 128 to x3           calibrated
  ● Starstruck                     x4  ████████████  Top tier, 252,192 stars  counted
  ● Pair Extraordinaire            x4                Top tier                 official
  ● Arctic Code Vault Contributor  ✓                 Since 2020
  ● Mars 2020 Contributor          ✓                 Since 2021

  Still to collect
  ○ Galaxy Brain                       ░░░░░░░░░░░░  0 / 2 answers to x1      counted
  ○ Quickdraw                                        Close an issue or pull request within 5 minutes of opening it
  …

  The archive: 5 badges you can no longer earn (--all lists them)
  Highlights: none of 6 yet

  official = GitHub's own pages · calibrated = counted, checked against GitHub · counted = counted only
```

- Shelves as on the web: **On display** (every earned badge, any kind), **Still to collect**
  (earnable, not earned; titled **Live counts** on a profile that hides its achievements), **The
  archive** (cannot be earned, not earned) summarised in one line.
- Wording is shared with the web, from one package: `72 / 128 to x3`, `0 / 2 answers to x1`,
  `Top tier, 252,192 stars`, `Top tier`, `x2, next tier at 24`, `First tier at 1`, `Since 2020`,
  the how-to for an unearned single-tier badge, `Switched off by GitHub`, `No longer earnable`,
  `GitHub staff only`, `Hidden on this profile`.
- Header numbers are the web's: earned of earnable, total tiers, closest next unlock.
- A bar only where there is a count (as on the web); the source label (official / calibrated /
  counted) at the end of every earnable tiered row.
- Colour: ● and the tier label in the tier's metal (bronze, silver, gold), the bar in gold,
  unearned rows dimmed, headings bold. `--no-color` and `NO_COLOR` unchanged.
- `--all` (new): lists the archive, every highlight and the notes (e.g. "checked at GitHub's #16
  subsurface/subsurface#1506: GitHub counts +0 …"). Without it the notes are hidden.
- `--json` unchanged.
- **Star line, shown once:** the first time a report is printed to a terminal, one line follows it:
  `★ Found it useful? Star it: https://github.com/nmloc/gh-achievements`. A marker file in the
  user's config directory (`<UserConfigDir>/gh-achievements/star-shown`) remembers it; if the
  marker cannot be written the line is not printed (never nag on every run). Never with `--json`
  or when stdout is not a terminal.

## README

- Top: the generated medal row, a one-line tagline, the "Check your GitHub achievements" button
  (restyled to the site's ink and gold), the link line.
- A screenshot of the result page for @torvalds, light and dark, via `<picture>` so it follows
  GitHub's theme. PNGs in `docs/`, captured with headless Chrome by a committed script
  (`scripts/readme-screenshots.mjs`); retaken by hand when the UI changes.
- "Two ways to check": web vs terminal, side by side.
- The terminal demo (`docs/demo.svg`) is regenerated from the new output, in the site's dark palette.
- "What it shows": the home page's three points.
- Kept: Install, Usage (+ `--all`), the generated achievements table, How accurate is it, FAQ, How it
  compares, JSON output, Contributing, Credits, License. A short star request before Credits.

## Constraints

No new Go dependencies; no npm dependencies (the screenshot script uses Node's built-in WebSocket
and fetch). The CLI never writes anywhere but its own marker file. Commits as nmloc with no
Co-Authored-By trailer. Every generated file stays in sync (`go run ./cmd/gendocs -check`).
