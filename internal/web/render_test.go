// internal/web/render_test.go
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
		`https://github.githubassets.com/assets/pull-shark-bronze-`} {
		if !strings.Contains(p, want) {
			t.Errorf("public page lacks %s", want)
		}
	}
	if !strings.Contains(s, `data-scope="private"`) || strings.Contains(s, "data-share") || !strings.Contains(s, "not shareable") {
		t.Error("private fragment must say it is not shareable and offer no share button")
	}
	if strings.Contains(s, "<html") {
		t.Error("the private result is a fragment, not a page")
	}
}
