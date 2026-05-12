package supersedecmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_MarksSupersededAndRecordsReplacement(t *testing.T) {
	dir := t.TempDir()
	oldBody := "---\nstatus: accepted\n---\n# 1. First\n\n## Decision\nx\n"
	newBody := "---\nstatus: accepted\n---\n# 2. Replacement\n\n## Decision\ny\n"
	oldPath := filepath.Join(dir, "0001-first.md")
	if err := os.WriteFile(oldPath, []byte(oldBody), 0644); err != nil {
		t.Fatalf("seed old: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "0002-replacement.md"), []byte(newBody), 0644); err != nil {
		t.Fatalf("seed new: %v", err)
	}

	var out bytes.Buffer
	if err := Run([]string{"1", "2"}, dir, &out); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	s := string(got)
	if !strings.Contains(s, "status: superseded") {
		t.Errorf("expected status: superseded; file is:\n%s", s)
	}
	if !strings.Contains(s, "superseded_by: \"0002\"") {
		t.Errorf("expected superseded_by: \"0002\"; file is:\n%s", s)
	}
}
