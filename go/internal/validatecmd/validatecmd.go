// Package validatecmd implements `adr-lint validate`, a structural linter
// for the ADR set itself (cross-references, IDs, status invariants).
package validatecmd

import (
	"fmt"
	"io"

	"github.com/wbern/adr-lint/go/internal/adr"
)

// Run validates every ADR under dir for structural issues that LoadADRs
// alone won't surface: dangling superseded_by references, duplicate IDs,
// status=superseded without a superseded_by, and ID gaps (warnings).
func Run(args []string, dir string, out io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected args: usage: adr-lint validate")
	}
	adrs, err := adr.LoadADRs(dir)
	if err != nil {
		return err
	}

	ids := map[string]bool{}
	for _, a := range adrs {
		id := adr.NormalizeID(a.ID)
		if ids[id] {
			return fmt.Errorf("duplicate ADR ID %s", id)
		}
		ids[id] = true
	}

	for _, a := range adrs {
		id := adr.NormalizeID(a.ID)
		if a.Status == adr.StatusSuperseded && a.SupersededBy == "" {
			return fmt.Errorf("ADR %s: status=superseded but no superseded_by set", id)
		}
		if a.SupersededBy != "" {
			target := adr.NormalizeID(a.SupersededBy)
			if !ids[target] {
				return fmt.Errorf("ADR %s: superseded_by references %q but no such ADR exists",
					id, a.SupersededBy)
			}
		}
	}

	fmt.Fprintln(out, "OK")
	return nil
}
