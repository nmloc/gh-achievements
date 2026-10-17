// gh-achievements: check your GitHub achievements progress from the terminal.
//
//	gh extension install nmloc/gh-achievements
//	gh achievements [username] [--json] [--all]
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/term"

	"github.com/nmloc/gh-achievements/internal/counters"
	"github.com/nmloc/gh-achievements/internal/official"
	"github.com/nmloc/gh-achievements/internal/render"
	"github.com/nmloc/gh-achievements/internal/report"
)

// version is set by the release build (-ldflags "-X main.version=v1.2.3").
var version = "dev"

const usage = `Check GitHub achievements progress: every badge, your tier and how far to the next one.

Usage:
  gh achievements [username] [flags]

With no username, shows the account you are signed in to with gh.

Flags:
  --all        also list the archive, every highlight and each badge's notes
  --json       print the report as JSON
  --no-color   plain text output (also honours NO_COLOR)
  --version    print the version

Docs: https://gh-achievements.vercel.app/
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gh-achievements:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("gh-achievements", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	all := fs.Bool("all", false, "")
	asJSON := fs.Bool("json", false, "")
	noColor := fs.Bool("no-color", false, "")
	showVersion := fs.Bool("version", false, "")
	if err := fs.Parse(reorder(args)); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *showVersion {
		fmt.Println("gh-achievements", buildVersion())
		return nil
	}
	if fs.NArg() > 1 {
		fs.Usage()
		return errors.New("give at most one username")
	}

	counts := counters.New()
	user := fs.Arg(0)
	if user != "" && !official.IsLogin(user) {
		return fmt.Errorf("%q is not a GitHub username", user)
	}
	if user == "" {
		login, err := currentLogin()
		if err != nil {
			return errors.New("no username given and not signed in to gh: run `gh achievements <username>` or `gh auth login`")
		}
		user = login
	}

	r, err := report.Build(user, official.NewClient(), counts)
	if errors.Is(err, official.ErrUserNotFound) {
		return fmt.Errorf("user %q not found", user)
	}
	if err != nil {
		return err
	}

	if *asJSON {
		return render.JSON(os.Stdout, r)
	}
	t := term.FromEnv()
	color := !*noColor && os.Getenv("NO_COLOR") == "" && t.IsTerminalOutput() && t.IsColorEnabled()
	render.Text(os.Stdout, r, render.Options{Color: color, All: *all})
	if t.IsTerminalOutput() {
		starOnce(os.Stdout, color, starMarker())
	}
	return nil
}

// starOnce prints the request to star the repository the first time a report reaches a
// terminal, and leaves a marker so it never repeats. If the marker cannot be written it
// prints nothing: asking on every run would be worse than not asking.
func starOnce(w io.Writer, color bool, marker string) {
	if marker == "" {
		return
	}
	if _, err := os.Stat(marker); err == nil {
		return // already asked
	}
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		return
	}
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		return
	}
	fmt.Fprintln(w, render.StarLine(color))
	fmt.Fprintln(w)
}

// starMarker is where starOnce remembers that it asked; "" if there is no config directory.
func starMarker() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "gh-achievements", "star-shown")
}

// buildVersion prefers -ldflags, then the VCS tag Go stamps into tagged builds.
func buildVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func currentLogin() (string, error) {
	rest, err := api.DefaultRESTClient()
	if err != nil {
		return "", err
	}
	var me struct{ Login string }
	if err := rest.Get("user", &me); err != nil {
		return "", err
	}
	return me.Login, nil
}

// reorder puts flags before the username, so `gh achievements torvalds --json` works:
// the standard flag package stops at the first positional argument.
func reorder(args []string) []string {
	var flags, rest []string
	for _, a := range args {
		if len(a) > 1 && a[0] == '-' {
			flags = append(flags, a)
		} else {
			rest = append(rest, a)
		}
	}
	return append(flags, rest...)
}
