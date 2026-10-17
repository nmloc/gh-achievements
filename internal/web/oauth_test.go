package web

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

var secret = []byte("test-state-secret-0123456789")

func TestState_roundTrip(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	c, err := NewState(now)
	if err != nil || len(c.Nonce) < 20 || len(c.Verifier) < 43 {
		t.Fatalf("NewState = %+v, %v", c, err)
	}
	got, err := VerifyState(secret, SignState(secret, c), c.Nonce, now.Add(time.Minute))
	if err != nil || got != c {
		t.Fatalf("VerifyState = %+v, %v", got, err)
	}
}

func TestState_rejects(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	c, _ := NewState(now)
	good := SignState(secret, c)
	cases := map[string]struct {
		cookie, nonce string
		at            time.Time
	}{
		"missing cookie":   {"", c.Nonce, now},
		"tampered payload": {"x" + good, c.Nonce, now},
		"other secret":     {SignState([]byte("another-secret"), c), c.Nonce, now},
		"nonce mismatch":   {good, "not-the-nonce", now},
		"expired":          {good, c.Nonce, now.Add(StateMaxAge + time.Second)},
		"from the future":  {good, c.Nonce, now.Add(-2 * time.Minute)},
	}
	for name, tc := range cases {
		if _, err := VerifyState(secret, tc.cookie, tc.nonce, tc.at); !errors.Is(err, ErrState) {
			t.Errorf("%s: err = %v, want ErrState", name, err)
		}
	}
}

func TestChallenge_isS256(t *testing.T) {
	// RFC 7636 appendix B test vector.
	if got := Challenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"); got != "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM" {
		t.Fatalf("Challenge = %s", got)
	}
}

func TestAuthorize(t *testing.T) {
	u, _ := url.Parse(DefaultGitHub().Authorize("cid", "https://x.test/auth/callback", "nonce", "chal"))
	q := u.Query()
	if u.Host != "github.com" || q.Get("client_id") != "cid" || q.Get("scope") != "repo" || q.Get("state") != "nonce" ||
		q.Get("redirect_uri") != "https://x.test/auth/callback" || q.Get("code_challenge") != "chal" || q.Get("code_challenge_method") != "S256" {
		t.Fatalf("authorize URL = %s", u)
	}
}

// fakeGitHub is GitHub's token, /user and revoke endpoints.
type fakeGitHub struct {
	srv                   *httptest.Server
	revoked               []string
	verifier              string
	userFails             bool          // GET /user answers 500
	tokenDelay, userDelay time.Duration // how long the token and /user endpoints stall
}

func newFakeGitHub(t *testing.T) *fakeGitHub {
	f := &fakeGitHub{}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		time.Sleep(f.tokenDelay)
		f.verifier = r.Form.Get("code_verifier")
		w.Header().Set("Content-Type", "application/json")
		if r.Form.Get("code") != "good-code" || r.Form.Get("client_secret") != "csecret" {
			_, _ = io.WriteString(w, `{"error":"bad_verification_code"}`)
			return
		}
		_, _ = io.WriteString(w, `{"access_token":"gho_TEST_TOKEN","refresh_token":"ghr_REFRESH","token_type":"bearer"}`)
	})
	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(f.userDelay)
		if f.userFails {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if r.Header.Get("Authorization") != "Bearer gho_TEST_TOKEN" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = io.WriteString(w, `{"login":"brannon"}`)
	})
	mux.HandleFunc("DELETE /applications/cid/token", func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		var body struct {
			AccessToken string `json:"access_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if user != "cid" || pass != "csecret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		f.revoked = append(f.revoked, body.AccessToken)
		w.WriteHeader(http.StatusNoContent)
	})
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeGitHub) endpoints() GitHub {
	return GitHub{AuthorizeURL: f.srv.URL + "/login/oauth/authorize", TokenURL: f.srv.URL + "/login/oauth/access_token",
		APIURL: f.srv.URL, HTTP: f.srv.Client()}
}

func TestExchangeLoginRevoke(t *testing.T) {
	f := newFakeGitHub(t)
	g, ctx := f.endpoints(), context.Background()
	tok, err := g.Exchange(ctx, "cid", "csecret", "good-code", "https://x.test/auth/callback", "the-verifier")
	if err != nil || tok != "gho_TEST_TOKEN" || f.verifier != "the-verifier" {
		t.Fatalf("Exchange = %q, %v (verifier sent %q)", tok, err, f.verifier)
	}
	if login, err := g.Login(ctx, tok); err != nil || login != "brannon" {
		t.Fatalf("Login = %q, %v", login, err)
	}
	if err := g.Revoke(ctx, "cid", "csecret", tok); err != nil || len(f.revoked) != 1 || f.revoked[0] != tok {
		t.Fatalf("Revoke err=%v revoked=%v", err, f.revoked)
	}
}

func TestExchange_badCodeIsErrExchange_andHidesTheSecret(t *testing.T) {
	f := newFakeGitHub(t)
	_, err := f.endpoints().Exchange(context.Background(), "cid", "csecret", "bad", "r", "v")
	if !errors.Is(err, ErrExchange) || strings.Contains(err.Error(), "csecret") {
		t.Fatalf("err = %v", err)
	}
}

func TestExchange_neverFollowsRedirects(t *testing.T) {
	var leaked int
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leaked++
		_, _ = io.WriteString(w, `{"access_token":"gho_LEAK"}`)
	}))
	defer other.Close()
	token := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL, http.StatusTemporaryRedirect)
	}))
	defer token.Close()
	g := DefaultGitHub()
	g.TokenURL = token.URL
	_, err := g.Exchange(context.Background(), "cid", "csecret", "code", "r", "v")
	if !errors.Is(err, ErrExchange) || leaked != 0 {
		t.Fatalf("err = %v, second server hit %d times", err, leaked)
	}
}
