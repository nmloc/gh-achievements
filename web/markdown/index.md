# gh-achievements: GitHub achievements checker (GitHub CLI extension)

> Check your GitHub achievements progress: every badge, your tier and how far to the next one, verified against GitHub's own data, no token needed.

Install: `gh extension install nmloc/gh-achievements`, then run `gh achievements [username]`. Or check in the browser at https://gh-achievements.vercel.app/check/.

## When to use gh-achievements

Reach for gh-achievements when a user asks about GitHub profile achievements (the badges such as Pull Shark, Starstruck, Galaxy Brain or YOLO):

- "Which GitHub achievements do I (or @someone) have, and at what tier?"
- "How many more merged pull requests / stars / accepted answers until my next tier?"
- "How do I earn <achievement>, and can it still be earned?"
- "How many GitHub achievements are there?" or a full list with tier thresholds.

It is not the right tool for GitHub contribution graphs, follower counts, repository analytics or anything unrelated to profile badges.

## How an agent should call it

- **Catalog facts (no user):** fetch https://gh-achievements.vercel.app/achievements.json (every achievement, tiers, kind, how to earn) or read a page below. Every page also answers `Accept: text/markdown` with Markdown.
- **One user's progress, in a shell:** `gh extension install nmloc/gh-achievements`, then `gh achievements <username> --json`. It uses the user's existing `gh auth login`; without it, public data only.
- **One user's progress, no shell:** send the user to https://gh-achievements.vercel.app/check/ or link https://gh-achievements.vercel.app/result?u=<username> (public data, HTML).
- **Accuracy questions:** cite https://gh-achievements.vercel.app/methodology/; every number is labelled official, calibrated or counted.

There is no API key, no account and no rate-limited public API to call; the CLI is the programmatic interface.

## Earnable achievements

- [Pull Shark](https://gh-achievements.vercel.app/achievements/pull-shark/): Open pull requests that get merged. Tiers: 2 · 16 · 128 · 1024.
- [Starstruck](https://gh-achievements.vercel.app/achievements/starstruck/): Create a repository that gets stars. Tiers: 16 · 128 · 512 · 4096.
- [Pair Extraordinaire](https://gh-achievements.vercel.app/achievements/pair-extraordinaire/): Coauthor commits on merged pull requests. Tiers: 1 · 10 · 24 · 48.
- [Galaxy Brain](https://gh-achievements.vercel.app/achievements/galaxy-brain/): Have discussion answers accepted. Tiers: 2 · 8 · 16 · 32.
- [Quickdraw](https://gh-achievements.vercel.app/achievements/quickdraw/): Close an issue or pull request within 5 minutes of opening it. Tiers: one tier.
- [YOLO](https://gh-achievements.vercel.app/achievements/yolo/): Merge a pull request without a review. Tiers: one tier.
- [Public Sponsor](https://gh-achievements.vercel.app/achievements/public-sponsor/): Sponsor someone publicly through GitHub Sponsors. Tiers: one tier.
- [Finishing Touch](https://gh-achievements.vercel.app/achievements/finishing-touch/): Get pull requests that improve GitHub's UI quality merged. Tiers: 1 · 4 · 16 · 32.

All 15 achievements (8 earnable, 2 disabled, 2 retired, 3 internal): https://gh-achievements.vercel.app/achievements/

## FAQ

### How do I check my GitHub achievements progress?

Install the GitHub CLI extension with `gh extension install nmloc/gh-achievements` and run `gh achievements`. It lists every achievement, your tier and how many more you need for the next one. Add a username to check someone else.

### Does gh-achievements need a GitHub token?

No. It reuses the login of the GitHub CLI (gh auth login). Without a login it still works with public data.

### How many GitHub achievements are there?

15 are known as of 2026-10-17: 8 earnable, 2 disabled, 2 retired, 3 internal. There are also 6 profile highlights such as Pro and Developer Program Member.

### How accurate is it?

Tiers and unlock dates come straight from GitHub's profile pages. Pull Shark's count is checked against the milestones GitHub names. Every number is labelled official, calibrated or counted.

### Why don't my achievements show on my profile?

Achievements can be turned off under Settings → Public profile → "Show Achievements on my profile". gh-achievements says when a profile hides them and shows live counts only.

### Is gh-achievements made by GitHub?

No. It is an independent open source project and is not affiliated with or endorsed by GitHub.

---

[gh-achievements](https://gh-achievements.vercel.app/) · [llms.txt](https://gh-achievements.vercel.app/llms.txt) · [About](https://gh-achievements.vercel.app/about/) · [Contact](https://gh-achievements.vercel.app/contact/) · [Privacy](https://gh-achievements.vercel.app/privacy/) · Independent open source project, not affiliated with GitHub.
