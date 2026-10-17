// internal/web/handlers.go
package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nmloc/gh-achievements/internal/counters"
	"github.com/nmloc/gh-achievements/internal/official"
	"github.com/nmloc/gh-achievements/internal/report"
)

// StateCookie carries the signed sign-in state between /auth/login and /api/private.
const StateCookie = "gha_oauth"

var errTimeout = errors.New("report took too long")

// Config is everything the handlers need. Tests replace GitHub, Official, NewCounts and Now.
type Config struct {
	ClientID, ClientSecret, BotToken string
	StateSecret                      []byte
	BaseURL                          string // https://gh-achievements.vercel.app or http://localhost:8787
	GitHub                           GitHub
	Official                         report.Official
	NewCounts                        func(token string) (counters.Source, error)
	Timeout                          time.Duration
	Now                              func() time.Time
	Log                              *log.Logger
}

// FromEnv reads the Vercel (or .env.local) environment. It never logs the values.
func FromEnv() (*Config, error) {
	c := &Config{
		ClientID: os.Getenv("GITHUB_CLIENT_ID"), ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		BotToken: os.Getenv("GH_BOT_TOKEN"), StateSecret: []byte(os.Getenv("STATE_SECRET")),
		BaseURL: strings.TrimSuffix(os.Getenv("BASE_URL"), "/"),
		GitHub:  DefaultGitHub(), Official: official.NewClient(), NewCounts: counters.NewWithToken,
		Timeout: 50 * time.Second, Now: time.Now, Log: log.New(os.Stderr, "gh-achievements ", 0),
	}
	if c.BaseURL == "" && os.Getenv("VERCEL_PROJECT_PRODUCTION_URL") != "" {
		c.BaseURL = "https://" + os.Getenv("VERCEL_PROJECT_PRODUCTION_URL")
	}
	var missing []string
	for name, v := range map[string]string{"GITHUB_CLIENT_ID": c.ClientID, "GITHUB_CLIENT_SECRET": c.ClientSecret,
		"GH_BOT_TOKEN": c.BotToken, "STATE_SECRET": string(c.StateSecret), "BASE_URL": c.BaseURL} {
		if v == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing environment: %s", strings.Join(missing, ", "))
	}
	return c, nil
}

var (
	envOnce sync.Once
	envCfg  *Config
	envErr  error
)

// Serve is what each Vercel function calls.
func Serve(w http.ResponseWriter, r *http.Request, h func(*Config, http.ResponseWriter, *http.Request)) {
	envOnce.Do(func() { envCfg, envErr = FromEnv() })
	if envErr != nil {
		log.Printf("gh-achievements: %v", envErr) // names only, never values
		setSecurityHeaders(w)
		http.Error(w, "The site is misconfigured. Please open an issue.", http.StatusInternalServerError)
		return
	}
	h(envCfg, w, r)
}

// build runs report.Build with a deadline. On timeout the goroutine finishes on its own;
// the function instance ends soon after.
func (c *Config) build(ctx context.Context, user, token string) (*report.Report, error) {
	counts, err := c.NewCounts(token)
	if err != nil {
		return nil, err
	}
	type result struct {
		r   *report.Report
		err error
	}
	done := make(chan result, 1)
	go func() {
		r, err := report.Build(user, c.Official, counts)
		done <- result{r, err}
	}()
	select {
	case res := <-done:
		return res.r, res.err
	case <-time.After(c.Timeout):
		return nil, errTimeout
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Result serves GET /result?u=<login>: public data, cached at the CDN, shareable.
func (c *Config) Result(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	user := r.URL.Query().Get("u")
	fail := func(status int, reason, msg string) {
		w.Header().Set("Cache-Control", "public, s-maxage=60")
		body, err := RenderErrorPage(status, reason, msg, user)
		if err != nil { // Render can return partial bytes with an error: never write them
			c.Log.Printf("result error page: %T", err)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(msg))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}
	if !official.IsLogin(user) {
		fail(http.StatusBadRequest, "invalid", "That is not a GitHub username.")
		return
	}
	rep, err := c.build(r.Context(), user, c.BotToken)
	switch {
	case errors.Is(err, official.ErrUserNotFound):
		fail(http.StatusNotFound, "not_found", fmt.Sprintf("GitHub has no user called %s.", user))
		return
	case errors.Is(err, errTimeout):
		fail(http.StatusGatewayTimeout, "timeout", "This account is too large to check on the web in time: use the CLI, gh extension install nmloc/gh-achievements.")
		return
	case err != nil:
		c.Log.Printf("result: %T", err)
		fail(http.StatusBadGateway, "github", "GitHub did not answer. Please try again in a minute.")
		return
	}
	body, err := RenderResultPage(NewResultView(rep, false))
	if err != nil {
		c.Log.Printf("result render: %T", err)
		fail(http.StatusInternalServerError, "internal", "Something broke on our side. Please try again.")
		return
	}
	w.Header().Set("Cache-Control", "public, s-maxage=3600, stale-while-revalidate=86400")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body)
}

// Login serves GET /auth/login: a signed state cookie, then GitHub's consent screen.
func (c *Config) Login(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	w.Header().Set("Cache-Control", "no-store")
	claims, err := NewState(c.Now())
	if err != nil {
		http.Error(w, "Could not start sign-in.", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: StateCookie, Value: SignState(c.StateSecret, claims), Path: "/", MaxAge: int(StateMaxAge.Seconds()),
		HttpOnly: true, Secure: strings.HasPrefix(c.BaseURL, "https://"), SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, c.GitHub.Authorize(c.ClientID, c.BaseURL+"/auth/callback", claims.Nonce, Challenge(claims.Verifier)), http.StatusFound)
}

// Private serves POST /api/private from callback.js: code → token → report → revoke.
func (c *Config) Private(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", "noindex")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fail := func(status int, reason string) {
		w.Header().Set("X-Error-Reason", reason) // callback.js shows the matching message
		w.WriteHeader(status)
	}
	http.SetCookie(w, &http.Cookie{Name: StateCookie, Value: "", Path: "/", MaxAge: -1})
	cookie, err := r.Cookie(StateCookie)
	if err != nil {
		fail(http.StatusBadRequest, "state")
		return
	}
	claims, err := VerifyState(c.StateSecret, cookie.Value, r.PostFormValue("state"), c.Now())
	if err != nil {
		fail(http.StatusBadRequest, "state")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), c.Timeout+5*time.Second)
	defer cancel()
	token, err := c.GitHub.Exchange(ctx, c.ClientID, c.ClientSecret, r.PostFormValue("code"), c.BaseURL+"/auth/callback", claims.Verifier)
	if err != nil {
		c.Log.Printf("private: %v", err) // ErrExchange text and a code such as bad_verification_code only
		fail(http.StatusBadGateway, "exchange")
		return
	}
	defer func() {
		// A fresh context: revoke even when the request context is already done.
		rctx, rcancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer rcancel()
		if err := c.GitHub.Revoke(rctx, c.ClientID, c.ClientSecret, token); err != nil {
			c.Log.Printf("private: %v", err)
		}
		token = ""
	}()
	user, err := c.GitHub.Login(ctx, token)
	if err != nil {
		c.Log.Printf("private: %v", err)
		fail(http.StatusBadGateway, "github")
		return
	}
	rep, err := c.build(ctx, user, token)
	switch {
	case errors.Is(err, errTimeout):
		fail(http.StatusGatewayTimeout, "timeout")
		return
	case err != nil:
		c.Log.Printf("private: %T", err)
		fail(http.StatusInternalServerError, "internal")
		return
	}
	body, err := RenderResultFragment(NewResultView(rep, true))
	if err != nil {
		c.Log.Printf("private render: %T", err)
		fail(http.StatusInternalServerError, "internal")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body)
}
