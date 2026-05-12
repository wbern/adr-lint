// Package deprecatecmd implements the `adr-lint deprecate <id>` subcommand,
// which flips an ADR's frontmatter status to "deprecated".
package deprecatecmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"

	"github.com/wbern/adr-lint/go/internal/adr"
)

// Run rewrites the ADR identified by args[0] so its frontmatter status is
// "deprecated".
func Run(args []string, dir string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("missing id: usage: adr-lint deprecate <id>")
	}
	want := normalizeID(args[0])
	adrs, err := adr.LoadADRs(dir)
	if err != nil {
		return err
	}
	for _, a := range adrs {
		if normalizeID(a.ID) != want {
			continue
		}
		body, err := os.ReadFile(a.FilePath)
		if err != nil {
			return err
		}
		updated := setStatus(string(body), "deprecated")
		if err := os.WriteFile(a.FilePath, []byte(updated), 0644); err != nil {
			return err
		}
		fmt.Fprintf(out, "Deprecated %s\n", a.FilePath)
		return nil
	}
	return fmt.Errorf("ADR %s not found", args[0])
}

var statusRE = regexp.MustCompile(`(?m)^status:\s*\S+\s*$`)

func setStatus(body, newStatus string) string {
	return statusRE.ReplaceAllString(body, "status: "+newStatus)
}

func normalizeID(s string) string {
	if n, err := strconv.Atoi(s); err == nil {
		return fmt.Sprintf("%04d", n)
	}
	return s
}
