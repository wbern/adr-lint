package validatecmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeADR(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
		t.Fatalf("seed %s: %v", name, err)
	}
}

func TestRun_FlagsDuplicateIDs(t *testing.T) {
	dir := t.TempDir()
	writeADR(t, dir, "0001-first.md",
		"---\nstatus: accepted\n---\n# 1. First\n\n## Decision\nx\n")
	writeADR(t, dir, "0001-dup.md",
		"---\nstatus: accepted\n---\n# 1. Dup\n\n## Decision\ny\n")

	var out bytes.Buffer
	err := Run(nil, dir, &out)
	if err == nil {
		t.Fatal("expected error for duplicate ID")
	}
	if !strings.Contains(err.Error(), "0001") || !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("err should describe duplicate ID 0001; got %q", err.Error())
	}
}

func TestRun_AcceptsValidSet(t *testing.T) {
	dir := t.TempDir()
	writeADR(t, dir, "0001-old.md",
		"---\nstatus: superseded\nsuperseded_by: \"0002\"\n---\n# 1. Old\n\n## Decision\nx\n")
	writeADR(t, dir, "0002-new.md",
		"---\nstatus: accepted\n---\n# 2. New\n\n## Decision\ny\n")

	var out bytes.Buffer
	if err := Run(nil, dir, &out); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRun_RejectsExtraArgs(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	if err := Run([]string{"extra"}, dir, &out); err == nil {
		t.Fatal("expected error for extra args")
	}
}

func TestRun_FlagsSupersededWithoutTarget(t *testing.T) {
	dir := t.TempDir()
	writeADR(t, dir, "0001-orphan.md",
		"---\nstatus: superseded\n---\n# 1. Orphan\n\n## Decision\nx\n")

	var out bytes.Buffer
	err := Run(nil, dir, &out)
	if err == nil {
		t.Fatal("expected error for status=superseded with no superseded_by")
	}
	if !strings.Contains(err.Error(), "superseded_by") {
		t.Errorf("err should mention missing superseded_by; got %q", err.Error())
	}
}

func TestRun_FlagsDanglingSupersededBy(t *testing.T) {
	dir := t.TempDir()
	writeADR(t, dir, "0001-old.md",
		"---\nstatus: superseded\nsuperseded_by: \"0099\"\n---\n# 1. Old\n\n## Decision\nx\n")

	var out bytes.Buffer
	err := Run(nil, dir, &out)
	if err == nil {
		t.Fatal("expected error for dangling superseded_by")
	}
	if !strings.Contains(err.Error(), "0099") {
		t.Errorf("err should mention dangling target 0099; got %q", err.Error())
	}
}
