// Package handler is the Vercel function behind every page requested with
// Accept: text/markdown (a vercel.json route).
package handler

import (
	"net/http"

	"github.com/nmloc/gh-achievements/web"
)

func Handler(w http.ResponseWriter, r *http.Request) { web.Markdown(w, r) }
