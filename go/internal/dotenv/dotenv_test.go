package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func unset(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		prev, had := os.LookupEnv(k)
		os.Unsetenv(k)
		if had {
			t.Cleanup(func() { os.Setenv(k, prev) })
		} else {
			t.Cleanup(func() { os.Unsetenv(k) })
		}
	}
}

func TestLoad_SetsValuesFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env.local")
	if err := os.WriteFile(path, []byte("FOO=bar\nBAZ=qux\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unset(t, "FOO", "BAZ")
	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := os.Getenv("FOO"); got != "bar" {
		t.Errorf("FOO = %q, want bar", got)
	}
	if got := os.Getenv("BAZ"); got != "qux" {
		t.Errorf("BAZ = %q, want qux", got)
	}
}

func TestLoad_MissingFileIsNoError(t *testing.T) {
	if err := Load(filepath.Join(t.TempDir(), "nonexistent")); err != nil {
		t.Errorf("expected nil for missing file, got %v", err)
	}
}

func TestLoad_DoesNotOverwriteExistingEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env.local")
	if err := os.WriteFile(path, []byte("PRESET=fromfile\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PRESET", "fromshell")
	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := os.Getenv("PRESET"); got != "fromshell" {
		t.Errorf("PRESET = %q, want fromshell (file should not override)", got)
	}
}

func TestLoad_IgnoresCommentsAndBlankLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env.local")
	content := "# a comment\n\nA=1\n  # indented comment\nB=2\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	unset(t, "A", "B")
	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if os.Getenv("A") != "1" || os.Getenv("B") != "2" {
		t.Errorf("A=%q B=%q, want 1/2", os.Getenv("A"), os.Getenv("B"))
	}
}

func TestLoad_StripsSurroundingQuotes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env.local")
	if err := os.WriteFile(path, []byte("DQ=\"hello world\"\nSQ='single'\nBARE=plain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unset(t, "DQ", "SQ", "BARE")
	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := os.Getenv("DQ"); got != "hello world" {
		t.Errorf("DQ = %q", got)
	}
	if got := os.Getenv("SQ"); got != "single" {
		t.Errorf("SQ = %q", got)
	}
	if got := os.Getenv("BARE"); got != "plain" {
		t.Errorf("BARE = %q", got)
	}
}

func TestLoad_SkipsMalformedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env.local")
	if err := os.WriteFile(path, []byte("nokey\nKEY=value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unset(t, "KEY")
	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if os.Getenv("KEY") != "value" {
		t.Error("KEY should still load")
	}
}
