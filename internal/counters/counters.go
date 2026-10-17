// Package counters counts live activity through the GitHub API.
//
// It reuses the token `gh auth login` already stored (or GH_TOKEN), so nobody creates a
// token for this tool. Without one it falls back to the public REST API: merged PRs and
// stars still count, GraphQL-only data (answers, sponsorships, highlight flags) does not.
package counters

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

// Facts are the GraphQL-only user fields.
type Facts struct {
	AcceptedAnswers int
	Sponsoring      int
	Flags           map[string]bool // highlight GraphQL field → value
}

// MergedPRs is the result of counting a user's merged pull requests.
type MergedPRs struct {
	Total int      // GitHub search's issueCount: right even past the listing cap
	Refs  []string // "owner/repo#123", oldest merge first; may be shorter than Total
}

// IsComplete reports whether every merged PR was listed, which calibration needs.
func (m MergedPRs) IsComplete() bool { return len(m.Refs) == m.Total }

// Source counts activity for one user.
type Source interface {
	IsAuthenticated() bool
	MergedPRs(user string) (MergedPRs, error)
	MaxOwnedStars(user string) (stars int, repo string, err error)
	RepoStars(repo string) (int, error)
	Facts(user string, flagFields []string) (*Facts, error) // nil without a token
}

// New returns a Source using the gh login when there is one.
func New() Source {
	rest, err := api.DefaultRESTClient()
	if err != nil {
		return newPublic()
	}
	gql, err := api.DefaultGraphQLClient()
	if err != nil {
		return newPublic()
	}
	return &authed{rest: rest, gql: gql}
}

// ---------- authenticated (gh login) ----------

type authed struct {
	rest *api.RESTClient
	gql  *api.GraphQLClient
}

func (a *authed) IsAuthenticated() bool { return true }

const mergedPRsQuery = `
query($q: String!, $after: String) {
  search(query: $q, type: ISSUE, first: 100, after: $after) {
    issueCount
    pageInfo { hasNextPage endCursor }
    nodes { ... on PullRequest { number mergedAt repository { nameWithOwner } } }
  }
}`

func (a *authed) MergedPRs(user string) (MergedPRs, error) {
	type pr struct {
		Number     int
		MergedAt   string
		Repository struct{ NameWithOwner string }
	}
	var all []pr
	vars := map[string]interface{}{"q": "author:" + user + " is:pr is:merged"}
	total := 0
	for {
		var resp struct {
			Search struct {
				IssueCount int
				PageInfo   struct {
					HasNextPage bool
					EndCursor   string
				}
				Nodes []*pr
			}
		}
		if err := a.gql.Do(mergedPRsQuery, vars, &resp); err != nil {
			return MergedPRs{}, fmt.Errorf("listing merged pull requests: %w", err)
		}
		total = resp.Search.IssueCount
		for _, n := range resp.Search.Nodes {
			if n != nil && n.MergedAt != "" {
				all = append(all, *n)
			}
		}
		if !resp.Search.PageInfo.HasNextPage {
			break
		}
		vars["after"] = resp.Search.PageInfo.EndCursor
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].MergedAt < all[j].MergedAt })
	refs := make([]string, len(all))
	for i, p := range all {
		refs[i] = fmt.Sprintf("%s#%d", p.Repository.NameWithOwner, p.Number)
	}
	return MergedPRs{Total: total, Refs: refs}, nil
}

type repo struct {
	FullName        string `json:"full_name"`
	Fork            bool   `json:"fork"`
	StargazersCount int    `json:"stargazers_count"`
}

func maxStars(repos []repo) (int, string) {
	best, name := 0, ""
	for _, r := range repos {
		if !r.Fork && (name == "" || r.StargazersCount > best) {
			best, name = r.StargazersCount, r.FullName
		}
	}
	return best, name
}

func (a *authed) MaxOwnedStars(user string) (int, string, error) {
	var all []repo
	for page := 1; ; page++ {
		var repos []repo
		path := fmt.Sprintf("users/%s/repos?type=owner&per_page=100&page=%d", url.PathEscape(user), page)
		if err := a.rest.Get(path, &repos); err != nil {
			return 0, "", fmt.Errorf("listing repositories: %w", err)
		}
		all = append(all, repos...)
		if len(repos) < 100 {
			break
		}
	}
	stars, name := maxStars(all)
	return stars, name, nil
}

func (a *authed) RepoStars(fullName string) (int, error) {
	var r repo
	if err := a.rest.Get("repos/"+fullName, &r); err != nil {
		return 0, fmt.Errorf("reading %s: %w", fullName, err)
	}
	return r.StargazersCount, nil
}

func (a *authed) Facts(user string, flagFields []string) (*Facts, error) {
	query := `query($login: String!) { user(login: $login) {
	  repositoryDiscussionComments(onlyAnswers: true) { totalCount }
	  sponsoring { totalCount }`
	for _, f := range flagFields {
		query += " " + f
	}
	query += " } }"
	var resp struct {
		User map[string]json.RawMessage
	}
	if err := a.gql.Do(query, map[string]interface{}{"login": user}, &resp); err != nil {
		return nil, fmt.Errorf("reading user facts: %w", err)
	}
	var count struct{ TotalCount int }
	facts := &Facts{Flags: map[string]bool{}}
	if err := json.Unmarshal(resp.User["repositoryDiscussionComments"], &count); err != nil {
		return nil, fmt.Errorf("reading accepted answers: %w", err)
	}
	facts.AcceptedAnswers = count.TotalCount
	if err := json.Unmarshal(resp.User["sponsoring"], &count); err != nil {
		return nil, fmt.Errorf("reading sponsorships: %w", err)
	}
	facts.Sponsoring = count.TotalCount
	for _, f := range flagFields {
		var v bool
		if err := json.Unmarshal(resp.User[f], &v); err != nil {
			return nil, fmt.Errorf("reading %s: %w", f, err)
		}
		facts.Flags[f] = v
	}
	return facts, nil
}

// ---------- public (no token) ----------

type public struct {
	http *http.Client
	base string // https://api.github.com/
}

func newPublic() *public {
	return &public{http: &http.Client{Timeout: 30 * time.Second}, base: "https://api.github.com/"}
}

func (p *public) IsAuthenticated() bool { return false }

func (p *public) get(path string, v interface{}) error {
	req, err := http.NewRequest(http.MethodGet, p.base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := p.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		return ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: HTTP %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// ErrRateLimited is the unauthenticated API saying no; `gh auth login` lifts it.
var ErrRateLimited = errors.New("GitHub's API rate limit for anonymous use was hit; run `gh auth login`")

// MergedPRs without a token: the total only. Ordering by merge date needs GraphQL,
// so there is nothing to calibrate against.
func (p *public) MergedPRs(user string) (MergedPRs, error) {
	var resp struct {
		TotalCount int `json:"total_count"`
	}
	q := url.QueryEscape("author:" + user + " is:pr is:merged")
	if err := p.get("search/issues?per_page=1&q="+q, &resp); err != nil {
		return MergedPRs{}, fmt.Errorf("counting merged pull requests: %w", err)
	}
	return MergedPRs{Total: resp.TotalCount}, nil
}

func (p *public) MaxOwnedStars(user string) (int, string, error) {
	var all []repo
	for page := 1; ; page++ {
		var repos []repo
		if err := p.get(fmt.Sprintf("users/%s/repos?type=owner&per_page=100&page=%d", url.PathEscape(user), page), &repos); err != nil {
			return 0, "", fmt.Errorf("listing repositories: %w", err)
		}
		all = append(all, repos...)
		if len(repos) < 100 {
			break
		}
	}
	stars, name := maxStars(all)
	return stars, name, nil
}

func (p *public) RepoStars(fullName string) (int, error) {
	var r repo
	if err := p.get("repos/"+fullName, &r); err != nil {
		return 0, fmt.Errorf("reading %s: %w", fullName, err)
	}
	return r.StargazersCount, nil
}

func (p *public) Facts(string, []string) (*Facts, error) { return nil, nil }
