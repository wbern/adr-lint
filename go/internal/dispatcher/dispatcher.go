// Package dispatcher routes the first positional argument to a registered
// subcommand. Flag-style args (anything starting with "-") and the empty
// argv pass straight through to the caller's default code path.
package dispatcher

import (
	"fmt"
	"io"
	"sort"
)

// Func matches the signature every management subcommand exposes.
type Func func(args []string, dir string, out io.Writer) error

// Command bundles a subcommand's runner with its one-line usage string.
// The Usage is shown by `adr-lint help` and by `adr-lint <sub> --help`.
type Command struct {
	Run   Func
	Usage string
}

// Dispatch routes args[0] to subs[args[0]]. It returns (true, err) when a
// subcommand was invoked, and (false, nil) when args should fall through
// to the default code path (no args, empty first arg, or first arg starts
// with "-").
//
// The built-in "help" subcommand prints a usage summary listing every
// registered subcommand. Passing `--help` or `-h` to any registered
// subcommand prints that subcommand's Usage instead of running it.
func Dispatch(args []string, dir string, out io.Writer, subs map[string]Command) (bool, error) {
	if len(args) == 0 {
		return false, nil
	}
	first := args[0]
	if first == "" || first[0] == '-' {
		return false, nil
	}
	if first == "help" || first == "--help" || first == "-h" {
		printHelp(out, subs)
		return true, nil
	}
	cmd, ok := subs[first]
	if !ok {
		return true, fmt.Errorf("unknown command: %s\n\nRun `adr-lint help` for usage", first)
	}
	rest := args[1:]
	if hasHelpFlag(rest) {
		fmt.Fprintf(out, "Usage: %s\n", cmd.Usage)
		return true, nil
	}
	return true, cmd.Run(rest, dir, out)
}

func hasHelpFlag(args []string) bool {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return true
		}
	}
	return false
}

func printHelp(out io.Writer, subs map[string]Command) {
	names := make([]string, 0, len(subs))
	for name := range subs {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Fprintln(out, "Usage: adr-lint [subcommand] [args...]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Subcommands:")
	for _, n := range names {
		if u := subs[n].Usage; u != "" {
			fmt.Fprintf(out, "  %s\n", u)
		} else {
			fmt.Fprintf(out, "  %s\n", n)
		}
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Without a subcommand, adr-lint runs the linter against staged changes.")
}
