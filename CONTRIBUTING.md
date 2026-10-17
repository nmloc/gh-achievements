# Contributing

Thanks for helping. The most useful contributions are reports: a badge the tool does
not know, or a number that disagrees with what GitHub shows. Use the issue forms.

## Layout

| Path | What it is |
|---|---|
| `internal/catalog` | Every known achievement. **The single source of truth** for the CLI, README table, site and `llms.txt`. |
| `internal/official` | Parsers for GitHub's profile pages and badge `/detail` history (HTML). |
| `internal/counters` | Live counts through the GitHub API, using the gh login. |
| `internal/progress` | Tier and calibration rules. |
| `internal/report`, `internal/render` | Joins both sources; text and JSON output. |
| `cmd/gendocs` | Generates the README table, `site/`, `achievements.json`, `llms.txt`, sitemap and demo image. |

## Rules the code keeps

- Never print a number that contradicts GitHub's official tier. If a count cannot be
  checked and GitHub's rule is not public, show the tier only.
- Every number carries its source: `official`, `calibrated` or `counted`.
- Every parser change comes with a fixture from a real profile.

## Develop

```bash
go test ./...                        # unit tests on saved github.com pages
go run . torvalds                    # run the CLI from source
gh extension install .               # install your checkout as `gh achievements`
go run ./cmd/gendocs                 # regenerate README table, site/ and docs/
go run ./cmd/gendocs -check          # what CI runs
./scripts/social-png.sh              # re-render the social preview PNG (needs rsvg-convert)
go test ./internal/official -run TestUpdateFixtures -update     # refresh fixtures
go test -tags canary ./internal/official -run Canary -v        # check live pages
```

## Adding an achievement

1. Add it to `internal/catalog/catalog.go` with tiers, how to earn it and its kind.
2. Add a profile that holds it to `fixtureUsers` in `internal/official/fixtures_test.go`
   and refresh the fixtures.
3. Add a parser test, then run `go run ./cmd/gendocs`.

## Releasing

Update `CHANGELOG.md`, then push a tag: `git tag v0.2.0 && git push origin v0.2.0`.
The release workflow builds every platform.

## Web (Vercel)

`api/*/index.go` are Vercel functions; `internal/web` holds the handlers; `cmd/gendocs`
writes `vercel.json` and the static pages. Run it all locally:

1. Create `.env.local` (gitignored) with `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET` (a dev
   OAuth App whose callback is `http://localhost:8787/auth/callback`), `GH_BOT_TOKEN`
   (fine-grained, public repositories, no permissions) and `STATE_SECRET`
   (`openssl rand -base64 32`).
2. `go run ./cmd/gendocs && go run ./cmd/devserver`, then open http://localhost:8787/check/.
