package supersedecmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_UnknownReplacementErrors(t *testing.T) {
	dir := t.TempDir()
	body := "---\nstatus: accepted\n---\n# 1. First\n\n## Decision\nx\n"
	path := filepath.Join(dir, "0001-first.md")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var out bytes.Buffer
	err := Run([]string{"1", "999"}, dir, &out)
	if err == nil {
		t.Fatal("expected error for missing replacement")
	}
	if !strings.Contains(err.Error(), "999") || !strings.Contains(err.Error(), "not found") {
		t.Errorf("err = %q", err.Error())
	}
	got, _ := os.ReadFile(path)
	if strings.Contains(string(got), "superseded") {
		t.Errorf("source ADR should be untouched; file is:\n%s", got)
	}
}

func TestRun_SelfSupersessionErrors(t *testing.T) {
	dir := t.TempDir()
	body := "---\nstatus: accepted\n---\n# 1. First\n\n## Decision\nx\n"
	path := filepath.Join(dir, "0001-first.md")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var out bytes.Buffer
	err := Run([]string{"1", "1"}, dir, &out)
	if err == nil {
		t.Fatal("expected error when oldID == newID")
	}
	if !strings.Contains(err.Error(), "itself") {
		t.Errorf("err = %q, want mention of self-supersession", err.Error())
	}
	got, _ := os.ReadFile(path)
	if strings.Contains(string(got), "superseded") {
		t.Errorf("source ADR should be untouched; file is:\n%s", got)
	}
}

func TestRun_NoStatusLineErrors(t *testing.T) {
	dir := t.TempDir()
	oldBody := "---\napplies_to:\n  - \"**/*\"\n---\n# 1. First\n\n## Decision\nx\n"
	newBody := "---\nstatus: accepted\n---\n# 2. Replacement\n\n## Decision\ny\n"
	oldPath := filepath.Join(dir, "0001-first.md")
	if err := os.WriteFile(oldPath, []byte(oldBody), 0644); err != nil {
		t.Fatalf("seed old: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "0002-replacement.md"), []byte(newBody), 0644); err != nil {
		t.Fatalf("seed new: %v", err)
	}

	var out bytes.Buffer
	err := Run([]string{"1", "2"}, dir, &out)
	if err == nil {
		t.Fatal("expected error when source ADR has no status line")
	}
	if !strings.Contains(err.Error(), "status") {
		t.Errorf("err = %q, want mention of status", err.Error())
	}
}

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
