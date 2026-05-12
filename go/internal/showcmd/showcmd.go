// Package showcmd implements the `adr-lint show <id>` subcommand, which
// prints the raw contents of one ADR.
package showcmd

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/wbern/adr-lint/go/internal/adr"
)

// Run finds the ADR whose ID matches args[0] and writes its file contents
// to out.
func Run(args []string, dir string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("missing id: usage: adr-lint show <id>")
	}
	want := normalizeID(args[0])
	adrs, err := adr.LoadADRs(dir)
	if err != nil {
		return err
	}
	for _, a := range adrs {
		if normalizeID(a.ID) == want {
			body, err := os.ReadFile(a.FilePath)
			if err != nil {
				return err
			}
			_, err = out.Write(body)
			return err
		}
	}
	return fmt.Errorf("ADR %s not found", args[0])
}

func normalizeID(s string) string {
	if n, err := strconv.Atoi(s); err == nil {
		return fmt.Sprintf("%04d", n)
	}
	return s
}
