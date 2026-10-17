# gh-achievements: check your GitHub achievements progress

**Check your GitHub achievements progress: every badge, your tier and how far to the next one, verified against GitHub's own data, no token needed.**

**[🏆 Check your GitHub achievements →](https://gh-achievements.vercel.app/check/)** (no install; sign in to include private work)

[![Release](https://img.shields.io/github/v/release/nmloc/gh-achievements)](https://github.com/nmloc/gh-achievements/releases)
[![Tests](https://github.com/nmloc/gh-achievements/actions/workflows/test.yml/badge.svg)](https://github.com/nmloc/gh-achievements/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

```bash
gh extension install nmloc/gh-achievements
gh achievements
```

![Terminal output of gh achievements torvalds: Pull Shark x2 at 72 of 128 merged pull requests, calibrated against GitHub's 16th milestone; Starstruck x4 from torvalds/linux; Pair Extraordinaire x4; retired Arctic Code Vault and Mars 2020 badges](docs/demo.svg)

`gh-achievements` is a [GitHub CLI](https://cli.github.com/) extension that lists **all GitHub profile achievements** (Pull Shark, Starstruck, Pair Extraordinaire, Galaxy Brain, Quickdraw, YOLO, Public Sponsor, Finishing Touch and the retired ones) and shows **your tier and progress to the next one**. Site: **[gh-achievements.vercel.app](https://gh-achievements.vercel.app/)**.

- **Every achievement**, not only the ones you have, plus profile highlights such as Pro and Developer Program Member.
- **Your tier and the next one**: unlock date, a progress bar, and how many more you need.
- **Checked against GitHub**: tiers and milestones come from GitHub's own badge history; Pull Shark's count is calibrated against the pull requests GitHub names.
- **No token**: it reuses `gh auth login`. Nothing goes through a third-party server.

## Install

You need the [GitHub CLI](https://cli.github.com/) (`brew install gh`, `winget install GitHub.cli`, or [another installer](https://github.com/cli/cli#installation)).

```bash
gh extension install nmloc/gh-achievements
```

Upgrade with `gh extension upgrade achievements`. Prebuilt binaries for macOS, Linux and Windows (amd64 and arm64) come from [Releases](https://github.com/nmloc/gh-achievements/releases), so no Go toolchain is needed.

## Usage

```bash
gh achievements              # the account you are signed in to with gh
gh achievements torvalds     # anyone's public profile
gh achievements --json       # machine-readable, for scripts
gh achievements --no-color   # plain text (NO_COLOR is honoured too)
```

Not signed in to gh? It still works with public data and tells you what a login would add.

## Web

[gh-achievements.vercel.app](https://gh-achievements.vercel.app/check/): type a username for public data, or sign in with GitHub to include private repositories. The sign-in token is used for one check and revoked at once; nothing is stored. Organizations that restrict third-party apps hide their private repositories until an admin approves gh-achievements (the CLI sees them because it uses gh's own login).

## All GitHub achievements

<!-- catalog:start -->
| Achievement | How to earn it | Tiers | Status | What `gh achievements` shows |
|---|---|---|---|---|
| [Pull Shark](https://gh-achievements.vercel.app/achievements/pull-shark/) | Open pull requests that get merged | 2 · 16 · 128 · 1024 | Earnable | Count, checked against GitHub's milestones |
| [Starstruck](https://gh-achievements.vercel.app/achievements/starstruck/) | Create a repository that gets stars | 16 · 128 · 512 · 4096 | Earnable | Stars of the repository GitHub names |
| [Pair Extraordinaire](https://gh-achievements.vercel.app/achievements/pair-extraordinaire/) | Coauthor commits on merged pull requests | 1 · 10 · 24 · 48 | Earnable | Official tier (no public counting rule) |
| [Galaxy Brain](https://gh-achievements.vercel.app/achievements/galaxy-brain/) | Have discussion answers accepted | 2 · 8 · 16 · 32 | Earnable | Count of accepted answers |
| [Quickdraw](https://gh-achievements.vercel.app/achievements/quickdraw/) | Close an issue or pull request within 5 minutes of opening it | one tier | Earnable | Earned or not (official) |
| [YOLO](https://gh-achievements.vercel.app/achievements/yolo/) | Merge a pull request without a review | one tier | Earnable | Earned or not (official) |
| [Public Sponsor](https://gh-achievements.vercel.app/achievements/public-sponsor/) | Sponsor someone publicly through GitHub Sponsors | one tier | Earnable | Earned or not (official) |
| [Finishing Touch](https://gh-achievements.vercel.app/achievements/finishing-touch/) | Get pull requests that improve GitHub's UI quality merged | 1 · 4 · 16 · 32 | Earnable | Official tier (no public counting rule) |
| [Heart On Your Sleeve](https://gh-achievements.vercel.app/achievements/heart-on-your-sleeve/) | React to something on GitHub with a ❤️ emoji | ? · 16 · 128 · ? | Disabled (tested, not earnable now) | Earned or not |
| [Open Sourcerer](https://gh-achievements.vercel.app/achievements/open-sourcerer/) | Have pull requests merged in several public repositories | ? · 8 · 16 · 64 | Disabled (tested, not earnable now) | Earned or not |
| [Arctic Code Vault Contributor](https://gh-achievements.vercel.app/achievements/arctic-code-vault-contributor/) | Code in a repository of the 2020 GitHub Archive Program | one tier | Retired (no longer earnable) | Earned or not |
| [Mars 2020 Contributor](https://gh-achievements.vercel.app/achievements/mars-2020-contributor/) | Code in a repository used by the Mars 2020 Helicopter Mission | one tier | Retired (no longer earnable) | Earned or not |
| [Proxima Pioneer](https://gh-achievements.vercel.app/achievements/proxima-pioneer/) | GitHub staff: contributed to the Proxima proof of concept | one tier | Internal (GitHub staff only) | Earned or not |
| [Proxima Staffshipper](https://gh-achievements.vercel.app/achievements/proxima-staffshipper/) | GitHub staff: shipped the Proxima staffship instance | one tier | Internal (GitHub staff only) | Earned or not |
| [Proxima Staffuser](https://gh-achievements.vercel.app/achievements/proxima-staffuser/) | GitHub staff: your team was onboarded to Proxima | one tier | Internal (GitHub staff only) | Earned or not |

Profile highlights it also reports: Developer Program Member, Security Bug Bounty Hunter, GitHub Campus Expert, GitHub Star, Pro, Security advisory credit.

_Catalog last verified against live profiles on 2026-10-17._
<!-- catalog:end -->

## How accurate is it?

Every number carries a label:

| Label | Meaning |
|---|---|
| `official` | Read from GitHub's own profile pages: tier, unlock date, the pull request or repository behind each milestone. |
| `calibrated` | Counted through the GitHub API, then checked against a milestone GitHub itself names. |
| `counted` | Counted through the API, with nothing public to check it against. |

GitHub keeps a history for each badge (the dialog you see when you click one on a profile). It names, for example, your 2nd, 16th and 128th merged pull request. `gh-achievements` reads that history, lists your merged pull requests in merge order, and checks that the two agree. For [@torvalds](https://github.com/torvalds?tab=achievements), GitHub's 16th merged pull request is `subsurface/subsurface#1506`, and it is 16th in the API's list too: **+0**.

A tier GitHub granted is never contradicted. When GitHub's counting rule is not public, as for Pair Extraordinaire, the tool shows the official tier and **no number**, rather than one that is wrong. Details and known limits: **[How accurate is it?](https://gh-achievements.vercel.app/methodology/)**

## FAQ

**How do I check my GitHub achievements progress?**
Install with `gh extension install nmloc/gh-achievements` and run `gh achievements`.

**Does it need a GitHub personal access token?**
No. It uses the credentials the GitHub CLI already stores (or `GH_TOKEN`). Without them it falls back to public data.

**How many GitHub achievements are there?**
15 are known: 8 earnable, 2 disabled, 2 retired and 3 internal to GitHub staff. See the table above.

**Why don't my achievements show on my profile?**
They can be turned off under *Settings → Public profile → Show Achievements on my profile*. The tool says when a profile hides them.

**Can it see private contributions?**
GitHub counts them, and its badge history lists them as "inaccessible". The tool counts what your gh login can see and calibrates the rest against GitHub's milestones.

## How it compares

| | gh-achievements | Achievement lists and guides | Web trackers |
|---|---|---|---|
| Lists every achievement | ✅ | ✅ | some |
| Shows *your* tier and progress | ✅ | ❌ | ✅ |
| Checks counts against GitHub's own history | ✅ | ❌ | ❌ |
| No personal access token to create | ✅ | n/a | usually ❌ |
| Runs on your machine, no third-party server | ✅ | n/a | ❌ |

## JSON output

`gh achievements --json` prints the full report, for scripts, dashboards or a README card:

```json
{
  "user": "torvalds",
  "achievementsHidden": false,
  "authenticated": true,
  "catalogVerified": "2026-10-17",
  "achievements": [
    {
      "slug": "pull-shark",
      "name": "Pull Shark",
      "kind": "earnable",
      "tiers": [2, 16, 128, 1024],
      "status": "earned",
      "officialTier": 2,
      "unlocked": "2017-01-17",
      "standing": { "tier": 2, "count": 72, "floor": 16, "target": 128, "remaining": 56, "isMax": false, "isAhead": false },
      "source": "calibrated"
    }
  ],
  "highlights": [{ "name": "Pro", "earned": false }]
}
```

The catalog alone is published as [achievements.json](https://gh-achievements.vercel.app/achievements.json).

## Contributing

Spotted a new badge, or a count that looks wrong? [Open an issue](https://github.com/nmloc/gh-achievements/issues/new/choose). See [CONTRIBUTING.md](CONTRIBUTING.md) for how the catalog, fixtures and site are built.

## Credits

The catalog builds on the community's work: [Schweinepriester/github-profile-achievements](https://github.com/Schweinepriester/github-profile-achievements) and [drknzz/GitHub-Achievements](https://github.com/drknzz/GitHub-Achievements).

## License

[MIT](LICENSE). `gh-achievements` is an independent project and is not affiliated with or endorsed by GitHub.
