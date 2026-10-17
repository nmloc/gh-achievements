package main

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

var reANSI = regexp.MustCompile("\x1b\\[([0-9;]*)m")

// GitHub's own tier label colours, plus a dark terminal palette.
var ansiFill = map[string]string{
	"2": "#8b949e", "32": "#3fb950",
	"38;5;216": "#f9bfa7", "38;5;252": "#e1e4e4", "38;5;221": "#fae57e",
}

// demoSections are the report sections the README demo keeps; the rest are summarised
// in one line, so the image stays readable.
var demoSections = map[string]bool{"EARNABLE": true, "RETIRED": true}

// sectionWords start the report's section headings (render.Text).
var sectionWords = map[string]bool{
	"EARNABLE": true, "DISABLED": true, "RETIRED": true, "INTERNAL": true, "NEW": true, "HIGHLIGHTS": true,
}

func plain(s string) string { return reANSI.ReplaceAllString(s, "") }

func keepDemoLines(out string) []string {
	var kept []string
	keep := true
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		p := strings.TrimSpace(plain(line))
		if f := strings.Fields(p); len(f) > 0 && sectionWords[f[0]] {
			keep = demoSections[f[0]]
		}
		if strings.HasPrefix(p, "official =") {
			keep = false
		}
		if keep {
			kept = append(kept, line)
		}
	}
	for len(kept) > 0 && plain(kept[len(kept)-1]) == "" {
		kept = kept[:len(kept)-1]
	}
	kept = append(kept, "", "  \x1b[2m… plus disabled and internal badges, and profile highlights\x1b[0m")
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
	fmt.Fprintf(&b, `<rect width="%d" height="%d" rx="10" fill="#0d1117"/>`+"\n", width, height)
	b.WriteString(`<circle cx="20" cy="20" r="6" fill="#ff5f56"/><circle cx="40" cy="20" r="6" fill="#ffbd2e"/><circle cx="60" cy="20" r="6" fill="#27c93f"/>` + "\n")
	fmt.Fprintf(&b, `<text x="%d" y="24" fill="#8b949e" font-family="ui-monospace,SFMono-Regular,Menlo,Consolas,monospace" font-size="13" text-anchor="middle">gh achievements torvalds</text>`+"\n", width/2)
	fmt.Fprintf(&b, `<g font-family="ui-monospace,SFMono-Regular,Menlo,Consolas,'Liberation Mono',monospace" font-size="%d" fill="#e6edf3">`+"\n", fontSize)
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
	b.WriteString(`<rect width="1280" height="640" fill="#0d1117"/>` + "\n")
	b.WriteString(`<rect x="0" y="0" width="1280" height="8" fill="#fae57e"/>` + "\n")
	fmt.Fprintf(&b, `<text x="80" y="150" %s font-size="76" font-weight="700" fill="#e6edf3">gh-achievements</text>`+"\n", sans)
	fmt.Fprintf(&b, `<text x="80" y="225" %s font-size="38" fill="#e6edf3">Check your GitHub achievements progress</text>`+"\n", sans)
	fmt.Fprintf(&b, `<text x="80" y="280" %s font-size="28" fill="#8b949e">Every badge · your tier · how far to the next · verified against GitHub · no token</text>`+"\n", sans)
	b.WriteString(`<rect x="80" y="330" width="1120" height="230" rx="14" fill="#161b22" stroke="#30363d"/>` + "\n")
	rows := []struct{ mark, fill, name, rest, restFill string }{
		{"x2", "#f9bfa7", "Pull Shark      ", "72 / 128 ██████░░░░░░ 56 to x3 [calibrated]", "#e6edf3"},
		{"x4", "#fae57e", "Starstruck      ", "252192  max tier  [counted]", "#e6edf3"},
		{"x4", "#fae57e", "Pair Extraord.  ", "max tier  [official]", "#e6edf3"},
		{"x1", "#3fb950", "Arctic Code Vault", " retired · since 2020-02-02", "#8b949e"},
	}
	for i, r := range rows {
		y := 385 + i*48
		fmt.Fprintf(&b, `<text x="110" y="%d" %s font-size="26" fill="%s" font-weight="700">%s</text>`+"\n", y, mono, r.fill, r.mark)
		fmt.Fprintf(&b, `<text x="170" y="%d" %s font-size="26" fill="#e6edf3" font-weight="700">%s</text>`+"\n", y, mono, html.EscapeString(r.name))
		fmt.Fprintf(&b, `<text x="470" y="%d" %s font-size="26" fill="%s">%s</text>`+"\n", y, mono, r.restFill, strings.ReplaceAll(html.EscapeString(r.rest), " ", "\u00a0"))
	}
	fmt.Fprintf(&b, `<text x="80" y="605" %s font-size="26" fill="#3fb950">$ gh extension install nmloc/gh-achievements</text>`+"\n", mono)
	b.WriteString("</svg>\n")
	return []byte(b.String())
}
