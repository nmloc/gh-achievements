// web/headers.go
package web

import "net/http"

// Header is one response header.
type Header struct{ Key, Value string }

// SecurityHeaders apply to every response. cmd/gendocs writes the same list into
// vercel.json for static files, so functions and static pages cannot drift.
func SecurityHeaders() []Header {
	return []Header{
		{"Content-Security-Policy", "default-src 'self'; " +
			"script-src 'self' https://www.googletagmanager.com; " +
			"connect-src 'self' https://*.google-analytics.com https://*.analytics.google.com https://www.googletagmanager.com; " +
			"img-src 'self' data: https://github.githubassets.com https://avatars.githubusercontent.com https://github.com https://*.google-analytics.com https://www.googletagmanager.com; " +
			"style-src 'self'; base-uri 'self'; object-src 'none'; form-action 'self' https://github.com; frame-ancestors 'none'"},
		{"X-Content-Type-Options", "nosniff"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Permissions-Policy", "camera=(), microphone=(), geolocation=()"},
	}
}

func setSecurityHeaders(w http.ResponseWriter) {
	for _, h := range SecurityHeaders() {
		w.Header().Set(h.Key, h.Value)
	}
}
