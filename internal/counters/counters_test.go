package counters

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/config"
)

func TestMaxStars_ignoresForks(t *testing.T) {
	stars, name := maxStars([]repo{
		{FullName: "me/fork", Fork: true, StargazersCount: 900},
		{FullName: "me/small", StargazersCount: 3},
		{FullName: "me/big", StargazersCount: 40},
	})
	if stars != 40 || name != "me/big" {
		t.Fatalf("maxStars = %d %s", stars, name)
	}
	if stars, name := maxStars(nil); stars != 0 || name != "" {
		t.Fatalf("maxStars(nil) = %d %q", stars, name)
	}
}

func TestMergedPRs_IsComplete(t *testing.T) {
	if !(MergedPRs{Total: 2, Refs: []string{"a", "b"}}).IsComplete() {
		t.Error("full listing reported incomplete")
	}
	if (MergedPRs{Total: 2241, Refs: make([]string, 1000)}).IsComplete() {
		t.Error("capped listing reported complete")
	}
}

func publicAgainst(t *testing.T, h http.HandlerFunc) *public {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &public{http: srv.Client(), base: srv.URL + "/"}
}

func TestPublic_mergedPRsTotal(t *testing.T) {
	p := publicAgainst(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/issues" || r.URL.Query().Get("q") != "author:octo is:pr is:merged" {
			t.Errorf("unexpected request %s", r.URL)
		}
		_, _ = w.Write([]byte(`{"total_count": 42, "items": []}`))
	})
	prs, err := p.MergedPRs("octo")
	if err != nil || prs.Total != 42 || len(prs.Refs) != 0 || prs.IsComplete() {
		t.Fatalf("MergedPRs = %+v, %v", prs, err)
	}
}

func TestPublic_rateLimitTellsTheUserToSignIn(t *testing.T) {
	p := publicAgainst(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	if _, err := p.MergedPRs("octo"); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
}

func TestPublic_hasNoGraphQLFacts(t *testing.T) {
	if f, err := newPublic().Facts("octo", nil); f != nil || err != nil {
		t.Fatalf("Facts = %+v, %v", f, err)
	}
}

func TestNewWithToken_isAuthenticatedAndNeedsAToken(t *testing.T) {
	if _, err := NewWithToken(""); err == nil {
		t.Fatal("empty token accepted")
	}
	s, err := NewWithToken("gho_test")
	if err != nil || !s.IsAuthenticated() {
		t.Fatalf("NewWithToken = %v, %v", s, err)
	}
}

// go-gh reads the gh config once per process (config.Read is guarded by a sync.Once), so
// a GH_CONFIG_DIR set inside one test is ignored when an earlier test already read the
// real config. Replacing the config.Read seam makes the test independent of run order.
func TestNewWithToken_ignoresTheGhConfig(t *testing.T) {
	const badAPIHost = "github.com:\n    api_host: \"not a host!\"\n"
	original := config.Read
	t.Cleanup(func() { config.Read = original })
	config.Read = func(*config.Config) (*config.Config, error) {
		return config.ReadFromString("hosts:\n  " + badAPIHost), nil
	}
	if _, err := NewWithToken("gho_test"); err != nil {
		t.Fatalf("NewWithToken read the gh config: %v", err)
	}
}

// roundTripFunc answers every request without a network call.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// GH_DEBUG=api makes go-gh write a request log, with headers, to stderr. go-gh's httpretty
// sanitizer redacts Authorization today, so the token is not printed; the guarantee tested
// here is that a signed-in user's requests write no log at all, so no later change to the
// sanitizer can leak the token through this path.
func TestNewWithToken_writesNoRequestLogEvenWithGHDebug(t *testing.T) {
	const token = "gho_SECRET_FOR_TEST"
	t.Setenv("GH_DEBUG", "api")

	logPath := filepath.Join(t.TempDir(), "stderr")
	logFile, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	originalStderr := os.Stderr
	os.Stderr = logFile
	t.Cleanup(func() {
		os.Stderr = originalStderr
		_ = logFile.Close()
	})

	// NewWithToken hands http.DefaultTransport to go-gh, so faking it keeps the call offline.
	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"stargazers_count": 1}`)),
			Request:    req,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	s, err := NewWithToken(token)
	if err != nil {
		t.Fatalf("NewWithToken: %v", err)
	}
	if _, err := s.RepoStars("octo/repo"); err != nil {
		t.Fatalf("RepoStars: %v", err)
	}

	logged, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(logged) != 0 {
		t.Fatalf("GH_DEBUG wrote a request log for a user token (token %q must never reach it):\n%s", token, logged)
	}
}
