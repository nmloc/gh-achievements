package web

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// StateMaxAge is how long a sign-in may take between GitHub's consent screen and our callback.
const StateMaxAge = 10 * time.Minute

var (
	ErrState    = errors.New("sign-in state is missing, invalid or expired")
	ErrExchange = errors.New("GitHub did not accept the sign-in code")
)

var b64 = base64.RawURLEncoding

// StateClaims travel in an HMAC-signed, HttpOnly cookie between /auth/login and /api/private.
type StateClaims struct {
	Nonce    string `json:"n"`
	Verifier string `json:"v"` // PKCE code_verifier
	Issued   int64  `json:"t"`
}

func random(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return b64.EncodeToString(b), nil
}

// NewState makes a fresh nonce and PKCE verifier.
func NewState(now time.Time) (StateClaims, error) {
	nonce, err := random(18)
	if err != nil {
		return StateClaims{}, err
	}
	verifier, err := random(32) // 43 characters, the RFC 7636 minimum
	if err != nil {
		return StateClaims{}, err
	}
	return StateClaims{Nonce: nonce, Verifier: verifier, Issued: now.Unix()}, nil
}

func mac(secret []byte, payload string) string {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(payload))
	return b64.EncodeToString(m.Sum(nil))
}

// SignState encodes claims as payload.signature.
func SignState(secret []byte, c StateClaims) string {
	raw, _ := json.Marshal(c)
	payload := b64.EncodeToString(raw)
	return payload + "." + mac(secret, payload)
}

// VerifyState checks the signature, that nonce is the one GitHub echoed back, and the age.
func VerifyState(secret []byte, cookie, nonce string, now time.Time) (StateClaims, error) {
	payload, sig, ok := strings.Cut(cookie, ".")
	if !ok || !hmac.Equal([]byte(sig), []byte(mac(secret, payload))) {
		return StateClaims{}, ErrState
	}
	raw, err := b64.DecodeString(payload)
	if err != nil {
		return StateClaims{}, ErrState
	}
	var c StateClaims
	if json.Unmarshal(raw, &c) != nil || subtle.ConstantTimeCompare([]byte(c.Nonce), []byte(nonce)) != 1 {
		return StateClaims{}, ErrState
	}
	age := now.Sub(time.Unix(c.Issued, 0))
	if age < -time.Minute || age > StateMaxAge {
		return StateClaims{}, ErrState
	}
	return c, nil
}

// Challenge is the PKCE S256 code_challenge for verifier.
func Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return b64.EncodeToString(sum[:])
}

// GitHub holds the endpoints, so tests can point them at a fake.
type GitHub struct {
	AuthorizeURL, TokenURL, APIURL string
	HTTP                           *http.Client
}

// DefaultGitHub is github.com.
func DefaultGitHub() GitHub {
	return GitHub{
		AuthorizeURL: "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		APIURL:       "https://api.github.com",
		HTTP:         &http.Client{Timeout: 20 * time.Second},
	}
}

// Authorize is the URL of GitHub's consent screen.
func (g GitHub) Authorize(clientID, redirectURI, nonce, challenge string) string {
	q := url.Values{
		"client_id": {clientID}, "redirect_uri": {redirectURI}, "scope": {"repo"}, "state": {nonce},
		"code_challenge": {challenge}, "code_challenge_method": {"S256"}, "allow_signup": {"false"},
	}
	return g.AuthorizeURL + "?" + q.Encode()
}

// Exchange trades the code for a user token. A refresh token, if any, is ignored.
func (g GitHub) Exchange(ctx context.Context, clientID, secret, code, redirectURI, verifier string) (string, error) {
	form := url.Values{"client_id": {clientID}, "client_secret": {secret}, "code": {code},
		"redirect_uri": {redirectURI}, "code_verifier": {verifier}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := g.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w (network: %T)", ErrExchange, err)
	}
	defer resp.Body.Close()
	var body struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.AccessToken == "" {
		return "", fmt.Errorf("%w (%s)", ErrExchange, body.Error) // body.Error is a code like bad_verification_code
	}
	return body.AccessToken, nil
}

func (g GitHub) api(ctx context.Context, method, path string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, g.APIURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return req, nil
}

// Login reads the signed-in user's login.
func (g GitHub) Login(ctx context.Context, token string) (string, error) {
	req, err := g.api(ctx, http.MethodGet, "/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := g.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("reading the signed-in user: %T", err)
	}
	defer resp.Body.Close()
	var me struct{ Login string }
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&me) != nil || me.Login == "" {
		return "", fmt.Errorf("reading the signed-in user: HTTP %d", resp.StatusCode)
	}
	return me.Login, nil
}

// Revoke deletes the token. The app's grant stays, so the next sign-in skips consent.
func (g GitHub) Revoke(ctx context.Context, clientID, secret, token string) error {
	body, _ := json.Marshal(map[string]string{"access_token": token})
	req, err := g.api(ctx, http.MethodDelete, "/applications/"+url.PathEscape(clientID)+"/token", body)
	if err != nil {
		return err
	}
	req.SetBasicAuth(clientID, secret)
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("revoking the token: %T", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("revoking the token: HTTP %d", resp.StatusCode)
	}
	return nil
}
