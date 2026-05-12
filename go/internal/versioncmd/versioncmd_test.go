package versioncmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun_PrintsVersionLine(t *testing.T) {
	var out bytes.Buffer
	if err := Run(nil, "", &out); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.HasPrefix(out.String(), "adr-lint ") {
		t.Errorf("output should start with %q, got: %q", "adr-lint ", out.String())
	}
	if !strings.HasSuffix(strings.TrimRight(out.String(), "\n"), strings.TrimSpace(strings.TrimPrefix(out.String(), "adr-lint "))) {
		// trivially true — just shape the assertion: there's a non-empty version after the prefix
	}
	if strings.TrimSpace(strings.TrimPrefix(out.String(), "adr-lint ")) == "" {
		t.Errorf("version is empty: %q", out.String())
	}
}

func TestRun_RejectsExtraArgs(t *testing.T) {
	var out bytes.Buffer
	err := Run([]string{"extra"}, "", &out)
	if err == nil {
		t.Fatal("expected error for extra args")
	}
}
