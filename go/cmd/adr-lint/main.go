// Command adr-lint is the CLI entrypoint. It wires real os/exec-backed
// clients into runner.Run.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wbern/adr-lint/go/internal/cache"
	"github.com/wbern/adr-lint/go/internal/claudeclient"
	"github.com/wbern/adr-lint/go/internal/cliparser"
	"github.com/wbern/adr-lint/go/internal/dotenv"
	"github.com/wbern/adr-lint/go/internal/gitcontext"
	"github.com/wbern/adr-lint/go/internal/runner"
	"github.com/wbern/adr-lint/go/internal/types"
)

func main() {
	opts, err := cliparser.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	git := gitcontext.NewDefaultClient()
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
