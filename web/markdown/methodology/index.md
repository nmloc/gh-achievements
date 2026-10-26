# How accurate is gh-achievements?

## Three labels

gh-achievements labels every number. official: read from GitHub's own profile pages. calibrated: counted through the GitHub API, then checked against a milestone GitHub itself names. counted: counted through the API with nothing public to check it against.

## GitHub's own record of each badge

When you open a badge on a profile, GitHub loads /users/<user>/achievements/<slug>/detail. That page lists the tier, the first unlock date and history rows naming the pull request, repository or answer behind each milestone. Rows in private repositories say "inaccessible". None of this is in GitHub's REST or GraphQL API.

## Calibration

For Pull Shark, GitHub names which merged pull request was its 2nd, 16th, 128th and 1024th. gh-achievements lists your merged pull requests in merge order and finds the highest milestone it can see. For @torvalds, GitHub's 16th is subsurface/subsurface#1506, and it is also 16th in the API's list, so the two counts agree (+0). Any difference is added to the live count, so pull requests the API cannot show you are still counted.

## An official tier is never contradicted

A tier GitHub granted is never taken back, so a count below that tier's threshold is raised to it. A count above GitHub's tier is shown with "count says xN", because GitHub can lag behind the count.

## Why Pair Extraordinaire has no counter

GitHub's rule for which coauthored commits count is not public, and the obvious rule is wrong. On the author's own account, 60 merged pull requests had commits with a Co-authored-by trailer while GitHub's tier said 10 to 23, and GitHub's first unlock (2024-12-17) came before the first such pull request (2026-05-28). Commit search's co-author: qualifier returned commits co-authored by other people. So the tool shows GitHub's tier and no number.

## Known limits

GitHub search lists at most 1000 pull requests, so above that Pull Shark keeps the exact total but skips calibration. In the CLI, counts only include what your gh login can see; on the web, what the bot token (public) or your sign-in (private) can see. Galaxy Brain is counted, not calibrated. Starstruck takes the most-starred repository among the ones GitHub names and the ones you own; the star count itself is live API data. Galaxy Brain's count includes answers in GitHub Community discussions, which stopped counting in 2024, so it can run high. The official pages are HTML, so a weekly canary job checks them against real profiles and opens an issue when GitHub changes them.

## No stored token

The CLI uses the credentials gh already stores (or GH_TOKEN). Without them it falls back to public data. It talks only to github.com and api.github.com. The web checker runs on a gh-achievements server on Vercel: public checks use a read-only bot token, and signing in with GitHub gives a token that is used for one check and revoked at once. Nothing is stored.

Last verified 2026-10-17.

---

[gh-achievements](https://gh-achievements.vercel.app/) · [llms.txt](https://gh-achievements.vercel.app/llms.txt) · [About](https://gh-achievements.vercel.app/about/) · [Contact](https://gh-achievements.vercel.app/contact/) · [Privacy](https://gh-achievements.vercel.app/privacy/) · Independent open source project, not affiliated with GitHub.
