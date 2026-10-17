// web/render_test.go
package web

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

func golden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		_ = os.MkdirAll("testdata", 0o755)
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run with -update)", err)
	}
	if string(want) != string(got) {
		t.Fatalf("%s differs (run with -update if intended)", name)
	}
}

func TestRenderResult_golden(t *testing.T) {
	pub, err := RenderResultPage(demoView(t, false))
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "result-public.html", pub)
	priv, err := RenderResultFragment(demoView(t, true))
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "result-private.html", priv)
}

func TestRenderResult_contract(t *testing.T) {
	pub, _ := RenderResultPage(demoView(t, false))
	priv, _ := RenderResultFragment(demoView(t, true))
	p, s := string(pub), string(priv)
	for _, want := range []string{`data-report data-scope="public" data-earned="3"`, `data-share`, `href="/auth/login"`,
		`<details class="badge-card" data-badge="pull-shark">`, `<progress value="56" max="112"`, `Closest next unlock`,
		`https://github.githubassets.com/assets/pull-shark-bronze-`, `<h1 class="result-user">@torvalds</h1>`,
		`3 of 8 badges you can still earn, 12 tiers`, `Showing public data.`, `<h2 class="shelf-head">On display</h2>`,
		`<h2 class="shelf-head">Still to collect</h2>`, `The archive: 5 badges you can no longer earn`,
		`data-star="result"`, `>Copy share link</button>`, `First unlocked 2020-02-02.`} {
		if !strings.Contains(p, want) {
			t.Errorf("public page lacks %s", want)
		}
	}
	if n := strings.Count(p, "<h1"); n != 1 {
		t.Errorf("result page has %d <h1>, want 1", n)
	}
	if strings.Index(p, `data-star="result"`) > strings.Index(p, `data-share`) {
		t.Error("the star banner must come before the share and check-another actions")
	}
	if !strings.Contains(s, `data-scope="private"`) || strings.Contains(s, "data-share") || !strings.Contains(s, "not shareable") {
		t.Error("private fragment must say it is not shareable and offer no share button")
	}
	if strings.Contains(s, "<html") || strings.Contains(s, "<h1") || !strings.Contains(s, `<h2 class="result-user">@torvalds</h2>`) {
		t.Error("the private result is a fragment under the callback page's h1")
	}
}

func TestRenderErrorPage_prefillsAndEscapes(t *testing.T) {
	out, err := RenderErrorPage(404, "not_found", `No user <b>"x"</b> & more`, "torvalds")
	if err != nil {
		t.Fatal(err)
	}
	page := string(out)
	for _, want := range []string{`value="torvalds"`, `<p class="notice error" role="alert">No user &lt;b&gt;&#34;x&#34;&lt;/b&gt; &amp; more</p>`} {
		if !strings.Contains(page, want) {
			t.Errorf("error page lacks %s", want)
		}
	}
	if n := strings.Count(page, "<h1"); n != 1 {
		t.Errorf("error page has %d <h1>, want 1", n)
	}

	hostile, err := RenderErrorPage(400, "bad_user", "Bad username", `"><x`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(hostile), `"><x`) {
		t.Error("a hostile username reached the page unescaped")
	}
	if !strings.Contains(string(hostile), `value="&#34;&gt;&lt;x"`) {
		t.Error("the hostile username is not escaped inside the value attribute")
	}
}
