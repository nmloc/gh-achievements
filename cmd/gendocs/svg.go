package main

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

var reANSI = regexp.MustCompile("\x1b\\[([0-9;]*)m")

// The site's dark palette: muted text, bronze, silver and gold metals.
var ansiFill = map[string]string{
	"2": "#93a9b7", "38;5;173": "#d39468", "38;5;250": "#c7d0d6", "38;5;178": "#e2b53a",
}

func plain(s string) string { return reANSI.ReplaceAllString(s, "") }

// keepDemoLines keeps the whole report except the legend line and trailing blank
// lines; the new layout already folds the archive into one line.
func keepDemoLines(out string) []string {
	var kept []string
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.HasPrefix(strings.TrimSpace(plain(line)), "official =") {
			continue
		}
		kept = append(kept, line)
	}
	for len(kept) > 0 && strings.TrimSpace(plain(kept[len(kept)-1])) == "" {
		kept = kept[:len(kept)-1]
	}
	return kept
}

// terminalSVG draws colored CLI output as a terminal window.
func terminalSVG(out, _ string) []byte {
	lines := keepDemoLines(out)
	const (
		fontSize = 14
		lineH    = 20
		charW    = 8.43
		padX     = 20
		top      = 48
	)
	cols := 0
	for _, l := range lines {
		cols = max(cols, utf8.RuneCountInString(plain(l)))
	}
	width := int(float64(cols)*charW) + 2*padX
	height := top + len(lines)*lineH + 16

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-labelledby="t">`+"\n", width, height, width, height)
	b.WriteString(`<title id="t">gh achievements torvalds: terminal output</title>` + "\n")
	fmt.Fprintf(&b, `<rect width="%d" height="%d" rx="10" fill="#0d1c26"/>`+"\n", width, height)
	b.WriteString(`<circle cx="20" cy="20" r="6" fill="#ff5f56"/><circle cx="40" cy="20" r="6" fill="#ffbd2e"/><circle cx="60" cy="20" r="6" fill="#27c93f"/>` + "\n")
	fmt.Fprintf(&b, `<text x="%d" y="24" fill="#93a9b7" font-family="ui-monospace,SFMono-Regular,Menlo,Consolas,monospace" font-size="13" text-anchor="middle">gh achievements torvalds</text>`+"\n", width/2)
	fmt.Fprintf(&b, `<g font-family="ui-monospace,SFMono-Regular,Menlo,Consolas,'Liberation Mono',monospace" font-size="%d" fill="#eaf1f5">`+"\n", fontSize)
	for i, line := range lines {
		fmt.Fprintf(&b, `<text x="%d" y="%d">%s</text>`+"\n", padX, top+i*lineH, spans(line))
	}
	b.WriteString("</g>\n</svg>\n")
	return []byte(b.String())
}

// spans turns one ANSI-colored line into SVG tspans.
func spans(line string) string {
	var b strings.Builder
	fill, bold := "", false
	pos := 0
	emit := func(text string) {
		if text == "" {
			return
		}
		// SVG collapses ordinary spaces whatever xml:space or CSS says once the
		// image is loaded through <img>; no-break spaces keep the columns aligned.
		text = strings.ReplaceAll(text, " ", "\u00a0")
		attrs := ""
		if fill != "" {
			attrs += ` fill="` + fill + `"`
		}
		if bold {
			attrs += ` font-weight="700"`
		}
		if attrs == "" {
			b.WriteString(html.EscapeString(text))
		} else {
			b.WriteString("<tspan" + attrs + ">" + html.EscapeString(text) + "</tspan>")
		}
	}
	for _, m := range reANSI.FindAllStringSubmatchIndex(line, -1) {
		emit(line[pos:m[0]])
		code := line[m[2]:m[3]]
		switch code {
		case "0", "":
			fill, bold = "", false
		case "1":
			bold = true
		default:
			fill = ansiFill[code]
		}
		pos = m[1]
	}
	emit(line[pos:])
	return b.String()
}

// socialSVG is the 1280×640 social preview (GitHub repository card, og:image).
func socialSVG() []byte {
	const mono = `font-family="DejaVu Sans Mono,Menlo,Consolas,monospace"`
	const sans = `font-family="DejaVu Sans,Helvetica,Arial,sans-serif"`
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="1280" height="640" viewBox="0 0 1280 640">` + "\n")
	b.WriteString(`<rect width="1280" height="640" fill="#0d1c26"/>` + "\n")
	b.WriteString(`<rect x="0" y="0" width="1280" height="8" fill="#e2b53a"/>` + "\n")
	fmt.Fprintf(&b, `<text x="80" y="150" %s font-size="76" font-weight="700" fill="#eaf1f5">gh-achievements</text>`+"\n", sans)
	fmt.Fprintf(&b, `<text x="80" y="225" %s font-size="38" fill="#eaf1f5">Check your GitHub achievements progress</text>`+"\n", sans)
	fmt.Fprintf(&b, `<text x="80" y="280" %s font-size="28" fill="#93a9b7">Every badge · your tier · how far to the next · verified against GitHub · no token</text>`+"\n", sans)
	b.WriteString(`<rect x="80" y="330" width="1120" height="230" rx="14" fill="#122633" stroke="#25414f"/>` + "\n")
	const ink = "#eaf1f5"
	rows := []struct{ mark, fill, name, rest string }{
		{"●", "#d39468", "Pull Shark", "x2  72 / 128 to x3  calibrated"},
		{"●", "#e2b53a", "Starstruck", "x4  Top tier, 252,192 stars  counted"},
		{"●", "#e2b53a", "Pair Extraord.", "x4  Top tier  official"},
		{"●", ink, "Arctic Code Vault", "✓   Since 2020"},
	}
	for i, r := range rows {
		y := 385 + i*48
		fmt.Fprintf(&b, `<text x="110" y="%d" %s font-size="26" fill="%s">%s</text>`+"\n", y, mono, r.fill, r.mark)
		fmt.Fprintf(&b, `<text x="170" y="%d" %s font-size="26" fill="%s" font-weight="700">%s</text>`+"\n", y, mono, ink, html.EscapeString(r.name))
		fmt.Fprintf(&b, `<text x="470" y="%d" %s font-size="26" fill="%s">%s</text>`+"\n", y, mono, ink, strings.ReplaceAll(html.EscapeString(r.rest), " ", "\u00a0"))
	}
	fmt.Fprintf(&b, `<text x="80" y="605" %s font-size="26" fill="#e2b53a">$ gh extension install nmloc/gh-achievements</text>`+"\n", mono)
	b.WriteString("</svg>\n")
	return []byte(b.String())
}

// checkButtonSVG is the README's call to action. It is self-contained vector art:
// GitHub loads README SVGs as images, so external images and web fonts never load.
func checkButtonSVG() []byte {
	const sans = `font-family="-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif"`
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="520" height="84" viewBox="0 0 520 84" role="img" aria-labelledby="t">` + "\n")
	b.WriteString(`<title id="t">Check your GitHub achievements: no install, sign in to include private work</title>` + "\n")
	b.WriteString(`<rect x="1" y="1" width="518" height="82" rx="14" fill="#142430" stroke="#e2b53a" stroke-width="2"/>` + "\n")
	// A trophy: cup, two handles, stem and base.
	b.WriteString(`<g transform="translate(26 18)" fill="#e2b53a" stroke="#e2b53a" stroke-width="3" stroke-linejoin="round">` +
		`<path d="M10 4h28v14a14 14 0 0 1-28 0z"/>` +
		`<path d="M10 9H4v4a8 8 0 0 0 7 8" fill="none"/><path d="M38 9h6v4a8 8 0 0 1-7 8" fill="none"/>` +
		`<path d="M21 32h6v7h-6z"/><path d="M14 41h20v5H14z"/></g>` + "\n")
	fmt.Fprintf(&b, `<text x="90" y="40" %s font-size="25" font-weight="700" textLength="400" lengthAdjust="spacingAndGlyphs" fill="#eaf1f5">Check your GitHub achievements →</text>`+"\n", sans)
	fmt.Fprintf(&b, `<text x="90" y="64" %s font-size="15" fill="#c7d0d6">No install · sign in with GitHub to include private work</text>`+"\n", sans)
	b.WriteString("</svg>\n")
	return []byte(b.String())
}
