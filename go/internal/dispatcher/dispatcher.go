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

// Dispatch routes args[0] to subs[args[0]]. It returns (true, err) when a
// subcommand was invoked, and (false, nil) when args should fall through
// to the default code path (no args, or first arg starts with "-").
//
// The built-in "help" subcommand prints a usage summary listing every
// registered subcommand.
func Dispatch(args []string, dir string, out io.Writer, subs map[string]Func) (bool, error) {
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
	fn, ok := subs[first]
	if !ok {
		return true, fmt.Errorf("unknown command: %s\n\nRun `adr-lint help` for usage", first)
	}
	return true, fn(args[1:], dir, out)
}

func printHelp(out io.Writer, subs map[string]Func) {
	names := make([]string, 0, len(subs))
	for name := range subs {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Fprintln(out, "Usage: adr-lint [subcommand] [args...]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Subcommands:")
	for _, n := range names {
		fmt.Fprintf(out, "  %s\n", n)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Without a subcommand, adr-lint runs the linter against staged changes.")
}
