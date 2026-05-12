// Package listcmd implements the `adr-lint list` subcommand, which prints
// a one-line summary of every ADR found under the given directory.
package listcmd

import (
	"fmt"
	"io"
	"strconv"

	"github.com/wbern/adr-lint/go/internal/adr"
)

// Run prints id, status, and title for each ADR under dir.
func Run(dir string, out io.Writer) error {
	adrs, err := adr.LoadADRs(dir)
	if err != nil {
		return err
	}
	if len(adrs) == 0 {
		fmt.Fprintln(out, "No ADRs found")
		return nil
	}
	for _, a := range adrs {
		id := a.ID
		if n, err := strconv.Atoi(id); err == nil {
			id = fmt.Sprintf("%04d", n)
		}
		fmt.Fprintf(out, "%s  %-10s  %s\n", id, a.Status, a.Title)
	}
	return nil
}
