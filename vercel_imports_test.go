package main

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Vercel's Go runtime compiles every .go file under api/ as its own program, under a
// synthetic import path (handler/api/<name>). Go's internal-package rule is evaluated
// against that path, not against this module's path, so a function importing
// github.com/nmloc/gh-achievements/internal/... builds locally and fails on Vercel with
// "use of internal package ... not allowed". Shared code lives in public packages.
func TestAPIFunctions_doNotImportInternalPackages(t *testing.T) {
	checked := 0
	err := filepath.WalkDir("api", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		checked++
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if strings.Contains(p+"/", "/internal/") {
				t.Errorf("%s imports %q: Vercel builds api/ files under the synthetic import path "+
					"handler/api/<name>, so Go rejects any internal package there; move the code "+
					"to a public package", path, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatal("found no .go files under api/; the guard would pass vacuously")
	}
}
