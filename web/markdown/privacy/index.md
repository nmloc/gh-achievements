# Privacy

## The CLI

The GitHub CLI extension runs on your computer. It reuses the login the GitHub CLI already stores (or GH_TOKEN) and talks only to github.com and api.github.com. It sends nothing to gh-achievements or any other server and writes no token anywhere.

## The web checker

When you check a username, the request goes to the gh-achievements server on Vercel, which reads that profile's public data from GitHub with a read-only bot token. When you sign in with GitHub, the one-time code is exchanged on the server, the token is used for a single check and revoked at once, and refresh tokens are discarded. No account is created and no report, username or token is stored. Server logs record error types only, never usernames or tokens. Vercel, the host, processes each request under its own privacy policy.

## Cookies and browser storage

Signing in sets one HttpOnly cookie, gha_oauth, holding a signed sign-in state that expires after 10 minutes and is cleared as soon as the check starts.Your browser's local storage remembers your theme, your analytics choice and the last username you signed in as, so the page can greet you; session storage caches the repository's star count. None of this leaves your browser.

## Analytics

Google Analytics loads only after you click Allow, and the site works the same if you say no. Page addresses are sent without their query string and result-page titles are replaced, so no username or repository name reaches Google. You can change your choice at any time with "Analytics settings" in the footer.

## Questions

Ask in a GitHub issue at https://github.com/nmloc/gh-achievements/issues, or report a privacy or security problem privately at https://github.com/nmloc/gh-achievements/security/advisories/new.

---

[gh-achievements](https://gh-achievements.vercel.app/) · [llms.txt](https://gh-achievements.vercel.app/llms.txt) · [About](https://gh-achievements.vercel.app/about/) · [Contact](https://gh-achievements.vercel.app/contact/) · [Privacy](https://gh-achievements.vercel.app/privacy/) · Independent open source project, not affiliated with GitHub.
