package deprecatecmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_MissingIDErrors(t *testing.T) {
	var out bytes.Buffer
	err := Run(nil, t.TempDir(), &out)
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestRun_ExtraArgsErrors(t *testing.T) {
	var out bytes.Buffer
	err := Run([]string{"1", "2"}, t.TempDir(), &out)
	if err == nil {
		t.Fatal("expected error for extra args")
	}
}

func TestRun_UnknownIDErrors(t *testing.T) {
	dir := t.TempDir()
	body := "---\nstatus: accepted\n---\n# 1. First\n\n## Decision\nx\n"
	if err := os.WriteFile(filepath.Join(dir, "0001-first.md"), []byte(body), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var out bytes.Buffer
	err := Run([]string{"99"}, dir, &out)
	if err == nil {
		t.Fatal("expected error for unknown id")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("err = %q, want mention of not found", err.Error())
	}
}

func TestRun_FlipsStatusToDeprecated(t *testing.T) {
	dir := t.TempDir()
	body := "---\nstatus: accepted\n---\n# 1. First\n\n## Decision\nx\n"
	path := filepath.Join(dir, "0001-first.md")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var out bytes.Buffer
	if err := Run([]string{"1"}, dir, &out); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(got), "status: deprecated") {
		t.Errorf("expected status: deprecated; file is:\n%s", got)
	}
	if strings.Contains(string(got), "status: accepted") {
		t.Errorf("old status should be gone; file is:\n%s", got)
	}
}
