// Package filefilter applies the project-wide exclude list to a slice of
// file paths. The pattern set lives in global-excludes.json, embedded so
// the binary is self-contained — no sibling JSON needed at runtime.
package filefilter

import (
	_ "embed"
	"encoding/json"

	"github.com/bmatcuk/doublestar/v4"
)

//go:embed global-excludes.json
var globalExcludesJSON []byte

// excludesFile is the decoded shape of global-excludes.json. The
// "$comment" key in the JSON is intentionally not modelled —
// encoding/json ignores unknown fields by default.
type excludesFile struct {
	Patterns []string `json:"patterns"`
}

var excludePatterns = loadPatterns()

func loadPatterns() []string {
	var f excludesFile
	if err := json.Unmarshal(globalExcludesJSON, &f); err != nil {
		panic("filefilter: embedded global-excludes.json is malformed: " + err.Error())
	}
	return f.Patterns
}

// FilterExcludedFiles returns the subset of files that do not match any
// global-exclude pattern. Pattern matching uses doublestar (minimatch
// semantics) — see pattern-matcher for the same rules applied to ADR globs.
func FilterExcludedFiles(files []string) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		if isExcluded(f) {
			continue
		}
		out = append(out, f)
	}
	return out
}

// GetExcludePatterns returns a copy of the exclude pattern list. Callers
// must not mutate the result.
func GetExcludePatterns() []string {
	cp := make([]string, len(excludePatterns))
	copy(cp, excludePatterns)
	return cp
}

func isExcluded(file string) bool {
	for _, p := range excludePatterns {
		ok, err := doublestar.PathMatch(p, file)
		if err == nil && ok {
			return true
		}
	}
	return false
}
