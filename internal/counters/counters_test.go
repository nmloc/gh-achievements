package counters

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
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
