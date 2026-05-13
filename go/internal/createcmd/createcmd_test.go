package createcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_NoTitleArgsReturnsError(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer

	err := Run(nil, dir, &out)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("err = %q, want mention of title", err.Error())
	}
}

func TestRun_CreatesADRAndPrintsPath(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer

	if err := Run([]string{"Use Testify"}, dir, &out); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := filepath.Join(dir, "0001-use-testify.md")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("expected file at %q: %v", want, err)
	}
	if !strings.Contains(out.String(), want) {
		t.Errorf("output %q missing path %q", out.String(), want)
	}
}

// When the ADR directory is inside cwd, the "Created" message should print
// the relative form (doc/adr/...) rather than leaking the absolute workspace
// path. NOTE: not parallel-safe — mutates os.Chdir.
func TestRun_OutputUsesRelativePathWhenInsideCwd(t *testing.T) {
	tmp := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Errorf("restore cwd: %v", err)
		}
	})
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	adrDir := filepath.Join(tmp, "doc", "adr")
	if err := os.MkdirAll(adrDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	var out bytes.Buffer
	if err := Run([]string{"Use Testify"}, adrDir, &out); err != nil {
		t.Fatalf("Run: %v", err)
	}

	wantRel := filepath.Join("doc", "adr", "0001-use-testify.md")
	if !strings.Contains(out.String(), wantRel) {
		t.Errorf("output %q missing relative path %q", out.String(), wantRel)
	}
	if strings.Contains(out.String(), tmp) {
		t.Errorf("output %q leaked absolute workspace path %q", out.String(), tmp)
	}
}
