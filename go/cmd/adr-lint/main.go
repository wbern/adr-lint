// Command adr-lint is the CLI entrypoint. It wires real os/exec-backed
// clients into runner.Run.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/wbern/adr-lint/go/internal/cache"
	"github.com/wbern/adr-lint/go/internal/claudeclient"
	"github.com/wbern/adr-lint/go/internal/cliparser"
	"github.com/wbern/adr-lint/go/internal/createcmd"
	"github.com/wbern/adr-lint/go/internal/deprecatecmd"
	"github.com/wbern/adr-lint/go/internal/dispatcher"
	"github.com/wbern/adr-lint/go/internal/dotenv"
	"github.com/wbern/adr-lint/go/internal/gitcontext"
	"github.com/wbern/adr-lint/go/internal/listcmd"
	"github.com/wbern/adr-lint/go/internal/runner"
	"github.com/wbern/adr-lint/go/internal/showcmd"
	"github.com/wbern/adr-lint/go/internal/supersedecmd"
	"github.com/wbern/adr-lint/go/internal/types"
)

var subcommands = map[string]dispatcher.Func{
	"create":    createcmd.Run,
	"show":      showcmd.Run,
	"deprecate": deprecatecmd.Run,
	"supersede": supersedecmd.Run,
	"list": func(_ []string, dir string, out io.Writer) error {
		return listcmd.Run(dir, out)
	},
}

func main() {
	git := gitcontext.NewDefaultClient()
	adrDir := filepath.Join(git.GitRoot(), "doc", "adr")
	if err := os.MkdirAll(adrDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	handled, err := dispatcher.Dispatch(os.Args[1:], adrDir, os.Stdout, subcommands)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if handled {
		return
	}

	opts, err := cliparser.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := dotenv.Load(filepath.Join(git.GitRoot(), ".env.local")); err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not load .env.local:", err)
	}
	claude := claudeclient.NewDefaultClient()

	lintFns := map[types.Provider]cache.LintFn{
		types.ProviderClaude: claude.Lint,
	}

	code, err := runner.Run(opts, runner.RunDeps{
		Out:     os.Stdout,
		Err:     os.Stderr,
		Git:     git,
		LintFns: lintFns,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "ADR Lint failed:", err)
		os.Exit(1)
	}
	os.Exit(code)
}
