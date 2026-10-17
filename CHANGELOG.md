# Changelog

## Unreleased

- New look for the site: your badges stand on a lit shelf, with a light and a dark theme (follows your system, switch in the header).
- Progress reads as count against the next threshold, e.g. `72 / 128 to x3`.
- A star button for the repository in the header and after your results.
- The terminal report now reads like the site: your badges on display, the ones still to collect, the archive in one line, and the same wording (`72 / 128 to x3`, `Top tier`, `Since 2020`).
- `gh achievements --all` lists the archive, every highlight and each badge's notes; notes that report a problem always show.
- The first report printed to a terminal ends with one request to star the repository. It never repeats: a marker file, `gh-achievements/star-shown` in your user config directory, remembers it.
- README: the achievements table shows every tier's medal and amount, and screenshots of the site in light and dark.

## v0.2.0

- Web checker at [gh-achievements.vercel.app/check/](https://gh-achievements.vercel.app/check/): public data for any username, or sign in with GitHub to include private repositories. The sign-in token is used for one check and revoked at once; nothing is stored.
- The site shows GitHub's own badge artwork.
- Site moved to Vercel. The old GitHub Pages URLs redirect.

## v0.1.0

First release.

- `gh achievements [username]`: all 15 known GitHub achievements and 6 profile highlights, your tier, unlock date and progress to the next tier.
- Official tiers, unlock dates and milestones read from GitHub's own badge history.
- Pull Shark counts calibrated against the merged pull requests GitHub names; Starstruck uses the repositories GitHub names.
- No token setup: reuses `gh auth login`, falls back to public data.
- `--json` output, `--no-color`, `NO_COLOR`.
- Prebuilt binaries for macOS, Linux and Windows (amd64, arm64).
