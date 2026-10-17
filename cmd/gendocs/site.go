package main

import (
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/nmloc/gh-achievements/internal/catalog"
	"github.com/nmloc/gh-achievements/internal/webui"
)

type para struct{ Heading, Text string }

// maxDescription is about what Google shows of a meta description.
const maxDescription = 155

// methodology is the original research the site and llms-full.txt publish.
var methodology = []para{
	{"Three labels", "gh-achievements labels every number. official: read from GitHub's own profile pages. " +
		"calibrated: counted through the GitHub API, then checked against a milestone GitHub itself names. " +
		"counted: counted through the API with nothing public to check it against."},
	{"GitHub's own record of each badge", "When you open a badge on a profile, GitHub loads /users/<user>/achievements/<slug>/detail. " +
		"That page lists the tier, the first unlock date and history rows naming the pull request, repository or answer behind each milestone. " +
		"Rows in private repositories say \"inaccessible\". None of this is in GitHub's REST or GraphQL API."},
	{"Calibration", "For Pull Shark, GitHub names which merged pull request was its 2nd, 16th, 128th and 1024th. " +
		"gh-achievements lists your merged pull requests in merge order and finds the highest milestone it can see. " +
		"For @torvalds, GitHub's 16th is subsurface/subsurface#1506, and it is also 16th in the API's list, so the two counts agree (+0). " +
		"Any difference is added to the live count, so pull requests the API cannot show you are still counted."},
	{"An official tier is never contradicted", "A tier GitHub granted is never taken back, so a count below that tier's threshold is raised to it. " +
		"A count above GitHub's tier is shown with \"count says xN\", because GitHub can lag behind the count."},
	{"Why Pair Extraordinaire has no counter", "GitHub's rule for which coauthored commits count is not public, and the obvious rule is wrong. " +
		"On the author's own account, 60 merged pull requests had commits with a Co-authored-by trailer while GitHub's tier said 10 to 23, " +
		"and GitHub's first unlock (2024-12-17) came before the first such pull request (2026-05-28). " +
		"Commit search's co-author: qualifier returned commits co-authored by other people. So the tool shows GitHub's tier and no number."},
	{"Known limits", "GitHub search lists at most 1000 pull requests, so above that Pull Shark keeps the exact total but skips calibration. " +
		"In the CLI, counts only include what your gh login can see; on the web, what the bot token (public) or your sign-in (private) can see. Galaxy Brain is counted, not calibrated. " +
		"Starstruck takes the most-starred repository among the ones GitHub names and the ones you own; the star count itself is live API data. " +
		"Galaxy Brain's count includes answers in GitHub Community discussions, which stopped counting in 2024, so it can run high. " +
		"The official pages are HTML, so a weekly canary job checks them against real profiles and opens an issue when GitHub changes them."},
	{"No stored token", "The CLI uses the credentials gh already stores (or GH_TOKEN). Without them it falls back to public data. " +
		"It talks only to github.com and api.github.com. " +
		"The web checker runs on a gh-achievements server on Vercel: public checks use a read-only bot token, and signing in with GitHub gives a token that is used for one check and revoked at once. Nothing is stored."},
}

func siteFAQ() []catalog.QA {
	return []catalog.QA{
		{Q: "How do I check my GitHub achievements progress?",
			A: "Install the GitHub CLI extension with `" + webui.Install + "` and run `gh achievements`. It lists every achievement, your tier and how many more you need for the next one. Add a username to check someone else."},
		{Q: "Does gh-achievements need a GitHub token?",
			A: "No. It reuses the login of the GitHub CLI (gh auth login). Without a login it still works with public data."},
		{Q: "How many GitHub achievements are there?",
			A: fmt.Sprintf("%d are known as of %s: %s. There are also %d profile highlights such as Pro and Developer Program Member.",
				len(catalog.Achievements), catalog.LastVerified, kindSummary(), len(catalog.Highlights))},
		{Q: "How accurate is it?",
			A: "Tiers and unlock dates come straight from GitHub's profile pages. Pull Shark's count is checked against the milestones GitHub names. Every number is labelled official, calibrated or counted."},
		{Q: "Why don't my achievements show on my profile?",
			A: "Achievements can be turned off under Settings → Public profile → \"Show Achievements on my profile\". gh-achievements says when a profile hides them and shows live counts only."},
		{Q: "Is gh-achievements made by GitHub?",
			A: "No. It is an independent open source project and is not affiliated with or endorsed by GitHub."},
	}
}

func ld(v interface{}) template.JS {
	b, _ := json.Marshal(v) // json.Marshal escapes <, > and & for script contexts
	return template.JS(b)
}

func faqLD(qas []catalog.QA) template.JS {
	items := make([]map[string]interface{}, len(qas))
	for i, qa := range qas {
		items[i] = map[string]interface{}{
			"@type": "Question", "name": qa.Q,
			"acceptedAnswer": map[string]interface{}{"@type": "Answer", "text": qa.A},
		}
	}
	return ld(map[string]interface{}{"@context": "https://schema.org", "@type": "FAQPage", "mainEntity": items})
}

func breadcrumbLD(crumbs []webui.Crumb) template.JS {
	items := make([]map[string]interface{}, len(crumbs))
	for i, c := range crumbs {
		items[i] = map[string]interface{}{"@type": "ListItem", "position": i + 1, "name": c.Name, "item": c.URL}
	}
	return ld(map[string]interface{}{"@context": "https://schema.org", "@type": "BreadcrumbList", "itemListElement": items})
}

type badgeView struct {
	catalog.Achievement
	TierRows   []tierRow
	Status     string
	Progress   string
	TierText   string
	Image      string
	Initials   string
	Prev, Next *catalog.Achievement
}

type tierRow struct {
	Label  string
	Needed string
}

func view(a catalog.Achievement) badgeView {
	v := badgeView{Achievement: a, Status: statusText(a.Kind), Progress: progressText(a), TierText: tierText(a),
		Image: catalog.Image(a.Slug, len(a.Tiers)), Initials: webui.Initials(a.Name)}
	if a.IsTiered() {
		for i, t := range a.Tiers {
			needed := "unknown"
			if t > 0 {
				needed = fmt.Sprintf("%d %s", t, a.Unit)
			}
			v.TierRows = append(v.TierRows, tierRow{tierNames[i], needed})
		}
	}
	return v
}

func sitePages() (map[string][]byte, error) {
	home := webui.Crumb{Name: "gh-achievements", URL: webui.SiteURL}
	all := webui.Crumb{Name: "All GitHub achievements", URL: webui.SiteURL + "achievements/"}
	var views []badgeView
	for _, a := range catalog.Achievements {
		views = append(views, view(a))
	}
	var earnable []badgeView
	for _, v := range views {
		if v.Kind == catalog.Earnable {
			earnable = append(earnable, v)
		}
	}

	pages := []webui.Page{
		{
			Path: "index.html", URL: webui.SiteURL,
			Title:       "GitHub Achievements Checker: tiers & progress · gh-achievements",
			Description: "See every GitHub achievement, your tier and how far you are from the next one. A free GitHub CLI extension, verified against GitHub, no token needed.",
			Heading:     "", // the hero holds the <h1>
			Content:     "home.html",
			JSONLD: []template.JS{
				ld(map[string]interface{}{
					"@context": "https://schema.org", "@type": "SoftwareApplication",
					"name": "gh-achievements", "description": tagline,
					"applicationCategory": "DeveloperApplication", "operatingSystem": "macOS, Linux, Windows",
					"url": webui.SiteURL, "downloadUrl": webui.RepoURL + "/releases", "codeRepository": webui.RepoURL,
					"license": "https://opensource.org/licenses/MIT", "isAccessibleForFree": true,
					"offers": map[string]interface{}{"@type": "Offer", "price": "0", "priceCurrency": "USD"},
					"author": map[string]interface{}{"@type": "Person", "name": "nmloc", "url": "https://github.com/nmloc"},
				}),
				faqLD(siteFAQ()),
			},
			Data: map[string]interface{}{"Earnable": earnable, "FAQ": siteFAQ(), "Install": webui.Install,
				"Summary": kindSummary(), "Count": len(catalog.Achievements), "HighlightCount": len(catalog.Highlights)},
		},
		{
			Path: "achievements/index.html", URL: all.URL,
			Title:       fmt.Sprintf("All %d GitHub Achievements: list, tiers & how to get them", len(catalog.Achievements)),
			Description: "The complete list of GitHub profile achievements, with tier thresholds, how to earn each one and whether it can still be earned.",
			Heading:     "All GitHub achievements",
			Content:     "list.html",
			Crumbs:      []webui.Crumb{home, all},
			Data: map[string]interface{}{"Badges": views, "Highlights": catalog.Highlights, "Verified": catalog.LastVerified,
				"Summary": kindSummary(), "Install": webui.Install},
		},
		{
			Path: "methodology/index.html", URL: webui.SiteURL + "methodology/",
			Title:       "How accurate is gh-achievements? Data sources & limits",
			Description: "Where gh-achievements gets each number: GitHub's own badge history, calibration against named milestones, and the limits of GitHub's API.",
			Heading:     "How accurate is gh-achievements?",
			Content:     "methodology.html",
			Crumbs:      []webui.Crumb{home, {Name: "How accurate is it", URL: webui.SiteURL + "methodology/"}},
			Data:        map[string]interface{}{"Paras": methodology, "Verified": catalog.LastVerified},
		},
		{
			Path: "check/index.html", URL: webui.SiteURL + "check/",
			Title:       "Check your GitHub achievements · gh-achievements",
			Description: "Sign in with GitHub or type a username to see every achievement, the tier and how far to the next one.",
			Heading:     "Check your GitHub achievements", Content: "check.html",
			Crumbs: []webui.Crumb{home, {Name: "Check yours", URL: webui.SiteURL + "check/"}},
		},
		{
			Path: "auth/callback/index.html", URL: webui.SiteURL + "auth/callback/", Title: "Checking… · gh-achievements",
			Description: "Checking your GitHub achievements.", Heading: "Your GitHub achievements",
			Content: "callback.html", NoIndex: true, NoReferrer: true, Scripts: []string{"/callback.js"},
		},
	}
	for i, v := range views {
		v := v
		if i > 0 {
			v.Prev = &catalog.Achievements[i-1]
		}
		if i < len(views)-1 {
			v.Next = &catalog.Achievements[i+1]
		}
		url := webui.SiteURL + "achievements/" + v.Slug + "/"
		crumbs := []webui.Crumb{home, all, {Name: v.Name, URL: url}}
		desc := fmt.Sprintf("%s is a GitHub achievement. %s.", v.Name, v.HowToEarn)
		if v.IsTiered() {
			desc += " Tiers: " + v.TierText + "."
		}
		tail := " " + v.Status + "."
		if v.Kind == catalog.Earnable {
			tail = " Check your progress with gh achievements."
		}
		if len(desc)+len(tail) <= maxDescription {
			desc += tail
		}
		p := webui.Page{
			Path: "achievements/" + v.Slug + "/index.html", URL: url,
			Title:       badgeTitle(v.Name),
			Description: desc,
			Heading:     v.Name,
			Content:     "badge.html",
			Crumbs:      crumbs,
			JSONLD:      []template.JS{breadcrumbLD(crumbs)},
			Data:        map[string]interface{}{"B": v, "Install": webui.Install, "Verified": catalog.LastVerified},
		}
		if len(v.FAQ) > 0 {
			p.JSONLD = append(p.JSONLD, faqLD(v.FAQ))
		}
		pages = append(pages, p)
	}
	pages = append(pages, webui.Page{
		Path: "404.html", URL: webui.SiteURL + "404.html", Title: "Page not found · gh-achievements",
		Description: "This page does not exist.", Heading: "Page not found", NoIndex: true,
		Content: "404.html",
	})

	out := map[string][]byte{}
	for _, name := range []string{"style.css", "theme.js", "app.js", "callback.js", "fonts/bricolage-grotesque.woff2", "fonts/OFL.txt"} {
		b, err := webui.Static(name)
		if err != nil {
			return nil, err
		}
		out["site/"+name] = b
	}
	for _, p := range pages {
		body, err := webui.Render(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.Path, err)
		}
		out["site/"+p.Path] = body
	}
	return out, nil
}

// badgeTitle keeps the <title> within what search results show (~60 characters).
func badgeTitle(name string) string {
	for _, suffix := range []string{" GitHub Achievement: how to get it & tiers", " GitHub Achievement: how to get it", " GitHub Achievement"} {
		if len(name+suffix) <= 60 {
			return name + suffix
		}
	}
	return name + " · GitHub achievement"
}
