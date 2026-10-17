# Security policy

Please report a vulnerability privately through
[GitHub's private vulnerability reporting](https://github.com/nmloc/gh-achievements/security/advisories/new),
not in a public issue. You will get an answer within a week.

gh-achievements reads the token the GitHub CLI already stores and sends it only to
api.github.com, through the official `go-gh` library. It never writes it anywhere,
and it sends nothing to any other server.
