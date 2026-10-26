// Command devserver runs the site and the functions locally on http://localhost:8787,
// with credentials from .env.local (the dev OAuth App). It never prints them.
package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nmloc/gh-achievements/web"
)

const addr = "127.0.0.1:8787"

func loadEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	env := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
			v = v[1 : len(v)-1]
		}
		env[strings.TrimSpace(k)] = v
	}
	return env, s.Err()
}

func main() {
	env, err := loadEnv(".env.local")
	if err != nil {
		log.Fatalf("devserver: %v (create .env.local, see CONTRIBUTING.md)", err)
	}
	for k, v := range env {
		os.Setenv(k, v)
	}
	os.Setenv("BASE_URL", "http://localhost:8787")
	cfg, err := web.FromEnv()
	if err != nil {
		log.Fatalf("devserver: %v", err)
	}
	files := http.FileServer(http.Dir("site"))
	mux := http.NewServeMux()
	mux.HandleFunc("/result", cfg.Result)
	mux.HandleFunc("/auth/login", cfg.Login)
	mux.HandleFunc("/api/private", cfg.Private)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// The vercel.json Markdown route, which `vercel dev` would not apply (no `has` locally).
		if web.WantsMarkdown(r) {
			q := r.URL.Query()
			q.Set("path", r.URL.Path)
			r.URL.RawQuery = q.Encode()
			web.Markdown(w, r)
			return
		}
		w.Header().Set("Vary", "Accept")
		for _, h := range web.SecurityHeaders() {
			w.Header().Set(h.Key, h.Value)
		}
		if strings.HasPrefix(r.URL.Path, "/auth/callback") {
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Cache-Control", "no-store")
		}
		files.ServeHTTP(w, r)
	})
	log.Printf("devserver: http://localhost:8787 (dev OAuth App, site/ from disk)")
	log.Fatal(http.ListenAndServe(addr, mux))
}
