package diffstats

import (
	"strings"
	"testing"
)

// When a diff is supplied from outside git (adr-lint --diff), its own headers
// are the ONLY source of truth for which files it touches. These tests pin the
// two properties everything downstream relies on: the path we resolve, and
// that slicing a diff apart and putting it back together loses nothing.

func TestSplit_EmptyDiff(t *testing.T) {
	if got := Split(""); len(got) != 0 {
		t.Errorf("expected no files, got %v", got)
	}
	if got := Split("   \n\n"); len(got) != 0 {
		t.Errorf("expected no files for whitespace-only diff, got %v", got)
	}
}

func TestSplit_ResolvesPathsPerChangeKind(t *testing.T) {
	tests := []struct {
		name    string
		diff    string
		oldPath string
		newPath string
	}{
		{
			name:    "modify",
			diff:    "diff --git a/pkg/foo.go b/pkg/foo.go\nindex abc..def 100644\n--- a/pkg/foo.go\n+++ b/pkg/foo.go\n@@ -1 +1 @@\n-a\n+b\n",
			oldPath: "pkg/foo.go", newPath: "pkg/foo.go",
		},
		{
			// +++ is /dev/null, so the new path must fall back to the old one
			// or the file drops out of ADR matching entirely.
			name:    "delete",
			diff:    "diff --git a/pkg/gone.go b/pkg/gone.go\ndeleted file mode 100644\nindex abc..000 100644\n--- a/pkg/gone.go\n+++ /dev/null\n@@ -1 +0,0 @@\n-a\n",
			oldPath: "pkg/gone.go", newPath: "pkg/gone.go",
		},
		{
			name:    "add",
			diff:    "diff --git a/pkg/new.go b/pkg/new.go\nnew file mode 100644\nindex 000..abc\n--- /dev/null\n+++ b/pkg/new.go\n@@ -0,0 +1 @@\n+a\n",
			oldPath: "pkg/new.go", newPath: "pkg/new.go",
		},
		{
			// applies_to must match where the code LANDS, not where it left.
			name:    "rename",
			diff:    "diff --git a/old/name.go b/new/name.go\nsimilarity index 95%\nrename from old/name.go\nrename to new/name.go\nindex abc..def 100644\n--- a/old/name.go\n+++ b/new/name.go\n@@ -1 +1 @@\n-a\n+b\n",
			oldPath: "old/name.go", newPath: "new/name.go",
		},
		{
			// The `a/(.*?) b/` header regex is non-greedy and mis-cuts here.
			// The --- / +++ lines are unambiguous, so prefer them.
			name:    "path containing the header separator",
			diff:    "diff --git a/pkg/w b/x.go b/pkg/w b/x.go\nindex abc..def 100644\n--- a/pkg/w b/x.go\n+++ b/pkg/w b/x.go\n@@ -1 +1 @@\n-a\n+b\n",
			oldPath: "pkg/w b/x.go", newPath: "pkg/w b/x.go",
		},
		{
			// git C-quotes paths with spaces or non-ASCII bytes.
			name:    "quoted path",
			diff:    "diff --git \"a/pkg/pa th.go\" \"b/pkg/pa th.go\"\nindex abc..def 100644\n--- \"a/pkg/pa th.go\"\n+++ \"b/pkg/pa th.go\"\n@@ -1 +1 @@\n-a\n+b\n",
			oldPath: "pkg/pa th.go", newPath: "pkg/pa th.go",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Split(tc.diff)
			if len(got) != 1 {
				t.Fatalf("expected 1 file, got %d: %v", len(got), got)
			}
			if got[0].OldPath != tc.oldPath {
				t.Errorf("OldPath = %q, want %q", got[0].OldPath, tc.oldPath)
			}
			if got[0].NewPath != tc.newPath {
				t.Errorf("NewPath = %q, want %q", got[0].NewPath, tc.newPath)
			}
		})
	}
}

func TestSplit_PreservesOrderAndEveryByte(t *testing.T) {
	// THE load-bearing property. Per-ADR slicing hands each ADR a subset of
	// these Texts; if Split drops or rewrites a byte, the reviewer's claim to
	// have shown the model the diff is false, and nothing downstream can catch
	// it. Byte equality is the only honest check.
	diff := "diff --git a/a.go b/a.go\nindex 1..2 100644\n--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@\n-a\n+A\n" +
		"diff --git a/b.go b/b.go\nindex 3..4 100644\n--- a/b.go\n+++ b/b.go\n@@ -1 +1 @@\n-b\n+B\n" +
		"diff --git a/c.go b/c.go\nindex 5..6 100644\n--- a/c.go\n+++ b/c.go\n@@ -1 +1 @@\n-c\n+C\n"

	got := Split(diff)
	if len(got) != 3 {
		t.Fatalf("expected 3 files, got %d", len(got))
	}
	for i, want := range []string{"a.go", "b.go", "c.go"} {
		if got[i].NewPath != want {
			t.Errorf("file %d NewPath = %q, want %q", i, got[i].NewPath, want)
		}
	}
	var rebuilt strings.Builder
	for _, fd := range got {
		rebuilt.WriteString(fd.Text)
	}
	if rebuilt.String() != diff {
		t.Errorf("round-trip lost bytes:\n got %q\nwant %q", rebuilt.String(), diff)
	}
}

func TestSplit_TrailingNewlineAbsent(t *testing.T) {
	// `gh pr diff` does not always end with a newline. Byte equality must hold
	// either way, or the last file silently grows or loses a byte.
	diff := "diff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@\n-a\n+A"
	got := Split(diff)
	if len(got) != 1 {
		t.Fatalf("expected 1 file, got %d", len(got))
	}
	if got[0].Text != diff {
		t.Errorf("Text = %q, want %q", got[0].Text, diff)
	}
}

func TestSplit_IgnoresPreambleBeforeFirstHeader(t *testing.T) {
	diff := "some preamble\nnot a diff\ndiff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@\n-a\n+A\n"
	got := Split(diff)
	if len(got) != 1 {
		t.Fatalf("expected 1 file, got %d", len(got))
	}
	if strings.Contains(got[0].Text, "preamble") {
		t.Errorf("preamble leaked into file text: %q", got[0].Text)
	}
}

func TestSplit_HeaderWithoutResolvablePath(t *testing.T) {
	// A header we cannot resolve must NOT silently become a file with an empty
	// path — the caller needs to be able to refuse.
	diff := "diff --git nonsense\n@@ -1 +1 @@\n-a\n+A\n"
	got := Split(diff)
	if len(got) != 1 {
		t.Fatalf("expected 1 file, got %d", len(got))
	}
	if got[0].NewPath != "" || got[0].OldPath != "" {
		t.Errorf("expected unresolved paths, got old=%q new=%q", got[0].OldPath, got[0].NewPath)
	}
}
