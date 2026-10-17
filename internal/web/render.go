// internal/web/render.go
package web

import (
	"github.com/nmloc/gh-achievements/internal/webui"
)

// RenderResultFragment is the result section alone, for /api/private.
func RenderResultFragment(v ResultView) ([]byte, error) { return webui.RenderFragment("result", v) }

// RenderResultPage is the full /result page.
func RenderResultPage(v ResultView) ([]byte, error) {
	return webui.Render(webui.Page{
		Path: "result", URL: webui.SiteURL + "result", Title: "@" + v.User + "'s GitHub achievements · gh-achievements",
		Description: "Every GitHub achievement, @" + v.User + "'s tier and how far to the next one.",
		Heading:     "GitHub achievements", Content: "result-page.html", NoIndex: true, Data: v,
	})
}

// ErrorView is shown instead of a result.
type ErrorView struct {
	Status          int
	Reason, Message string
	User            string // pre-fills the username box
}

// RenderErrorPage renders the check form with one message.
func RenderErrorPage(status int, reason, message, user string) ([]byte, error) {
	return webui.Render(webui.Page{
		Path: "result", URL: webui.SiteURL + "check/", Title: "Check your GitHub achievements · gh-achievements",
		Description: message, Heading: "Check your GitHub achievements", Content: "check.html", NoIndex: true,
		Data: ErrorView{Status: status, Reason: reason, Message: message, User: user},
	})
}
