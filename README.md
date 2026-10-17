# gh-achievements: check your GitHub achievements progress

<!-- badges:start -->
<p align="center">
  <a href="https://gh-achievements.vercel.app/achievements/pull-shark/"><img src="https://github.githubassets.com/assets/pull-shark-gold-90985540b385.png" width="72" height="72" alt="Pull Shark" title="Pull Shark"></a>
  <a href="https://gh-achievements.vercel.app/achievements/starstruck/"><img src="https://github.githubassets.com/assets/starstruck-gold-a35e077693cc.png" width="72" height="72" alt="Starstruck" title="Starstruck"></a>
  <a href="https://gh-achievements.vercel.app/achievements/pair-extraordinaire/"><img src="https://github.githubassets.com/assets/pair-extraordinaire-gold-5335a3b8ae7d.png" width="72" height="72" alt="Pair Extraordinaire" title="Pair Extraordinaire"></a>
  <a href="https://gh-achievements.vercel.app/achievements/galaxy-brain/"><img src="https://github.githubassets.com/assets/galaxy-brain-gold-9c2c03fd1acf.png" width="72" height="72" alt="Galaxy Brain" title="Galaxy Brain"></a>
  <a href="https://gh-achievements.vercel.app/achievements/quickdraw/"><img src="https://github.githubassets.com/assets/quickdraw-default-39c6aec8ff89.png" width="72" height="72" alt="Quickdraw" title="Quickdraw"></a>
  <a href="https://gh-achievements.vercel.app/achievements/yolo/"><img src="https://github.githubassets.com/assets/yolo-default-be0bbff04951.png" width="72" height="72" alt="YOLO" title="YOLO"></a>
  <a href="https://gh-achievements.vercel.app/achievements/public-sponsor/"><img src="https://github.githubassets.com/assets/public-sponsor-default-9fa68986b057.png" width="72" height="72" alt="Public Sponsor" title="Public Sponsor"></a>
  <a href="https://gh-achievements.vercel.app/achievements/finishing-touch/"><img src="https://github.githubassets.com/assets/finishing-touch-gold-e19e1fcbfee2.png" width="72" height="72" alt="Finishing Touch" title="Finishing Touch"></a>
</p>
<!-- badges:end -->

<p align="center"><strong>Your GitHub trophy case: every achievement, your tier and how far the next one is, verified against GitHub's own data.</strong></p>

<p align="center">
  <a href="https://gh-achievements.vercel.app/check/"><img src="docs/check-button.svg" width="520" alt="🏆 Check your GitHub achievements → no install; sign in with GitHub to include private work"></a>
</p>

<p align="center">
  <a href="https://gh-achievements.vercel.app/check/"><strong>gh-achievements.vercel.app/check</strong></a> · no install · or use the CLI below
</p>

<p align="center">
  <a href="https://gh-achievements.vercel.app/result?u=torvalds">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="docs/screenshot-dark.png">
      <img src="docs/screenshot-light.png" width="880" alt="The gh-achievements result page for @torvalds: Pull Shark x2 at 72 / 128 to x3, Starstruck and Pair Extraordinaire at their top tier, and the retired Arctic Code Vault and Mars 2020 badges on a lit shelf; Galaxy Brain, Quickdraw, YOLO, Public Sponsor and Finishing Touch still to collect">
    </picture>
  </a>
</p>

[![Release](https://img.shields.io/github/v/release/nmloc/gh-achievements)](https://github.com/nmloc/gh-achievements/releases)
[![Tests](https://github.com/nmloc/gh-achievements/actions/workflows/test.yml/badge.svg)](https://github.com/nmloc/gh-achievements/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Two ways to check

| On the web | In your terminal |
|---|---|
| **[gh-achievements.vercel.app/check](https://gh-achievements.vercel.app/check/)** | `gh extension install nmloc/gh-achievements` then `gh achievements` |
| Type any username, or sign in with GitHub to include private repositories. The token is used for one check and revoked at once; nothing is stored. | Reuses your `gh auth login`, so it counts everything that login can see, including organizations that restrict third-party apps. |

![Terminal output of gh achievements torvalds: on display Pull Shark x2 at 72 / 128 to x3 (calibrated), Starstruck at the top tier with 252,192 stars, Pair Extraordinaire at the top tier, the retired Arctic Code Vault and Mars 2020 badges; still to collect Galaxy Brain at 0 / 2 answers, Quickdraw, YOLO, Public Sponsor and Finishing Touch](docs/demo.svg)

## What it shows

- **Every badge, not just yours.** All 15 known achievements (8 earnable, 2 switched off by GitHub, 2 retired, 3 for GitHub staff) and 6 profile highlights such as Pro.
- **How far to the next tier.** Your count against the next threshold, such as `72 / 128 to x3` for Pull Shark.
- **Checked against GitHub.** Tiers, unlock dates and milestones come from GitHub's own badge history, and every number is labelled `official`, `calibrated` or `counted`.

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
gh achievements --all        # also the archive, every highlight and each badge's notes
gh achievements --no-color   # plain text (NO_COLOR is honoured too)
```

Not signed in to gh? It still works with public data and tells you what a login would add.

## All GitHub achievements

<!-- catalog:start -->
### Earnable

You can still earn these.

| Badge | Achievement | How to earn it | Tiers |
|:-:|---|---|:-:|
| <img src="https://github.githubassets.com/assets/pull-shark-gold-90985540b385.png" width="40" height="40" alt="Pull Shark badge"> | [Pull Shark](https://gh-achievements.vercel.app/achievements/pull-shark/) | Open pull requests that get merged | <table><tr><th>Default</th><th>Bronze</th><th>Silver</th><th>Gold</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/pull-shark-default-498c279a747d.png" width="48" height="48" alt="Pull Shark base"></td><td align="center"><img src="https://github.githubassets.com/assets/pull-shark-bronze-a37accb528d1.png" width="48" height="48" alt="Pull Shark bronze"></td><td align="center"><img src="https://github.githubassets.com/assets/pull-shark-silver-0643f87ac9fd.png" width="48" height="48" alt="Pull Shark silver"></td><td align="center"><img src="https://github.githubassets.com/assets/pull-shark-gold-90985540b385.png" width="48" height="48" alt="Pull Shark gold"></td></tr><tr><td align="center">2</td><td align="center">16</td><td align="center">128</td><td align="center">1024</td></tr></table> |
| <img src="https://github.githubassets.com/assets/starstruck-gold-a35e077693cc.png" width="40" height="40" alt="Starstruck badge"> | [Starstruck](https://gh-achievements.vercel.app/achievements/starstruck/) | Create a repository that gets stars | <table><tr><th>Default</th><th>Bronze</th><th>Silver</th><th>Gold</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/starstruck-default-b6610abad518.png" width="48" height="48" alt="Starstruck base"></td><td align="center"><img src="https://github.githubassets.com/assets/starstruck-bronze-b71ec4d7766a.png" width="48" height="48" alt="Starstruck bronze"></td><td align="center"><img src="https://github.githubassets.com/assets/starstruck-silver-275c101223c0.png" width="48" height="48" alt="Starstruck silver"></td><td align="center"><img src="https://github.githubassets.com/assets/starstruck-gold-a35e077693cc.png" width="48" height="48" alt="Starstruck gold"></td></tr><tr><td align="center">16</td><td align="center">128</td><td align="center">512</td><td align="center">4096</td></tr></table> |
| <img src="https://github.githubassets.com/assets/pair-extraordinaire-gold-5335a3b8ae7d.png" width="40" height="40" alt="Pair Extraordinaire badge"> | [Pair Extraordinaire](https://gh-achievements.vercel.app/achievements/pair-extraordinaire/) | Coauthor commits on merged pull requests | <table><tr><th>Default</th><th>Bronze</th><th>Silver</th><th>Gold</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/pair-extraordinaire-default-579438a20e01.png" width="48" height="48" alt="Pair Extraordinaire base"></td><td align="center"><img src="https://github.githubassets.com/assets/pair-extraordinaire-bronze-797214f41e7e.png" width="48" height="48" alt="Pair Extraordinaire bronze"></td><td align="center"><img src="https://github.githubassets.com/assets/pair-extraordinaire-silver-a0b2b472348c.png" width="48" height="48" alt="Pair Extraordinaire silver"></td><td align="center"><img src="https://github.githubassets.com/assets/pair-extraordinaire-gold-5335a3b8ae7d.png" width="48" height="48" alt="Pair Extraordinaire gold"></td></tr><tr><td align="center">1</td><td align="center">10</td><td align="center">24</td><td align="center">48</td></tr></table> |
| <img src="https://github.githubassets.com/assets/galaxy-brain-gold-9c2c03fd1acf.png" width="40" height="40" alt="Galaxy Brain badge"> | [Galaxy Brain](https://gh-achievements.vercel.app/achievements/galaxy-brain/) | Have discussion answers accepted | <table><tr><th>Default</th><th>Bronze</th><th>Silver</th><th>Gold</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/galaxy-brain-default-847262c21056.png" width="48" height="48" alt="Galaxy Brain base"></td><td align="center"><img src="https://github.githubassets.com/assets/galaxy-brain-bronze-da3c4829fd22.png" width="48" height="48" alt="Galaxy Brain bronze"></td><td align="center"><img src="https://github.githubassets.com/assets/galaxy-brain-silver-2ce8da282a85.png" width="48" height="48" alt="Galaxy Brain silver"></td><td align="center"><img src="https://github.githubassets.com/assets/galaxy-brain-gold-9c2c03fd1acf.png" width="48" height="48" alt="Galaxy Brain gold"></td></tr><tr><td align="center">2</td><td align="center">8</td><td align="center">16</td><td align="center">32</td></tr></table> |
| <img src="https://github.githubassets.com/assets/quickdraw-default-39c6aec8ff89.png" width="40" height="40" alt="Quickdraw badge"> | [Quickdraw](https://gh-achievements.vercel.app/achievements/quickdraw/) | Close an issue or pull request within 5 minutes of opening it | <table><tr><th>Default</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/quickdraw-default-39c6aec8ff89.png" width="48" height="48" alt="Quickdraw base"></td></tr><tr><td align="center">1</td></tr></table> |
| <img src="https://github.githubassets.com/assets/yolo-default-be0bbff04951.png" width="40" height="40" alt="YOLO badge"> | [YOLO](https://gh-achievements.vercel.app/achievements/yolo/) | Merge a pull request without a review | <table><tr><th>Default</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/yolo-default-be0bbff04951.png" width="48" height="48" alt="YOLO base"></td></tr><tr><td align="center">1</td></tr></table> |
| <img src="https://github.githubassets.com/assets/public-sponsor-default-9fa68986b057.png" width="40" height="40" alt="Public Sponsor badge"> | [Public Sponsor](https://gh-achievements.vercel.app/achievements/public-sponsor/) | Sponsor someone publicly through GitHub Sponsors | <table><tr><th>Default</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/public-sponsor-default-9fa68986b057.png" width="48" height="48" alt="Public Sponsor base"></td></tr><tr><td align="center">1</td></tr></table> |
| <img src="https://github.githubassets.com/assets/finishing-touch-gold-e19e1fcbfee2.png" width="40" height="40" alt="Finishing Touch badge"> | [Finishing Touch](https://gh-achievements.vercel.app/achievements/finishing-touch/) | Get pull requests that improve GitHub's UI quality merged | <table><tr><th>Default</th><th>Bronze</th><th>Silver</th><th>Gold</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/finishing-touch-default-764b3a996d62.png" width="48" height="48" alt="Finishing Touch base"></td><td align="center"><img src="https://github.githubassets.com/assets/finishing-touch-bronze-7c49185628f3.png" width="48" height="48" alt="Finishing Touch bronze"></td><td align="center"><img src="https://github.githubassets.com/assets/finishing-touch-silver-75ab9a5c22ba.png" width="48" height="48" alt="Finishing Touch silver"></td><td align="center"><img src="https://github.githubassets.com/assets/finishing-touch-gold-e19e1fcbfee2.png" width="48" height="48" alt="Finishing Touch gold"></td></tr><tr><td align="center">1</td><td align="center">4</td><td align="center">16</td><td align="center">32</td></tr></table> |

### Switched off by GitHub

GitHub tested these and switched them off; nobody can earn them now.

| Badge | Achievement | How to earn it | Tiers |
|:-:|---|---|:-:|
| <img src="https://raw.githubusercontent.com/Schweinepriester/github-profile-achievements/549ba7291ccb68efcff5daf8af8c2c0ac6817b6f/images/tiers/heart-on-your-sleeve-gold.png" width="40" height="40" alt="Heart On Your Sleeve badge"> | [Heart On Your Sleeve](https://gh-achievements.vercel.app/achievements/heart-on-your-sleeve/) | React to something on GitHub with a ❤️ emoji | <table><tr><th>Default</th><th>Bronze</th><th>Silver</th><th>Gold</th></tr><tr><td align="center"><img src="https://raw.githubusercontent.com/Schweinepriester/github-profile-achievements/549ba7291ccb68efcff5daf8af8c2c0ac6817b6f/images/heart-on-your-sleeve-default.png" width="48" height="48" alt="Heart On Your Sleeve base"></td><td align="center"><img src="https://raw.githubusercontent.com/Schweinepriester/github-profile-achievements/549ba7291ccb68efcff5daf8af8c2c0ac6817b6f/images/tiers/heart-on-your-sleeve-bronze.png" width="48" height="48" alt="Heart On Your Sleeve bronze"></td><td align="center"><img src="https://raw.githubusercontent.com/Schweinepriester/github-profile-achievements/549ba7291ccb68efcff5daf8af8c2c0ac6817b6f/images/tiers/heart-on-your-sleeve-silver.png" width="48" height="48" alt="Heart On Your Sleeve silver"></td><td align="center"><img src="https://raw.githubusercontent.com/Schweinepriester/github-profile-achievements/549ba7291ccb68efcff5daf8af8c2c0ac6817b6f/images/tiers/heart-on-your-sleeve-gold.png" width="48" height="48" alt="Heart On Your Sleeve gold"></td></tr><tr><td align="center">?</td><td align="center">16</td><td align="center">128</td><td align="center">?</td></tr></table> |
| <img src="https://raw.githubusercontent.com/Schweinepriester/github-profile-achievements/549ba7291ccb68efcff5daf8af8c2c0ac6817b6f/images/tiers/open-sourcerer-gold.png" width="40" height="40" alt="Open Sourcerer badge"> | [Open Sourcerer](https://gh-achievements.vercel.app/achievements/open-sourcerer/) | Have pull requests merged in several public repositories | <table><tr><th>Default</th><th>Bronze</th><th>Silver</th><th>Gold</th></tr><tr><td align="center"><img src="https://raw.githubusercontent.com/Schweinepriester/github-profile-achievements/549ba7291ccb68efcff5daf8af8c2c0ac6817b6f/images/open-sourcerer-default.png" width="48" height="48" alt="Open Sourcerer base"></td><td align="center"><img src="https://raw.githubusercontent.com/Schweinepriester/github-profile-achievements/549ba7291ccb68efcff5daf8af8c2c0ac6817b6f/images/tiers/open-sourcerer-bronze.png" width="48" height="48" alt="Open Sourcerer bronze"></td><td align="center"><img src="https://raw.githubusercontent.com/Schweinepriester/github-profile-achievements/549ba7291ccb68efcff5daf8af8c2c0ac6817b6f/images/tiers/open-sourcerer-silver.png" width="48" height="48" alt="Open Sourcerer silver"></td><td align="center"><img src="https://raw.githubusercontent.com/Schweinepriester/github-profile-achievements/549ba7291ccb68efcff5daf8af8c2c0ac6817b6f/images/tiers/open-sourcerer-gold.png" width="48" height="48" alt="Open Sourcerer gold"></td></tr><tr><td align="center">?</td><td align="center">8</td><td align="center">16</td><td align="center">64</td></tr></table> |

### Retired

No longer awarded. Profiles that have them keep them.

| Badge | Achievement | How to earn it | Tiers |
|:-:|---|---|:-:|
| <img src="https://github.githubassets.com/assets/arctic-code-vault-contributor-default-df8d74122a06.png" width="40" height="40" alt="Arctic Code Vault Contributor badge"> | [Arctic Code Vault Contributor](https://gh-achievements.vercel.app/achievements/arctic-code-vault-contributor/) | Code in a repository of the 2020 GitHub Archive Program | <table><tr><th>Default</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/arctic-code-vault-contributor-default-df8d74122a06.png" width="48" height="48" alt="Arctic Code Vault Contributor base"></td></tr><tr><td align="center">1</td></tr></table> |
| <img src="https://github.githubassets.com/assets/mars-2020-contributor-default-c9cad8495acf.png" width="40" height="40" alt="Mars 2020 Contributor badge"> | [Mars 2020 Contributor](https://gh-achievements.vercel.app/achievements/mars-2020-contributor/) | Code in a repository used by the Mars 2020 Helicopter Mission | <table><tr><th>Default</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/mars-2020-contributor-default-c9cad8495acf.png" width="48" height="48" alt="Mars 2020 Contributor base"></td></tr><tr><td align="center">1</td></tr></table> |

### GitHub staff only

Only GitHub staff could earn these.

| Badge | Achievement | How to earn it | Tiers |
|:-:|---|---|:-:|
| <img src="https://github.githubassets.com/assets/proxima-pioneer-default-51cd4e4969d5.png" width="40" height="40" alt="Proxima Pioneer badge"> | [Proxima Pioneer](https://gh-achievements.vercel.app/achievements/proxima-pioneer/) | GitHub staff: contributed to the Proxima proof of concept | <table><tr><th>Default</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/proxima-pioneer-default-51cd4e4969d5.png" width="48" height="48" alt="Proxima Pioneer base"></td></tr><tr><td align="center">1</td></tr></table> |
| <img src="https://github.githubassets.com/assets/proxima-staffshipper-default-84f658288021.png" width="40" height="40" alt="Proxima Staffshipper badge"> | [Proxima Staffshipper](https://gh-achievements.vercel.app/achievements/proxima-staffshipper/) | GitHub staff: shipped the Proxima staffship instance | <table><tr><th>Default</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/proxima-staffshipper-default-84f658288021.png" width="48" height="48" alt="Proxima Staffshipper base"></td></tr><tr><td align="center">1</td></tr></table> |
| <img src="https://github.githubassets.com/assets/proxima-staffuser-default-02bf6163ea31.png" width="40" height="40" alt="Proxima Staffuser badge"> | [Proxima Staffuser](https://gh-achievements.vercel.app/achievements/proxima-staffuser/) | GitHub staff: your team was onboarded to Proxima | <table><tr><th>Default</th></tr><tr><td align="center"><img src="https://github.githubassets.com/assets/proxima-staffuser-default-02bf6163ea31.png" width="48" height="48" alt="Proxima Staffuser base"></td></tr><tr><td align="center">1</td></tr></table> |

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
| CLI runs on your machine, no server of ours | ✅ | n/a | ❌ |

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

## Like it?

If gh-achievements helped you, a ⭐ helps other people find it: [star it on GitHub](https://github.com/nmloc/gh-achievements/stargazers).

## Credits

The catalog builds on the community's work: [Schweinepriester/github-profile-achievements](https://github.com/Schweinepriester/github-profile-achievements) and [drknzz/GitHub-Achievements](https://github.com/drknzz/GitHub-Achievements).

## License

[MIT](LICENSE). `gh-achievements` is an independent project and is not affiliated with or endorsed by GitHub.
