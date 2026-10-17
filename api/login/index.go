// Package handler is the Vercel function behind /auth/login (vercel.json rewrite).
package handler

import (
	"net/http"

	"github.com/nmloc/gh-achievements/internal/web"
)

func Handler(w http.ResponseWriter, r *http.Request) { web.Serve(w, r, (*web.Config).Login) }
