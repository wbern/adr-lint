package acceptcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_FlipsStatusToAccepted(t *testing.T) {
	dir := t.TempDir()
	body := "---\nstatus: proposed\n---\n# 1. First\n\n## Decision\nx\n"
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
	if !strings.Contains(string(got), "status: accepted") {
		t.Errorf("expected status: accepted; file is:\n%s", got)
	}
}

func TestRun_MissingIDErrors(t *testing.T) {
	var out bytes.Buffer
	if err := Run(nil, t.TempDir(), &out); err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestRun_ExtraArgsErrors(t *testing.T) {
	var out bytes.Buffer
	if err := Run([]string{"1", "2"}, t.TempDir(), &out); err == nil {
		t.Fatal("expected error for extra args")
	}
}

func TestRun_UnknownIDErrors(t *testing.T) {
	var out bytes.Buffer
	if err := Run([]string{"99"}, t.TempDir(), &out); err == nil {
		t.Fatal("expected error for unknown id")
	}
}
