# Web checker: design

Date: 2026-10-17 · Status: approved in brainstorming, pending spec review

## Goal

From the README, one click opens a page where anyone sees every GitHub achievement, their
tier and how far they are from the next one, with GitHub's own badge images. Signing in
with GitHub adds work in private repositories. Same accuracy rules as the CLI: every number
carries its source, and nothing contradicts GitHub's official tier.

## Constraints (measured, not assumed)

- A browser cannot read GitHub's badge history: `github.com/users/<u>/achievements/<slug>/detail`
  sends no CORS headers.
- `api.github.com` allows CORS, but anonymously only 60 requests per hour per IP.
- OAuth code exchange and device flow (`github.com/login/oauth/access_token`,
  `/login/device/code`) send no CORS headers, so login needs a server.
- No site can learn who a visitor is without a sign-in. GitHub's referrer to us is the repo
  URL, never the visitor.
- OAuth Apps have no read-only private scope: counting private work needs `repo`.
  An organisation that restricts third-party OAuth apps hides its private repos until an
  admin approves the app.

## Decisions

| Topic | Decision |
|---|---|
| Hosting | Vercel Hobby, one origin: `https://gh-achievements.vercel.app`. Go serverless functions import `internal/*` |
| Analytics | GA4, loaded only after consent |
| Private work | Included when signed in (OAuth `repo` scope) |
| Entry flow | Always show the choice card, last username pre-filled |
| Results layout | Summary + "closest next unlock" + badge grid; tap a card for details |
| Icons | GitHub's own badge artwork, hotlinked from `github.githubassets.com` |
| Token handling | Used inside one request, revoked immediately, never stored or sent to the browser |

## Routes

| Route | Kind | Behaviour |
|---|---|---|
| `/`, `/achievements/…`, `/methodology/` | static (gendocs) | As today, plus badge images and a "Check yours" entry |
| `/check` | static + small JS | The choice card: **Sign in with GitHub** (includes private) or a username box (public). The box is pre-filled from `localStorage` (username only) |
| `/check?u=<login>` | Go function | Public report built with the site's bot token. `Cache-Control: public, s-maxage=3600, stale-while-revalidate=86400`. Shareable |
| `/auth/login` | Go function | Signed `state` cookie, PKCE if GitHub supports it for OAuth Apps, redirect to GitHub with `scope=repo` |
| `/auth/callback` | Go function | Verify state → exchange code → `GET /user` → build the full report with the user's token → render → revoke the token. `Cache-Control: no-store`, `X-Robots-Tag: noindex`, `Referrer-Policy: no-referrer` |

GitHub Pages: every old URL becomes a redirect page (meta refresh + canonical) to the same
path on Vercel. Canonical URLs, sitemap, `llms.txt`, `achievements.json` URLs and the README
move to the Vercel domain.

## Results page

- **Header:** avatar, @login, "N of 8 earnable · M tiers", data scope (`public` or
  `public + private`).
- **Closest next unlock:** the counted badge with the smallest remaining count.
  **Quick wins:** unearned single-tier earnable badges (Quickdraw, YOLO, Public Sponsor),
  each linked to its how-to page.
- **Grid:** earnable badges first. "Other badges" (retired, disabled, internal) collapsed.
  Card = badge image for the tier (x1 default, x2 bronze, x3 silver, x4 gold; greyscale when
  not earned), tier pill, `<progress>` bar, one line ("869 to x4").
- **Detail panel** (tap a card): how to earn, tier table, source label with its meaning,
  notes, milestones (links; private ones shown as "private"), link to the badge page.
- **Highlights** row.
- **Actions:** public view: Share (copy `/check?u=` link), Check another, Include private
  work. Private view: "includes private work, not shareable", Refresh (signs in again).
- Server-rendered `html/template`, one renderer for both public and private results. JS
  only for the panel, `localStorage`, copy link and the consent banner. Reuses
  `site/style.css` (dark mode, 2-column grid on mobile).
- Accessibility: image alt "Pull Shark, silver tier (x3)", labelled `<progress>`,
  keyboard-focusable cards.

### Badge images

- The catalog gets an image URL per tier, harvested from the fixtures (13 of 15 badges have
  artwork; the two disabled badges never shipped and get a neutral placeholder).
- `<img onerror>` falls back to a text tile.
- The weekly canary checks every URL still answers 200.
- Footer: "Badge artwork © GitHub, loaded from GitHub's CDN."

## Login and token handling

1. `/auth/login`: random state, HMAC-SHA256 with `STATE_SECRET`, in an
   `HttpOnly; Secure; SameSite=Lax` cookie for 10 minutes and in the authorize URL.
2. `/auth/callback`: state must match the cookie, verify and be younger than 10 minutes;
   the cookie is cleared. Code exchange with the client secret on the server. The OAuth App
   expires user tokens, so the exchange also returns a refresh token: it is discarded.
3. The report is built with the user token (`counters` gets a constructor that takes a
   token), then the token is revoked with `DELETE /applications/{client_id}/token`. The
   app's grant stays, so the next sign-in is one click without a consent screen.
4. The token never reaches a log, cookie, store or response. Tests assert it.
5. CSP: `default-src 'self'`; images from `github.githubassets.com` and
   `avatars.githubusercontent.com`; scripts from self and GA; `frame-ancestors 'none'`.
   `history.replaceState` removes `?code=` from the address bar.

**Bot token** (`GH_BOT_TOKEN`): fine-grained, public repositories, no permissions. Usernames
are validated with the CLI's rule before any request.

**Accepted risk (v1):** a burst of random usernames can use up the bot's search quota (30
per minute). Affected rows say "could not count"; the page still renders. Per-IP limiting
waits until it happens.

## Analytics (GA4)

- Loaded only after the consent banner's **Accept**; the choice is kept in `localStorage`.
- Events: `check_start {method: public|signin}`, `signin_click`, `signin_complete`,
  `report_view {scope, badges_earned}`, `badge_open {badge}`, `share_click`, `cli_copy`,
  `check_error {reason}`.
- Never send a username or repository name.
- Custom dimensions in GA4: `method`, `scope`, `badge`, `reason`.

## Errors (one failure, one message)

| Case | Shown |
|---|---|
| Invalid or unknown username | Message and the username box |
| Consent denied on GitHub (`error=access_denied`) | Back to `/check` with an explanation |
| State missing, tampered or expired | "Sign-in expired, try again" |
| Rate limited | Affected rows note it; a whole-report failure offers retry |
| Account too large | 50-second internal deadline, partial report, "use the CLI for the full count" |
| Org restricts the OAuth app | Note on the report: some private repos may be hidden |
| Badge image fails | Text tile |

## Testing

- **Unit:** state sign/verify (tampered, expired, missing); callback against a fake GitHub
  (exchange, `/user`, revoke called, token absent from body and logs); response headers
  per route; closest-next-unlock and quick wins; every catalog slug has images or a
  placeholder; username validation on every route.
- **Golden:** results page HTML for a public and a private report, from fixtures.
- **Canary:** badge image URLs answer 200.
- **Local:** `cmd/devserver` on `http://localhost:8787` serves `site/` and the handlers,
  reading `.env.local` (dev OAuth App). The sign-in click is done by Loc in the browser.

## Verify first (before building on them)

1. Vercel Go runtime version and whether `maxDuration: 60` is allowed on Hobby.
2. Whether Vercel Go functions can stream (needed for real progress steps; otherwise the
   wait screen says "Checking… about 10 seconds" with no fake progress).
3. Whether GitHub OAuth Apps accept PKCE (`code_challenge`).
4. That a fine-grained token with public-repo access works with GraphQL search.

## Rollout

1. Build behind the existing static deploy; preview deployments for review.
2. Switch canonical URLs, sitemap and `llms.txt` to the Vercel domain; GitHub Pages pages
   become redirects.
3. README: a large "Check your GitHub achievements" button linking to `/check`.
4. Search Console for the Vercel domain (HTML tag verification), submit the sitemap.

## Loc's setup (outside the repo)

Vercel project (done), two OAuth Apps (prod, dev: done), bot token, `STATE_SECRET`, Vercel
env vars (Production only): `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GH_BOT_TOKEN`,
`STATE_SECRET`, `GA_MEASUREMENT_ID`; GA4 property; `.env.local` for the dev app; Search
Console after launch.

## Out of scope

Per-IP rate limiting, a stored history of results, a README SVG card, Pair Extraordinaire
and Finishing Touch counters, a custom domain.
