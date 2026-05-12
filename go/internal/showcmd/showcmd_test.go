package showcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_UnknownIDErrors(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	err := Run([]string{"42"}, dir, &out)
	if err == nil {
		t.Fatal("expected error for unknown ID")
	}
	if !strings.Contains(err.Error(), "42") || !strings.Contains(err.Error(), "not found") {
		t.Errorf("err = %q", err.Error())
	}
}

func TestRun_PrintsADRBodyByID(t *testing.T) {
	dir := t.TempDir()
	body := "---\nstatus: accepted\n---\n# 1. First Decision\n\n## Context\nbecause\n\n## Decision\ndo the thing\n"
	if err := os.WriteFile(filepath.Join(dir, "0001-first.md"), []byte(body), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var out bytes.Buffer
	if err := Run([]string{"1"}, dir, &out); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out.String(), "do the thing") {
		t.Errorf("output should contain ADR body; got %q", out.String())
	}
}
