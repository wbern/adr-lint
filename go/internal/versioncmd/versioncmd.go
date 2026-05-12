// Package versioncmd implements the `adr-lint version` subcommand, which
// prints the binary's module version as embedded by the Go toolchain.
package versioncmd

import (
	"fmt"
	"io"
	"runtime/debug"
)

// Run prints a single "adr-lint <version>" line to out. It takes no args.
func Run(args []string, _ string, out io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected args: usage: adr-lint version")
	}
	fmt.Fprintf(out, "adr-lint %s\n", buildVersion())
	return nil
}

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" {
		return "(unknown)"
	}
	return info.Main.Version
}
