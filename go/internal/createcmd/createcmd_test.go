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
