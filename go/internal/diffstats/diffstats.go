// Package diffstats parses a unified diff into per-file added / removed /
// context line counts.
package diffstats

import (
	"regexp"
	"strings"
)

// FileStats holds the per-file added/removed/context line counts
// extracted from a unified diff.
type FileStats struct {
	Path    string
	Added   int
	Removed int
	Context int
}

// diffHeaderRe captures the source path from a `diff --git a/<path> b/<path>` header.
var diffHeaderRe = regexp.MustCompile(`^diff --git a/(.*?) b/`)

// ParseDiffStats walks a unified diff and returns one FileStats per file.
// Lines preceding the first `diff --git` header are ignored.
func ParseDiffStats(diff string) []FileStats {
	var stats []FileStats
	var (
		currentFile         string
		added, removed, ctx int
		inFile              bool
	)

	flush := func() {
		if !inFile {
			return
		}
		stats = append(stats, FileStats{
			Path: currentFile, Added: added, Removed: removed, Context: ctx,
		})
	}

	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			flush()
			if m := diffHeaderRe.FindStringSubmatch(line); m != nil {
				currentFile = m[1]
				inFile = true
			} else {
				currentFile = ""
				inFile = false
			}
			added, removed, ctx = 0, 0, 0
			continue
		}

		if strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "--- ") ||
			strings.HasPrefix(line, "+++ ") ||
			strings.HasPrefix(line, "@@ ") {
			continue
		}

		if !inFile {
			continue
		}

		switch {
		case strings.HasPrefix(line, "+"):
			added++
		case strings.HasPrefix(line, "-"):
			removed++
		case strings.HasPrefix(line, " "):
			ctx++
		}
	}

	flush()
	return stats
}
