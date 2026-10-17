// Package handler is the Vercel function at /api/private, called by callback.js.
package handler

import (
	"net/http"

	"github.com/nmloc/gh-achievements/web"
)

func Handler(w http.ResponseWriter, r *http.Request) { web.Serve(w, r, (*web.Config).Private) }
