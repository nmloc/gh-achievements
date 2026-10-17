package official

import (
	"bytes"
	"compress/gzip"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Fixtures are real github.com pages, gzipped. Refresh them with:
//
//	go test ./internal/official -run TestUpdateFixtures -update
var update = flag.Bool("update", false, "re-download fixtures from github.com")

// fixtureUsers cover every parser branch: tiered and single badges, retired badges,
// repository milestones, private milestones, highlights, a hidden profile.
var fixtureUsers = []string{
	"torvalds", "ljharb", "Schweinepriester", "brannon", "dylanatsmith",
	"kentcdodds", "octocat", "sindresorhus", "timrogers",
}

func fixturePath(user, name string) string {
	return filepath.Join("testdata", user, name+".html.gz")
}

func loadFixture(t *testing.T, user, name string) string {
	t.Helper()
	f, err := os.Open(fixturePath(user, name))
	if err != nil {
		t.Fatalf("fixture %s/%s: %v (run with -update)", user, name, err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func saveFixture(t *testing.T, user, name, body string) {
	t.Helper()
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if _, err := zw.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(fixturePath(user, name)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixturePath(user, name), buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateFixtures(t *testing.T) {
	if !*update {
		t.Skip("run with -update to refresh fixtures")
	}
	c := NewClient()
	for _, user := range fixtureUsers {
		tab, found, err := c.get("/"+user+"?tab=achievements", false)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			main, found, err := c.get("/"+user, false)
			if err != nil || !found {
				t.Fatalf("%s: profile missing: %v", user, err)
			}
			saveFixture(t, user, "main", main)
			continue
		}
		saveFixture(t, user, "tab", tab)
		for _, slug := range ParseEarned(tab) {
			frag, found, err := c.get("/users/"+user+"/achievements/"+slug+"/detail", true)
			if err != nil || !found {
				t.Fatalf("%s/%s: detail missing: %v", user, slug, err)
			}
			saveFixture(t, user, "detail-"+slug, frag)
		}
	}
}
