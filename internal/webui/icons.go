package webui

// GitHub deleted the artwork of the two badges it tested and switched off (its CDN
// answers 404), so these are our own icons: a dashed ring marks them as not GitHub's.
// They are served from our origin, which the CSP allows, and reused by the README.
const iconRing = `<circle cx="32" cy="32" r="29" fill="#eaeef2" stroke="#8c959f" stroke-width="2" stroke-dasharray="5 4"/>`

var ownIcons = map[string]string{
	"heart-on-your-sleeve": `<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64" role="img" aria-labelledby="t">` +
		`<title id="t">Heart On Your Sleeve (GitHub never released its artwork)</title>` + iconRing +
		`<path d="M32 47C19 38 14 31 18 24c3.5-6 10.5-5.5 14 0 3.5-5.5 10.5-6 14 0 4 7-1 14-14 23z" fill="#d1677f"/></svg>` + "\n",
	"open-sourcerer": `<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64" role="img" aria-labelledby="t">` +
		`<title id="t">Open Sourcerer (GitHub never released its artwork)</title>` + iconRing +
		`<path d="M20 45 39 26" stroke="#57606a" stroke-width="5" stroke-linecap="round"/>` +
		`<path d="M43 14l2 6 6 2-6 2-2 6-2-6-6-2 6-2z" fill="#bf8700"/>` +
		`<path d="M26 18l1 3 3 1-3 1-1 3-1-3-3-1 3-1z" fill="#bf8700"/>` +
		`<path d="M48 33l1 2.5 2.5 1-2.5 1-1 2.5-1-2.5-2.5-1 2.5-1z" fill="#bf8700"/></svg>` + "\n",
}

// OwnIcon is the path of our icon for a badge without GitHub artwork, or "".
func OwnIcon(slug string) string {
	if _, ok := ownIcons[slug]; !ok {
		return ""
	}
	return BasePath + "icons/" + slug + ".svg"
}

// OwnIconSVG is that icon's SVG, or nil.
func OwnIconSVG(slug string) []byte {
	s, ok := ownIcons[slug]
	if !ok {
		return nil
	}
	return []byte(s)
}
