package diffchunker

import (
	"strings"
	"testing"
)

func TestChunkDiffByFile_SplitsMultiFileDiff(t *testing.T) {
	diff := `diff --git a/file1.go b/file1.go
index abc123..def456 100644
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
 function foo() {
-  const x = 1;
+  const x = 2;
+  const y = 3;
 }
diff --git a/file2.go b/file2.go
index 111222..333444 100644
--- a/file2.go
+++ b/file2.go
@@ -1,2 +1,3 @@
 function bar() {
+  return true;
 }
`
	chunks := ChunkDiffByFile(diff, 0, false)

	if len(chunks) != 2 {
		t.Fatalf("len(chunks) = %d, want 2", len(chunks))
	}
	if !strings.Contains(chunks[0], "file1.go") || !strings.Contains(chunks[0], "const x = 2") {
		t.Errorf("chunks[0] missing file1 content: %q", chunks[0])
	}
	if strings.Contains(chunks[0], "file2.go") {
		t.Errorf("chunks[0] should not contain file2.go")
	}
	if !strings.Contains(chunks[1], "file2.go") || !strings.Contains(chunks[1], "return true") {
		t.Errorf("chunks[1] missing file2 content: %q", chunks[1])
	}
	if strings.Contains(chunks[1], "file1.go") {
		t.Errorf("chunks[1] should not contain file1.go")
	}
}

func TestChunkDiffByFile_SingleFile(t *testing.T) {
	diff := `diff --git a/file1.go b/file1.go
index abc123..def456 100644
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
 function foo() {
+  return 42;
 }
`
	chunks := ChunkDiffByFile(diff, 0, false)

	if len(chunks) != 1 {
		t.Fatalf("len(chunks) = %d, want 1", len(chunks))
	}
	if !strings.Contains(chunks[0], "file1.go") || !strings.Contains(chunks[0], "return 42") {
		t.Errorf("chunks[0] missing expected content: %q", chunks[0])
	}
}

func TestChunkDiffByFile_GroupsSmallFilesUnderTokenLimit(t *testing.T) {
	diff := `diff --git a/small1.go b/small1.go
index abc..def 100644
--- a/small1.go
+++ b/small1.go
@@ -1 +1,2 @@
+const x = 1;
diff --git a/small2.go b/small2.go
index 111..222 100644
--- a/small2.go
+++ b/small2.go
@@ -1 +1,2 @@
+const y = 2;
diff --git a/small3.go b/small3.go
index 333..444 100644
--- a/small3.go
+++ b/small3.go
@@ -1 +1,2 @@
+const z = 3;
`
	chunks := ChunkDiffByFile(diff, 200, false)

	if len(chunks) >= 3 {
		t.Errorf("expected fewer than 3 chunks, got %d", len(chunks))
	}
	if !strings.Contains(chunks[0], "small1.go") || !strings.Contains(chunks[0], "small2.go") {
		t.Errorf("first chunk should group small1 and small2: %q", chunks[0])
	}
}

func TestChunkDiffByFile_SeparatesChunkWhenFileExceedsLimit(t *testing.T) {
	diff := `diff --git a/large.go b/large.go
index abc..def 100644
--- a/large.go
+++ b/large.go
@@ -1,5 +1,10 @@
+const line1 = "a very long line of code that takes up space";
+const line2 = "another very long line of code that takes up space";
+const line3 = "yet another very long line of code that takes up space";
+const line4 = "more long lines of code that take up space";
+const line5 = "even more long lines of code";
diff --git a/small.go b/small.go
index 111..222 100644
--- a/small.go
+++ b/small.go
@@ -1 +1,2 @@
+const x = 1;
`
	chunks := ChunkDiffByFile(diff, 50, false)

	if len(chunks) != 2 {
		t.Fatalf("len(chunks) = %d, want 2", len(chunks))
	}
	if !strings.Contains(chunks[0], "large.go") || strings.Contains(chunks[0], "small.go") {
		t.Errorf("chunks[0] should contain large.go only: %q", chunks[0])
	}
	if !strings.Contains(chunks[1], "small.go") || strings.Contains(chunks[1], "large.go") {
		t.Errorf("chunks[1] should contain small.go only: %q", chunks[1])
	}
}

func TestChunkDiffByFile_InsertsReminderAboveEachFile(t *testing.T) {
	diff := `diff --git a/file1.go b/file1.go
index abc..def 100644
--- a/file1.go
+++ b/file1.go
@@ -1 +1,2 @@
+const x = 1;
diff --git a/file2.go b/file2.go
index 111..222 100644
--- a/file2.go
+++ b/file2.go
@@ -1 +1,2 @@
+const y = 2;
`
	chunks := ChunkDiffByFile(diff, 1000, true)

	joined := strings.Join(chunks, "\n")
	if !strings.Contains(joined, "⚠️ REMINDER") {
		t.Error("expected reminder marker in chunks")
	}
	if !strings.Contains(joined, "Only check ADDED lines") {
		t.Error("expected reminder body text")
	}
	count := strings.Count(joined, "⚠️ REMINDER")
	if count != 2 {
		t.Errorf("expected 2 reminders (one per file), got %d", count)
	}
}

func TestChunkDiffByFile_EmptyDiff(t *testing.T) {
	chunks := ChunkDiffByFile("", 100000, true)
	if len(chunks) != 0 {
		t.Errorf("expected empty result, got %v", chunks)
	}
}

func TestChunkDiffByFile_ReminderDistinguishesAddedVsHeaders(t *testing.T) {
	diff := `diff --git a/pkg/file.go b/pkg/file.go
index abc..def 100644
--- a/pkg/file.go
+++ b/pkg/file.go
@@ -1 +1,2 @@
+x := 1
`
	chunks := ChunkDiffByFile(diff, 1000, true)

	if len(chunks) != 1 {
		t.Fatalf("len(chunks) = %d, want 1", len(chunks))
	}
	r := chunks[0]
	if !strings.Contains(r, "ADDED code") {
		t.Error("reminder should mention ADDED code")
	}
	if !strings.Contains(r, "file headers — IGNORE") {
		t.Error("reminder should call out file headers as IGNORE")
	}
	if !strings.Contains(r, "existing context — IGNORE") {
		t.Error("reminder should call out existing context as IGNORE")
	}
}
