package diffstats

import (
	"reflect"
	"testing"
)

func TestParseDiffStats_EmptyDiff(t *testing.T) {
	if got := ParseDiffStats(""); len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestParseDiffStats_SingleFile(t *testing.T) {
	diff := `diff --git a/pkg/foo.go b/pkg/foo.go
index abc..def 100644
--- a/pkg/foo.go
+++ b/pkg/foo.go
@@ -1,3 +1,4 @@
 func foo() {
-  x := 1
+  x := 2
+  y := 3
 }
`
	got := ParseDiffStats(diff)
	want := []FileStats{
		{Path: "pkg/foo.go", Added: 2, Removed: 1, Context: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestParseDiffStats_MultipleFiles(t *testing.T) {
	diff := `diff --git a/a.go b/a.go
index 111..222 100644
--- a/a.go
+++ b/a.go
@@ -1 +1,2 @@
 existing
+new line
diff --git a/b.go b/b.go
index 333..444 100644
--- a/b.go
+++ b/b.go
@@ -1,2 +1 @@
 keep
-gone
`
	got := ParseDiffStats(diff)
	want := []FileStats{
		{Path: "a.go", Added: 1, Removed: 0, Context: 1},
		{Path: "b.go", Added: 0, Removed: 1, Context: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestParseDiffStats_SkipsMetadataLines(t *testing.T) {
	// `index `, `--- `, `+++ `, `@@ ` must not be counted as add/remove/context
	diff := `diff --git a/x.go b/x.go
index abc..def 100644
--- a/x.go
+++ b/x.go
@@ -1 +1 @@
-old
+new
`
	got := ParseDiffStats(diff)
	want := []FileStats{
		{Path: "x.go", Added: 1, Removed: 1, Context: 0},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestParseDiffStats_HandlesNestedPaths(t *testing.T) {
	diff := `diff --git a/internal/api/server.go b/internal/api/server.go
@@ -1 +1,2 @@
+added
`
	got := ParseDiffStats(diff)
	if len(got) != 1 || got[0].Path != "internal/api/server.go" {
		t.Errorf("unexpected path parse: %+v", got)
	}
}

func TestParseDiffStats_LinesBeforeFirstDiffHeaderAreIgnored(t *testing.T) {
	// `currentFile == nil` guards counts — any leading content before the
	// first `diff --git` should not produce a phantom entry.
	diff := `commit abc123
Author: someone
Date: today

    Some commit message

diff --git a/real.go b/real.go
@@ -1 +1,2 @@
+yes
`
	got := ParseDiffStats(diff)
	if len(got) != 1 {
		t.Fatalf("expected 1 file, got %d (%+v)", len(got), got)
	}
	if got[0].Path != "real.go" || got[0].Added != 1 {
		t.Errorf("unexpected stats: %+v", got)
	}
}
