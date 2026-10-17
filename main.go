// gh-achievements: check your GitHub achievements progress from the terminal.
//
//	gh extension install nmloc/gh-achievements
//	gh achievements [username] [--json]
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime/debug"
	"strings"

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
  --json       print the report as JSON
  --no-color   plain text output (also honours NO_COLOR)
  --version    print the version

Docs: https://nmloc.github.io/gh-achievements/
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
	if user != "" && !isLogin(user) {
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
	render.Text(os.Stdout, r, color)
	return nil
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

// reLogin is GitHub's username shape: letters, digits and hyphens, at most 39.
var reLogin = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,38}$`)

func isLogin(s string) bool {
	return reLogin.MatchString(s) && !strings.HasSuffix(s, "-") && !strings.Contains(s, "--")
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
