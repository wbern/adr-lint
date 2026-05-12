// Package deprecatecmd implements the `adr-lint deprecate <id>` subcommand,
// which flips an ADR's frontmatter status to "deprecated".
package deprecatecmd

import (
	"fmt"
	"io"
	"os"

	"github.com/wbern/adr-lint/go/internal/adr"
)

// Run rewrites the ADR identified by args[0] so its frontmatter status is
// "deprecated".
func Run(args []string, dir string, out io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 id: usage: adr-lint deprecate <id>")
	}
	want := adr.NormalizeID(args[0])
	adrs, err := adr.LoadADRs(dir)
	if err != nil {
		return err
	}
	for _, a := range adrs {
		if adr.NormalizeID(a.ID) != want {
			continue
		}
		body, err := os.ReadFile(a.FilePath)
		if err != nil {
			return err
		}
		updated, ok := adr.SetStatus(string(body), "deprecated")
		if !ok {
			return fmt.Errorf("ADR %s has no status line in frontmatter", args[0])
		}
		if err := os.WriteFile(a.FilePath, []byte(updated), 0644); err != nil {
			return err
		}
		fmt.Fprintf(out, "Deprecated %s\n", a.FilePath)
		return nil
	}
	return fmt.Errorf("ADR %s not found", args[0])
}
