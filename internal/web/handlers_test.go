// internal/web/handlers_test.go
package web

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/nmloc/gh-achievements/internal/counters"
	"github.com/nmloc/gh-achievements/internal/fixtures"
)

func testConfig(t *testing.T, gh *fakeGitHub) (*Config, *bytes.Buffer) {
	var logs bytes.Buffer
	c := &Config{
		ClientID: "cid", ClientSecret: "csecret", BotToken: "bot", StateSecret: secret,
		BaseURL: "https://x.test", Official: fixtures.Official{},
		NewCounts: func(token string) (counters.Source, error) { return fixtures.TorvaldsDemo(), nil },
		Timeout:   5 * time.Second, Now: func() time.Time { return time.Unix(1_800_000_000, 0) },
		Log: log.New(&logs, "", 0),
	}
	if gh != nil {
		c.GitHub = gh.endpoints()
	} else {
		c.GitHub = DefaultGitHub()
	}
	return c, &logs
}

func TestResult_publicReportIsCachedAndShareable(t *testing.T) {
	c, _ := testConfig(t, nil)
	w := httptest.NewRecorder()
	c.Result(w, httptest.NewRequest("GET", "/result?u=torvalds", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "@torvalds") {
		t.Fatalf("status %d", w.Code)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "public, s-maxage=3600, stale-while-revalidate=86400" {
		t.Fatalf("Cache-Control = %q", cc)
	}
	if w.Header().Get("Content-Security-Policy") == "" || w.Header().Get("Set-Cookie") != "" {
		t.Fatal("missing CSP or unexpected cookie")
	}
}

func TestResult_badAndUnknownUsernames(t *testing.T) {
	c, _ := testConfig(t, nil)
	for path, want := range map[string]int{"/result?u=a%20b": 400, "/result": 400, "/result?u=torvalds%3Fx": 400, "/result?u=nobody-here": 404} {
		w := httptest.NewRecorder()
		c.Result(w, httptest.NewRequest("GET", path, nil))
		if w.Code != want || !strings.Contains(w.Body.String(), `name="u"`) {
			t.Errorf("%s: status %d, want %d with the form", path, w.Code, want)
		}
		if !strings.HasPrefix(w.Header().Get("Cache-Control"), "public, s-maxage=60") {
			t.Errorf("%s: errors must be cached briefly, got %q", path, w.Header().Get("Cache-Control"))
		}
	}
}

func TestResult_tooSlowSaysUseTheCLI(t *testing.T) {
	c, _ := testConfig(t, nil)
	c.Timeout = 10 * time.Millisecond
	c.NewCounts = func(string) (counters.Source, error) { return slowCounts{fixtures.TorvaldsDemo()}, nil }
	w := httptest.NewRecorder()
	c.Result(w, httptest.NewRequest("GET", "/result?u=torvalds", nil))
	if w.Code != 504 || !strings.Contains(w.Body.String(), "use the CLI") {
		t.Fatalf("status %d", w.Code)
	}
}

type slowCounts struct{ *fixtures.Counts }

func (s slowCounts) MergedPRs(u string) (counters.MergedPRs, error) {
	time.Sleep(200 * time.Millisecond)
	return s.Counts.MergedPRs(u)
}

func TestLogin_setsSignedStateAndRedirects(t *testing.T) {
	c, _ := testConfig(t, nil)
	w := httptest.NewRecorder()
	c.Login(w, httptest.NewRequest("GET", "/auth/login", nil))
	if w.Code != http.StatusFound {
		t.Fatalf("status %d", w.Code)
	}
	loc, _ := url.Parse(w.Header().Get("Location"))
	q := loc.Query()
	if q.Get("redirect_uri") != "https://x.test/auth/callback" || q.Get("client_id") != "cid" {
		t.Fatalf("Location = %s", loc)
	}
	cookie := w.Result().Cookies()[0]
	if cookie.Name != StateCookie || !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge != 600 || cookie.Path != "/" {
		t.Fatalf("cookie = %+v", cookie)
	}
	claims, err := VerifyState(secret, cookie.Value, q.Get("state"), c.Now())
	if err != nil || q.Get("code_challenge") != Challenge(claims.Verifier) {
		t.Fatalf("state does not verify: %v", err)
	}
}

// privateRequest runs /auth/login, then posts the callback's code and state like callback.js does.
func privateRequest(t *testing.T, c *Config, code string, tamper func(cookie, state string) (string, string)) *httptest.ResponseRecorder {
	lw := httptest.NewRecorder()
	c.Login(lw, httptest.NewRequest("GET", "/auth/login", nil))
	loc, _ := url.Parse(lw.Header().Get("Location"))
	cookie, state := lw.Result().Cookies()[0].Value, loc.Query().Get("state")
	if tamper != nil {
		cookie, state = tamper(cookie, state)
	}
	form := url.Values{"code": {code}, "state": {state}}
	req := httptest.NewRequest("POST", "/api/private", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: StateCookie, Value: cookie})
	}
	w := httptest.NewRecorder()
	c.Private(w, req)
	return w
}

func TestPrivate_buildsWithTheUserTokenThenRevokesIt(t *testing.T) {
	gh := newFakeGitHub(t)
	c, logs := testConfig(t, gh)
	var tokenUsed string
	c.NewCounts = func(token string) (counters.Source, error) { tokenUsed = token; return fixtures.TorvaldsDemo(), nil }
	w := privateRequest(t, c, "good-code", nil)
	body := w.Body.String()
	if w.Code != 200 || !strings.Contains(body, "@brannon") || !strings.Contains(body, `data-scope="private"`) {
		t.Fatalf("status %d body %.200s", w.Code, body)
	}
	if tokenUsed != "gho_TEST_TOKEN" || len(gh.revoked) != 1 || gh.revoked[0] != "gho_TEST_TOKEN" {
		t.Fatalf("token used %q, revoked %v", tokenUsed, gh.revoked)
	}
	if gh.verifier == "" {
		t.Fatal("PKCE verifier not sent")
	}
	for name, h := range map[string]string{"Cache-Control": "no-store", "X-Robots-Tag": "noindex", "Referrer-Policy": "no-referrer"} {
		if w.Header().Get(name) != h {
			t.Errorf("%s = %q, want %q", name, w.Header().Get(name), h)
		}
	}
	if !strings.Contains(w.Header().Get("Set-Cookie"), StateCookie+"=; Path=/; Max-Age=0") {
		t.Errorf("state cookie not cleared: %q", w.Header().Get("Set-Cookie"))
	}
	for _, leak := range []string{"gho_TEST_TOKEN", "ghr_REFRESH", "csecret"} {
		if strings.Contains(body, leak) || strings.Contains(logs.String(), leak) || strings.Contains(w.Header().Get("Set-Cookie"), leak) {
			t.Errorf("%s leaked", leak)
		}
	}
}

func TestPrivate_revokesEvenWhenTheReportFails(t *testing.T) {
	gh := newFakeGitHub(t)
	c, logs := testConfig(t, gh)
	c.NewCounts = func(string) (counters.Source, error) { return nil, errors.New("boom") }
	w := privateRequest(t, c, "good-code", nil)
	if w.Code != 500 || w.Header().Get("X-Error-Reason") != "internal" || len(gh.revoked) != 1 {
		t.Fatalf("status %d reason %q revoked %v", w.Code, w.Header().Get("X-Error-Reason"), gh.revoked)
	}
	// The failure path logs too: it must not carry the token, the secret or the error text.
	for _, leak := range []string{"gho_TEST_TOKEN", "ghr_REFRESH", "csecret", "boom"} {
		if strings.Contains(logs.String(), leak) || strings.Contains(w.Body.String(), leak) {
			t.Errorf("%s leaked on the failure path", leak)
		}
	}
}

func TestPrivate_rejectsBadState(t *testing.T) {
	gh := newFakeGitHub(t)
	c, _ := testConfig(t, gh)
	cases := map[string]func(cookie, state string) (string, string){
		"no cookie":   func(_, s string) (string, string) { return "", s },
		"other nonce": func(c, _ string) (string, string) { return c, "other" },
		"forged cookie": func(_, s string) (string, string) {
			return SignState([]byte("x"), StateClaims{Nonce: s, Issued: 1_800_000_000}), s
		},
	}
	for name, tamper := range cases {
		w := privateRequest(t, c, "good-code", tamper)
		if w.Code != 400 || w.Header().Get("X-Error-Reason") != "state" {
			t.Errorf("%s: status %d reason %q", name, w.Code, w.Header().Get("X-Error-Reason"))
		}
	}
	if len(gh.revoked) != 0 {
		t.Fatal("no token should exist to revoke")
	}
}

func TestPrivate_badCodeAndWrongMethod(t *testing.T) {
	gh := newFakeGitHub(t)
	c, _ := testConfig(t, gh)
	if w := privateRequest(t, c, "bad-code", nil); w.Code != 502 || w.Header().Get("X-Error-Reason") != "exchange" {
		t.Fatalf("bad code: %d %q", w.Code, w.Header().Get("X-Error-Reason"))
	}
	w := httptest.NewRecorder()
	c.Private(w, httptest.NewRequest("GET", "/api/private", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET: %d", w.Code)
	}
}

func TestSecurityHeaders_cspAllowsWhatThePagesLoad(t *testing.T) {
	var csp string
	for _, h := range SecurityHeaders() {
		if h.Key == "Content-Security-Policy" {
			csp = h.Value
		}
	}
	for _, want := range []string{"script-src 'self' https://www.googletagmanager.com", "https://github.githubassets.com",
		"https://avatars.githubusercontent.com", "frame-ancestors 'none'", "form-action 'self' https://github.com"} {
		if !strings.Contains(csp, want) {
			t.Errorf("CSP lacks %s", want)
		}
	}
	if strings.Contains(csp, "unsafe-inline") {
		t.Error("CSP allows inline code")
	}
}
