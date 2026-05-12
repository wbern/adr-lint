// Package supersedecmd implements the `adr-lint supersede <old> <new>`
// subcommand, which marks an ADR as superseded by another.
package supersedecmd

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/wbern/adr-lint/go/internal/adr"
)

// Run flips the ADR identified by args[0] to status "superseded" and
// records args[1] as its replacement in the frontmatter.
func Run(args []string, dir string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("missing ids: usage: adr-lint supersede <old-id> <new-id>")
	}
	oldID := adr.NormalizeID(args[0])
	newID := adr.NormalizeID(args[1])

	adrs, err := adr.LoadADRs(dir)
	if err != nil {
		return err
	}
	foundReplacement := false
	for _, a := range adrs {
		if adr.NormalizeID(a.ID) == newID {
			foundReplacement = true
			break
		}
	}
	if !foundReplacement {
		return fmt.Errorf("replacement ADR %s not found", args[1])
	}
	for _, a := range adrs {
		if adr.NormalizeID(a.ID) != oldID {
			continue
		}
		body, err := os.ReadFile(a.FilePath)
		if err != nil {
			return err
		}
		updated, ok := adr.SetStatus(string(body), "superseded")
		if !ok {
			return fmt.Errorf("ADR %s has no status line in frontmatter", args[0])
		}
		updated = addOrReplaceSupersededBy(updated, newID)
		if err := os.WriteFile(a.FilePath, []byte(updated), 0644); err != nil {
			return err
		}
		fmt.Fprintf(out, "Superseded %s by %s\n", oldID, newID)
		return nil
	}
	return fmt.Errorf("ADR %s not found", args[0])
}

var statusLineRE = regexp.MustCompile(`(?m)^status:\s*\S+\s*$`)
var supersededByRE = regexp.MustCompile(`(?m)^superseded_by:.*$`)

func addOrReplaceSupersededBy(body, newID string) string {
	line := fmt.Sprintf("superseded_by: %q", newID)
	if supersededByRE.MatchString(body) {
		return supersededByRE.ReplaceAllString(body, line)
	}
	return statusLineRE.ReplaceAllStringFunc(body, func(match string) string {
		return match + "\n" + line
	})
}
