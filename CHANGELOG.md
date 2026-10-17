# Changelog

## v0.1.0

First release.

- `gh achievements [username]`: all 15 known GitHub achievements and 6 profile highlights, your tier, unlock date and progress to the next tier.
- Official tiers, unlock dates and milestones read from GitHub's own badge history.
- Pull Shark counts calibrated against the merged pull requests GitHub names; Starstruck uses the repositories GitHub names.
- No token setup: reuses `gh auth login`, falls back to public data.
- `--json` output, `--no-color`, `NO_COLOR`.
- Prebuilt binaries for macOS, Linux and Windows (amd64, arm64).
