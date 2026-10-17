// Package catalog is the single source of truth for every known GitHub achievement.
//
// GitHub publishes no official list any more. This one follows the community lists
// (Schweinepriester/github-profile-achievements, drknzz/GitHub-Achievements) and was
// checked against live profiles. A badge GitHub adds later still shows up in the CLI,
// because earned badges are read from the profile, not from this table.
package catalog

// LastVerified is the date the catalog was last checked against live profiles.
const LastVerified = "2026-10-17"

// Kind says whether an achievement can be earned today.
type Kind string

const (
	Earnable Kind = "earnable"
	Disabled Kind = "disabled" // tested by GitHub, switched off, not earnable now
	Retired  Kind = "retired"  // a one-off event that is over
	Internal Kind = "internal" // GitHub staff only
)

// Counter names how progress towards the next tier is measured.
type Counter string

const (
	CounterNone      Counter = "none"       // official tier only
	CounterMergedPRs Counter = "merged-prs" // Pull Shark
	CounterStars     Counter = "stars"      // Starstruck
	CounterAnswers   Counter = "answers"    // Galaxy Brain
)

// QA is one FAQ entry. Answers state only what is verified.
type QA struct {
	Q string `json:"q"`
	A string `json:"a"`
}

// Achievement describes one badge.
type Achievement struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Kind Kind   `json:"kind"`
	// Tiers are the counts for x1, x2, x3, x4. A single-tier badge has one entry.
	// 0 marks a threshold nobody has measured yet.
	Tiers     []int   `json:"tiers"`
	Unit      string  `json:"unit,omitempty"`
	HowToEarn string  `json:"howToEarn"`
	Details   string  `json:"details"`
	Counter   Counter `json:"counter"`
	// NoCounterReason says why a tiered badge shows its tier only.
	NoCounterReason string `json:"noCounterReason,omitempty"`
	FAQ             []QA   `json:"faq,omitempty"`
}

// IsTiered reports whether the badge has more than one tier.
func (a Achievement) IsTiered() bool { return len(a.Tiers) > 1 }

// Highlight is a profile "Highlights" entry, shown under the achievements.
type Highlight struct {
	Name string `json:"name"`
	// GraphQLField is the User field that reports it, when GitHub's API has one.
	GraphQLField string `json:"graphqlField,omitempty"`
	HowToEarn    string `json:"howToEarn"`
}

// Achievements lists every known badge, in display order.
var Achievements = []Achievement{
	{
		Slug: "pull-shark", Name: "Pull Shark", Kind: Earnable,
		Tiers: []int{2, 16, 128, 1024}, Unit: "merged pull requests", Counter: CounterMergedPRs,
		HowToEarn: "Open pull requests that get merged",
		Details: "Pull Shark counts pull requests you opened that were merged. " +
			"GitHub's own history for the badge names the 2nd, 16th, 128th and 1024th merged pull request.",
		FAQ: []QA{
			{"How many merged pull requests do you need for Pull Shark?",
				"2 for the base badge, 16 for x2 (bronze), 128 for x3 (silver) and 1024 for x4 (gold)."},
			{"Do pull requests in private repositories count towards Pull Shark?",
				"Yes. GitHub's Pull Shark history lists milestones in private repositories as \"inaccessible\", so they are counted even though visitors cannot see them."},
		},
	},
	{
		Slug: "starstruck", Name: "Starstruck", Kind: Earnable,
		Tiers: []int{16, 128, 512, 4096}, Unit: "stars on one repository", Counter: CounterStars,
		HowToEarn: "Create a repository that gets stars",
		Details: "Starstruck counts the stars of a repository you created. " +
			"GitHub's history for the badge names the repository that reached each threshold.",
		FAQ: []QA{
			{"How many stars do you need for Starstruck?",
				"One repository with 16 stars for the base badge, 128 for x2, 512 for x3 and 4096 for x4."},
			{"Do stars across several repositories add up for Starstruck?",
				"No. The thresholds apply to a single repository; GitHub's history names one repository per threshold."},
		},
	},
	{
		Slug: "pair-extraordinaire", Name: "Pair Extraordinaire", Kind: Earnable,
		Tiers: []int{1, 10, 24, 48}, Unit: "coauthored merged pull requests", Counter: CounterNone,
		NoCounterReason: "GitHub's counting rule is not public: tier only, no counter",
		HowToEarn:       "Coauthor commits on merged pull requests",
		Details: "Pair Extraordinaire is awarded for coauthored commits on merged pull requests. " +
			"GitHub's history for the badge lists the pull request and the coauthors for each threshold, " +
			"including pull requests opened by someone else that list you as a coauthor.",
		FAQ: []QA{
			{"How many coauthored pull requests do you need for Pair Extraordinaire?",
				"1 for the base badge, 10 for x2, 24 for x3 and 48 for x4."},
			{"Why does gh-achievements show no counter for Pair Extraordinaire?",
				"GitHub's counting rule is not public. On a real account, 60 merged pull requests had commits with a Co-authored-by trailer while GitHub's tier said 10 to 23, so any count from the API would be wrong. The tool shows the official tier only."},
		},
	},
	{
		Slug: "galaxy-brain", Name: "Galaxy Brain", Kind: Earnable,
		Tiers: []int{2, 8, 16, 32}, Unit: "accepted answers", Counter: CounterAnswers,
		HowToEarn: "Have discussion answers accepted",
		Details: "Galaxy Brain counts your GitHub Discussions answers that were marked as the accepted answer. " +
			"Since 2024-02-06 answers in GitHub's own Community Discussions (github.com/orgs/community) no longer count.",
		FAQ: []QA{
			{"How many accepted answers do you need for Galaxy Brain?",
				"2 for the base badge, 8 for x2, 16 for x3 and 32 for x4."},
			{"Do answers in GitHub Community Discussions count for Galaxy Brain?",
				"Not since 2024-02-06, when GitHub stopped awarding achievements there to counter spam. Answers in other repositories' discussions still count."},
		},
	},
	{
		Slug: "quickdraw", Name: "Quickdraw", Kind: Earnable, Tiers: []int{1}, Counter: CounterNone,
		HowToEarn: "Close an issue or pull request within 5 minutes of opening it",
		Details:   "Quickdraw (\"Gitty up!\") is a one-time badge for closing an issue or pull request within 5 minutes of opening it.",
	},
	{
		Slug: "yolo", Name: "YOLO", Kind: Earnable, Tiers: []int{1}, Counter: CounterNone,
		HowToEarn: "Merge a pull request without a review",
		Details:   "YOLO (\"You want it? You merge it.\") is a one-time badge for merging a pull request without a code review.",
	},
	{
		Slug: "public-sponsor", Name: "Public Sponsor", Kind: Earnable, Tiers: []int{1}, Counter: CounterNone,
		HowToEarn: "Sponsor someone publicly through GitHub Sponsors",
		Details:   "Public Sponsor is a one-time badge for sponsoring an open source contributor or organization through GitHub Sponsors with a public sponsorship. Private sponsorships do not count.",
	},
	{
		Slug: "finishing-touch", Name: "Finishing Touch", Kind: Earnable,
		Tiers: []int{1, 4, 16, 32}, Unit: "merged UI quality pull requests", Counter: CounterNone,
		NoCounterReason: "GitHub decides which pull requests count as UI quality: tier only, no counter",
		HowToEarn:       "Get pull requests that improve GitHub's UI quality merged",
		Details: "Finishing Touch counts merged pull requests that GitHub classed as improving the quality of its own user interface. " +
			"GitHub's history for the badge names the 1st, 4th, 16th and 32nd UI quality pull request merged.",
		FAQ: []QA{
			{"How many pull requests do you need for Finishing Touch?",
				"1 merged UI quality pull request for the base badge, 4 for x2, 16 for x3 and 32 for x4."},
		},
	},
	{
		Slug: "heart-on-your-sleeve", Name: "Heart On Your Sleeve", Kind: Disabled,
		Tiers: []int{0, 16, 128, 0}, Unit: "❤️ reactions", Counter: CounterNone,
		HowToEarn: "React to something on GitHub with a ❤️ emoji",
		Details:   "Heart On Your Sleeve was tested by GitHub and is switched off: nobody can earn it now. The known thresholds are 16 ❤️ reactions for x2 and 128 for x3.",
	},
	{
		Slug: "open-sourcerer", Name: "Open Sourcerer", Kind: Disabled,
		Tiers: []int{0, 8, 16, 64}, Unit: "merged open source pull requests", Counter: CounterNone,
		HowToEarn: "Have pull requests merged in several public repositories",
		Details:   "Open Sourcerer was tested by GitHub and is switched off: nobody can earn it now. The known thresholds are 8 merged open source pull requests for x2, 16 for x3 and 64 for x4.",
	},
	{
		Slug: "arctic-code-vault-contributor", Name: "Arctic Code Vault Contributor", Kind: Retired,
		Tiers: []int{1}, Counter: CounterNone,
		HowToEarn: "Code in a repository of the 2020 GitHub Archive Program",
		Details:   "Arctic Code Vault Contributor went to people who contributed code to repositories stored in the 2020 GitHub Archive Program. It can no longer be earned.",
	},
	{
		Slug: "mars-2020-contributor", Name: "Mars 2020 Contributor", Kind: Retired,
		Tiers: []int{1}, Counter: CounterNone,
		HowToEarn: "Code in a repository used by the Mars 2020 Helicopter Mission",
		Details:   "Mars 2020 Contributor went to people who contributed code to repositories used in the Mars 2020 Helicopter Mission. It can no longer be earned.",
	},
	{
		Slug: "proxima-pioneer", Name: "Proxima Pioneer", Kind: Internal, Tiers: []int{1}, Counter: CounterNone,
		HowToEarn: "GitHub staff: contributed to the Proxima proof of concept",
		Details:   "Proxima Pioneer is an internal badge for GitHub staff who contributed to the Proxima proof of concept. It cannot be earned through normal activity.",
	},
	{
		Slug: "proxima-staffshipper", Name: "Proxima Staffshipper", Kind: Internal, Tiers: []int{1}, Counter: CounterNone,
		HowToEarn: "GitHub staff: shipped the Proxima staffship instance",
		Details:   "Proxima Staffshipper is an internal badge for GitHub staff who shipped the Proxima staffship instance. It cannot be earned through normal activity.",
	},
	{
		Slug: "proxima-staffuser", Name: "Proxima Staffuser", Kind: Internal, Tiers: []int{1}, Counter: CounterNone,
		HowToEarn: "GitHub staff: your team was onboarded to Proxima",
		Details:   "Proxima Staffuser is an internal badge for GitHub staff whose team was onboarded to Proxima. It cannot be earned through normal activity.",
	},
}

// Highlights lists the profile highlights.
var Highlights = []Highlight{
	{Name: "Developer Program Member", GraphQLField: "isDeveloperProgramMember", HowToEarn: "Register for the GitHub Developer Program"},
	{Name: "Security Bug Bounty Hunter", GraphQLField: "isBountyHunter", HowToEarn: "Report a valid vulnerability through GitHub Security Bug Bounty"},
	{Name: "GitHub Campus Expert", GraphQLField: "isCampusExpert", HowToEarn: "Join the GitHub Campus Expert program"},
	{Name: "GitHub Star", GraphQLField: "isGitHubStar", HowToEarn: "Be selected for the GitHub Stars program"},
	{Name: "Pro", HowToEarn: "Subscribe to GitHub Pro"},
	{Name: "Security advisory credit", HowToEarn: "Have a security advisory you submitted accepted into the GitHub Advisory Database"},
}

// BySlug returns the catalog entry for slug.
func BySlug(slug string) (Achievement, bool) {
	for _, a := range Achievements {
		if a.Slug == slug {
			return a, true
		}
	}
	return Achievement{}, false
}

// KindTitle is the section heading for a kind.
func KindTitle(k Kind) string {
	switch k {
	case Earnable:
		return "EARNABLE"
	case Disabled:
		return "DISABLED · tested by GitHub, not earnable now"
	case Retired:
		return "RETIRED · can no longer be earned"
	case Internal:
		return "INTERNAL · GitHub staff only"
	}
	return string(k)
}

// Kinds is the display order of sections.
var Kinds = []Kind{Earnable, Disabled, Retired, Internal}
