# Security policy

Please report a vulnerability privately through
[GitHub's private vulnerability reporting](https://github.com/nmloc/gh-achievements/security/advisories/new),
not in a public issue. You will get an answer within a week.

**CLI.** gh-achievements reads the token the GitHub CLI already stores and sends it only
to api.github.com, through the official `go-gh` library. It never writes it anywhere,
and it sends nothing to any other server.

**Web checker.** Signing in with GitHub exchanges the one-time code on the gh-achievements
server on Vercel. The resulting token is used for one request and revoked immediately;
refresh tokens are discarded and nothing is stored. Google Analytics loads only after you
consent and never receives usernames or repository names. Private vulnerability reports
cover the web checker too.
