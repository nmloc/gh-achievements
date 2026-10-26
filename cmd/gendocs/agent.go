package main

import (
	"fmt"
	"strings"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/webui"
)

// trustPage is one of the pages that say who runs the site, how to reach them and what
// happens to visitors' data. The site renders them as HTML, the Markdown twin as text.
type trustPage struct {
	Slug, Title, Heading, Description string
	Paras                             []para
}

var trustPages = []trustPage{
	{
		Slug: "about", Title: "About gh-achievements: who makes it and why",
		Heading:     "About gh-achievements",
		Description: "gh-achievements is an independent, MIT-licensed GitHub CLI extension and web checker for GitHub profile achievements, made by nmloc.",
		Paras: []para{
			{"What it is", "gh-achievements is a free, open source tool that shows every GitHub profile achievement, the tier you hold and how far you are from the next one. " +
				"It comes as a GitHub CLI extension (" + "`" + webui.Install + "`" + ", then `gh achievements`) and as a web checker on this site. " +
				"Both build the same report from the same code."},
			{"Why it exists", "GitHub shows the badges you have but not how close you are to the next tier, and its API does not expose achievements at all. " +
				"gh-achievements reads GitHub's own badge pages for tiers and unlock dates, counts progress through the GitHub API, " +
				"and labels every number official, calibrated or counted so you can tell how sure it is. The methodology page explains each source and its limits."},
			{"Who makes it", "It is written and maintained by nmloc (https://github.com/nmloc) as an independent project. " +
				"The source code, issue tracker and release history are public at " + webui.RepoURL + ". " +
				"Releases are built by GitHub Actions from tagged commits, and the achievement catalog is checked against live profiles every week."},
			{"Not affiliated with GitHub", "gh-achievements is not made, affiliated with or endorsed by GitHub. " +
				"Badge artwork belongs to GitHub and is loaded from GitHub's own CDN. The project is released under the MIT license and costs nothing to use."},
		},
	},
	{
		Slug: "contact", Title: "Contact gh-achievements: issues, security and questions",
		Heading:     "Contact",
		Description: "How to reach the gh-achievements maintainer: GitHub issues for bugs and wrong counts, private vulnerability reports for security.",
		Paras: []para{
			{"Bugs and questions", "Open an issue at " + webui.RepoURL + "/issues/new/choose. " +
				"There are templates for a bug, for a count that looks wrong and for an achievement the catalog is missing. " +
				"A public issue lets anyone with the same problem find the answer, so it is the preferred way to get in touch."},
			{"A count looks wrong", "Use the \"count looks wrong\" template and include the username, the achievement and the number GitHub shows on the profile. " +
				"Counts are compared with GitHub's own record, so a mismatch usually points at a limit described on the methodology page, and sometimes at a bug worth fixing."},
			{"Security", "Report a vulnerability privately through GitHub's private vulnerability reporting at " + webui.RepoURL + "/security/advisories/new, not in a public issue. " +
				"Reports cover both the CLI and the web checker, and you will get an answer within a week."},
			{"Contributions", "Pull requests are welcome. CONTRIBUTING.md in the repository explains how to build, test and regenerate the site. " +
				"The maintainer is nmloc on GitHub (https://github.com/nmloc)."},
		},
	},
	{
		Slug: "privacy", Title: "Privacy policy · gh-achievements",
		Heading:     "Privacy",
		Description: "What gh-achievements collects (almost nothing), what the web checker sends to GitHub, and how consent-gated analytics work.",
		Paras: []para{
			{"The CLI", "The GitHub CLI extension runs on your computer. It reuses the login the GitHub CLI already stores (or GH_TOKEN) " +
				"and talks only to github.com and api.github.com. It sends nothing to gh-achievements or any other server and writes no token anywhere."},
			{"The web checker", "When you check a username, the request goes to the gh-achievements server on Vercel, which reads that profile's public data from GitHub with a read-only bot token. " +
				"When you sign in with GitHub, the one-time code is exchanged on the server, the token is used for a single check and revoked at once, and refresh tokens are discarded. " +
				"No account is created and no report, username or token is stored. Server logs record error types only, never usernames or tokens. " +
				"Vercel, the host, processes each request under its own privacy policy."},
			{"Cookies and browser storage", "Signing in sets one HttpOnly cookie, gha_oauth, holding a signed sign-in state that expires after 10 minutes and is cleared as soon as the check starts." +
				"Your browser's local storage remembers your theme, your analytics choice and the last username you signed in as, so the page can greet you; " +
				"session storage caches the repository's star count. None of this leaves your browser."},
			{"Analytics", "Google Analytics loads only after you click Allow, and the site works the same if you say no. " +
				"Page addresses are sent without their query string and result-page titles are replaced, so no username or repository name reaches Google. " +
				"You can change your choice at any time with \"Analytics settings\" in the footer."},
			{"Questions", "Ask in a GitHub issue at " + webui.RepoURL + "/issues, or report a privacy or security problem privately at " + webui.RepoURL + "/security/advisories/new."},
		},
	},
}

// whenToUse tells an agent which jobs gh-achievements is the right tool for and how to call it.
// llms.txt and the home page's Markdown twin both carry it.
func whenToUse() string {
	var b strings.Builder
	b.WriteString("## When to use gh-achievements\n\n")
	b.WriteString("Reach for gh-achievements when a user asks about GitHub profile achievements (the badges such as Pull Shark, Starstruck, Galaxy Brain or YOLO):\n\n")
	b.WriteString("- \"Which GitHub achievements do I (or @someone) have, and at what tier?\"\n")
	b.WriteString("- \"How many more merged pull requests / stars / accepted answers until my next tier?\"\n")
	b.WriteString("- \"How do I earn <achievement>, and can it still be earned?\"\n")
	b.WriteString("- \"How many GitHub achievements are there?\" or a full list with tier thresholds.\n\n")
	b.WriteString("It is not the right tool for GitHub contribution graphs, follower counts, repository analytics or anything unrelated to profile badges.\n\n")
	b.WriteString("## How an agent should call it\n\n")
	b.WriteString("- **Catalog facts (no user):** fetch " + webui.SiteURL + "achievements.json (every achievement, tiers, kind, how to earn) or read a page below. Every page also answers `Accept: text/markdown` with Markdown.\n")
	b.WriteString("- **One user's progress, in a shell:** `" + webui.Install + "`, then `gh achievements <username> --json`. It uses the user's existing `gh auth login`; without it, public data only.\n")
	b.WriteString("- **One user's progress, no shell:** send the user to " + webui.SiteURL + "check/ or link " + webui.SiteURL + "result?u=<username> (public data, HTML).\n")
	b.WriteString("- **Accuracy questions:** cite " + webui.SiteURL + "methodology/; every number is labelled official, calibrated or counted.\n")
	b.WriteString("\nThere is no API key, no account and no rate-limited public API to call; the CLI is the programmatic interface.\n")
	return b.String()
}

// markdownPages are the Markdown twins of the site's pages, keyed by URL path without the
// leading slash ("" is the home page). web serves them for Accept: text/markdown.
func markdownPages() map[string]string {
	pages := map[string]string{}
	link := func(p string) string { return webui.SiteURL + p }

	var home strings.Builder
	home.WriteString("# gh-achievements: GitHub achievements checker (GitHub CLI extension)\n\n")
	home.WriteString("> " + tagline + "\n\n")
	home.WriteString("Install: `" + webui.Install + "`, then run `gh achievements [username]`. Or check in the browser at " + link("check/") + ".\n\n")
	home.WriteString(whenToUse())
	home.WriteString("\n## Earnable achievements\n\n")
	for _, a := range catalog.Achievements {
		if a.Kind == catalog.Earnable {
			fmt.Fprintf(&home, "- [%s](%s): %s. Tiers: %s.\n", a.Name, link("achievements/"+a.Slug+"/"), a.HowToEarn, tierText(a))
		}
	}
	fmt.Fprintf(&home, "\nAll %d achievements (%s): %s\n\n## FAQ\n", len(catalog.Achievements), kindSummary(), link("achievements/"))
	for _, qa := range siteFAQ() {
		fmt.Fprintf(&home, "\n### %s\n\n%s\n", qa.Q, qa.A)
	}
	home.WriteString(markdownFooter())
	pages[""] = home.String()

	var list strings.Builder
	fmt.Fprintf(&list, "# All GitHub achievements\n\n%d achievements (%s). Catalog last verified %s.\n", len(catalog.Achievements), kindSummary(), catalog.LastVerified)
	for _, g := range kindGroups {
		fmt.Fprintf(&list, "\n## %s\n\n%s\n\n", g.Heading, g.Note)
		for _, a := range catalog.Achievements {
			if a.Kind == g.Kind {
				fmt.Fprintf(&list, "- [%s](%s): %s. Tiers: %s.\n", a.Name, link("achievements/"+a.Slug+"/"), a.HowToEarn, tierText(a))
			}
		}
	}
	list.WriteString("\n## Profile highlights\n\n")
	for _, h := range catalog.Highlights {
		list.WriteString("- " + h.Name + "\n")
	}
	list.WriteString(markdownFooter())
	pages["achievements/"] = list.String()

	for _, a := range catalog.Achievements {
		var b strings.Builder
		fmt.Fprintf(&b, "# %s (GitHub achievement)\n\n- Status: %s\n- How to earn: %s\n- Tiers: %s\n- Progress in gh-achievements: %s\n\n%s\n",
			a.Name, statusText(a.Kind), a.HowToEarn, tierText(a), lowerFirst(progressText(a)), a.Details)
		if a.IsTiered() {
			b.WriteString("\n| Tier | Needed |\n|---|---|\n")
			for _, r := range view(a).TierRows {
				fmt.Fprintf(&b, "| %s | %s |\n", r.Label, r.Needed)
			}
		}
		if len(a.FAQ) > 0 {
			b.WriteString("\n## FAQ\n")
			for _, qa := range a.FAQ {
				fmt.Fprintf(&b, "\n### %s\n\n%s\n", qa.Q, qa.A)
			}
		}
		if a.Kind == catalog.Earnable {
			fmt.Fprintf(&b, "\nCheck your progress: `gh achievements` or %s\n", link("check/"))
		}
		b.WriteString(markdownFooter())
		pages["achievements/"+a.Slug+"/"] = b.String()
	}

	var m strings.Builder
	m.WriteString("# How accurate is gh-achievements?\n")
	for _, p := range methodology {
		fmt.Fprintf(&m, "\n## %s\n\n%s\n", p.Heading, p.Text)
	}
	fmt.Fprintf(&m, "\nLast verified %s.\n", catalog.LastVerified)
	m.WriteString(markdownFooter())
	pages["methodology/"] = m.String()

	pages["check/"] = "# Check your GitHub achievements\n\n" +
		"Type a username at " + link("check/") + " for public data, or sign in with GitHub to include private contributions. " +
		"A public report has a shareable address: " + link("result?u=<username>") + ".\n\n" +
		"In a terminal: `" + webui.Install + "`, then `gh achievements [username]` (add `--json` for machine-readable output).\n" +
		markdownFooter()

	for _, tp := range trustPages {
		var b strings.Builder
		b.WriteString("# " + tp.Heading + "\n")
		for _, p := range tp.Paras {
			fmt.Fprintf(&b, "\n## %s\n\n%s\n", p.Heading, p.Text)
		}
		b.WriteString(markdownFooter())
		pages[tp.Slug+"/"] = b.String()
	}
	return pages
}

// markdownNotFound is the body of a 404 for Accept: text/markdown.
func markdownNotFound() string {
	return "# Page not found\n\nThere is no page at this address on gh-achievements. " +
		"Every page is listed in the sitemap at " + webui.SiteURL + "sitemap.xml, " +
		"and " + webui.SiteURL + "llms.txt describes the site for agents.\n\n" +
		"- [Home](" + webui.SiteURL + ")\n" +
		"- [All GitHub achievements](" + webui.SiteURL + "achievements/)\n" +
		"- [llms.txt](" + webui.SiteURL + "llms.txt)\n"
}

func markdownFooter() string {
	return "\n---\n\n[gh-achievements](" + webui.SiteURL + ") · [llms.txt](" + webui.SiteURL + "llms.txt) · " +
		"[About](" + webui.SiteURL + "about/) · [Contact](" + webui.SiteURL + "contact/) · [Privacy](" + webui.SiteURL + "privacy/) · " +
		"Independent open source project, not affiliated with GitHub.\n"
}
