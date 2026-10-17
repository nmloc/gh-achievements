// web/handlers.go
package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nmloc/gh-achievements/internal/counters"
	"github.com/nmloc/gh-achievements/internal/official"
	"github.com/nmloc/gh-achievements/internal/report"
)

// StateCookie carries the signed sign-in state between /auth/login and /api/private.
const StateCookie = "gha_oauth"

var (
	errTimeout  = errors.New("report took too long")
	errInternal = errors.New("report could not run") // building the source failed, or the report panicked

	// maxPrivateBody caps POST /api/private: the form holds a code and a state, nothing else.
	maxPrivateBody int64 = 4 << 10
	// revokeTimeout is the revoke's own budget. Timeout (50s) + revokeTimeout stays under Vercel's 60s maxDuration.
	revokeTimeout = 8 * time.Second
)

// minStateSecret is the shortest STATE_SECRET FromEnv accepts (openssl rand -base64 32 gives 44).
const minStateSecret = 32

// statusClientClosed is nginx's 499: the client left before we answered. No body is written.
const statusClientClosed = 499

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
		sort.Strings(missing)
		return nil, fmt.Errorf("missing environment: %s", strings.Join(missing, ", "))
	}
	if len(c.StateSecret) < minStateSecret {
		return nil, fmt.Errorf("STATE_SECRET must be at least %d bytes (openssl rand -base64 32)", minStateSecret)
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

// build runs report.Build under ctx's deadline. On timeout the goroutine finishes on its own;
// the function instance ends soon after. A panic in the report is contained here: it would
// otherwise end the process and skip every in-flight request's deferred revoke.
func (c *Config) build(ctx context.Context, user, token string) (*report.Report, error) {
	counts, err := c.NewCounts(token)
	if err != nil {
		return nil, fmt.Errorf("%w: %T", errInternal, err)
	}
	type result struct {
		r   *report.Report
		err error
	}
	done := make(chan result, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil { // the type only, never the value
				done <- result{nil, fmt.Errorf("%w: report panic: %T", errInternal, p)}
			}
		}()
		r, err := report.Build(user, c.Official, counts)
		done <- result{r, err}
	}()
	select {
	case res := <-done:
		return res.r, res.err
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, errTimeout
		}
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
	ctx, cancel := context.WithTimeout(r.Context(), c.Timeout)
	defer cancel()
	rep, err := c.build(ctx, user, c.BotToken)
	switch {
	case errors.Is(err, context.Canceled):
		c.Log.Printf("result: cancelled")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(statusClientClosed)
		return
	case errors.Is(err, errInternal):
		c.Log.Printf("result: %v", err) // build wraps types only, never values
		fail(http.StatusInternalServerError, "internal", "Something broke on our side. Please try again.")
		return
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
		HttpOnly: true, Secure: !isLocalhost(c.BaseURL), SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, c.GitHub.Authorize(c.ClientID, c.BaseURL+"/auth/callback", claims.Nonce, Challenge(claims.Verifier)), http.StatusFound)
}

// isLocalhost is true only for a local dev server, the one place a Secure cookie cannot be set.
func isLocalhost(baseURL string) bool {
	u, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	h := u.Hostname()
	return h == "localhost" || h == "127.0.0.1"
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
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/x-www-form-urlencoded" {
		fail(http.StatusUnsupportedMediaType, "internal")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPrivateBody)
	if err := r.ParseForm(); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			fail(http.StatusRequestEntityTooLarge, "internal")
			return
		}
		fail(http.StatusBadRequest, "state")
		return
	}
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
	// One deadline for the exchange, the user lookup and the report: whichever step is
	// slow, the user sees one timeout message and the revoke below still has its own time.
	ctx, cancel := context.WithTimeout(r.Context(), c.Timeout)
	defer cancel()
	// failCtx answers a dead ctx (timeout or client gone) and reports whether it did.
	failCtx := func(err error) bool {
		switch {
		case errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled):
			c.Log.Printf("private: cancelled")
			w.WriteHeader(statusClientClosed)
		case errors.Is(err, errTimeout) || errors.Is(ctx.Err(), context.DeadlineExceeded):
			fail(http.StatusGatewayTimeout, "timeout")
		default:
			return false
		}
		return true
	}
	token, err := c.GitHub.Exchange(ctx, c.ClientID, c.ClientSecret, r.PostFormValue("code"), c.BaseURL+"/auth/callback", claims.Verifier)
	if err != nil {
		if !failCtx(err) {
			c.Log.Printf("private: %v", err) // ErrExchange text and a code such as bad_verification_code only
			fail(http.StatusBadGateway, "exchange")
		}
		return
	}
	defer func() {
		// A fresh context: revoke even when the request context is already done.
		rctx, rcancel := context.WithTimeout(context.Background(), revokeTimeout)
		defer rcancel()
		if err := c.GitHub.Revoke(rctx, c.ClientID, c.ClientSecret, token); err != nil {
			c.Log.Printf("private: %v", err)
		}
		token = ""
	}()
	user, err := c.GitHub.Login(ctx, token)
	if err != nil {
		if !failCtx(err) {
			c.Log.Printf("private: %v", err)
			fail(http.StatusBadGateway, "github")
		}
		return
	}
	rep, err := c.build(ctx, user, token)
	switch {
	case err == nil:
	case failCtx(err):
		return
	case errors.Is(err, errInternal):
		c.Log.Printf("private: %v", err)
		fail(http.StatusInternalServerError, "internal")
		return
	default:
		c.Log.Printf("private: %T", err)
		fail(http.StatusBadGateway, "github")
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
