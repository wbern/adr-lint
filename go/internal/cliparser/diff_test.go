package cliparser

import (
	"strings"
	"testing"
)

func TestParseArgs_DiffFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
	}{
		{"space form", []string{"--diff", "pr.diff"}, "pr.diff"},
		{"equals form", []string{"--diff=pr.diff"}, "pr.diff"},
		// A bare "-" means stdin. The shared "next token is a value" guard
		// rejects anything starting with "-", so this needs its own case or
		// piping a diff in silently degrades to a staged-diff run.
		{"stdin", []string{"--diff", "-"}, "-"},
		{"stdin equals form", []string{"--diff=-"}, "-"},
		{"absolute path", []string{"--diff", "/tmp/pr-1335.diff"}, "/tmp/pr-1335.diff"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts, err := ParseArgs(tc.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !opts.DiffSet {
				t.Error("DiffSet = false, want true")
			}
			if opts.DiffPath != tc.path {
				t.Errorf("DiffPath = %q, want %q", opts.DiffPath, tc.path)
			}
		})
	}
}

func TestParseArgs_DiffAbsentLeavesModeUnset(t *testing.T) {
	opts, err := ParseArgs([]string{"--ci"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.DiffSet {
		t.Error("DiffSet = true with no --diff flag")
	}
}

func TestParseArgs_DiffRequiresValue(t *testing.T) {
	for _, args := range [][]string{{"--diff"}, {"--diff", "--ci"}, {"--diff="}} {
		if _, err := ParseArgs(args); err == nil {
			t.Errorf("ParseArgs(%v) = nil error, want a missing-value error", args)
		}
	}
}

func TestParseArgs_DiffConflictsWithOtherDiffSources(t *testing.T) {
	// Both of these supply a diff. Accepting both means silently honouring one
	// and ignoring the other, and the caller cannot tell which.
	conflicts := [][]string{
		{"--diff", "pr.diff", "--branch", "main"},
		{"--branch", "main", "--diff", "pr.diff"},
		{"--diff", "pr.diff", "-b", "main"},
		{"--diff", "pr.diff", "--files", "a.go"},
		{"--files", "a.go", "--diff", "pr.diff"},
	}
	for _, args := range conflicts {
		_, err := ParseArgs(args)
		if err == nil {
			t.Errorf("ParseArgs(%v) = nil error, want a mutual-exclusion error", args)
			continue
		}
		if !strings.Contains(err.Error(), "--diff") {
			t.Errorf("ParseArgs(%v) error %q does not name --diff", args, err)
		}
	}
}

func TestParseArgs_DiffComposesWithBoundedFlags(t *testing.T) {
	// These narrow or shape the run without supplying a diff, so they must stay
	// usable — --adrs in particular is how a shadow comparison re-checks one ADR.
	opts, err := ParseArgs([]string{"--diff", "pr.diff", "--adrs", "21", "--per-file", "--no-cache", "--dry-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.DiffSet || opts.DiffPath != "pr.diff" {
		t.Errorf("diff mode lost: DiffSet=%v DiffPath=%q", opts.DiffSet, opts.DiffPath)
	}
	if len(opts.ADRs) != 1 || opts.ADRs[0] != "21" {
		t.Errorf("ADRs = %v, want [21]", opts.ADRs)
	}
	if !opts.PerFile || !opts.NoCache || !opts.DryRun {
		t.Errorf("bounded flags lost: PerFile=%v NoCache=%v DryRun=%v", opts.PerFile, opts.NoCache, opts.DryRun)
	}
}

func TestParseArgs_CacheAndReportDirs(t *testing.T) {
	// Without these, a run with cwd set to a shared read-only checkout writes
	// .cache/adr-lint and adr-lint-report into it. Neither is gitignored there.
	opts, err := ParseArgs([]string{"--cache-dir", "/tmp/c", "--report-dir=/tmp/r"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.CacheDir != "/tmp/c" {
		t.Errorf("CacheDir = %q, want /tmp/c", opts.CacheDir)
	}
	if opts.ReportDir != "/tmp/r" {
		t.Errorf("ReportDir = %q, want /tmp/r", opts.ReportDir)
	}
	for _, args := range [][]string{{"--cache-dir"}, {"--report-dir"}} {
		if _, err := ParseArgs(args); err == nil {
			t.Errorf("ParseArgs(%v) = nil error, want a missing-value error", args)
		}
	}
}

func TestParseArgs_NoPreFilter(t *testing.T) {
	// The differential switch. Without it, a pre_filter that stopped matching
	// its own rule's vocabulary is undetectable — the skip looks like a pass.
	opts, err := ParseArgs([]string{"--diff", "pr.diff", "--no-pre-filter"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.NoPreFilter {
		t.Error("NoPreFilter = false, want true")
	}
	plain, err := ParseArgs([]string{"--diff", "pr.diff"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plain.NoPreFilter {
		t.Error("NoPreFilter defaulted to true; pre-filtering must stay on by default")
	}
}
