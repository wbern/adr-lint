package dispatcher

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func noopRun(_ []string, _ string, _ io.Writer) error { return nil }

func subs(usage string) map[string]Command {
	return map[string]Command{"create": {Run: noopRun, Usage: usage}}
}

func TestDispatch_UnknownSubcommandErrors(t *testing.T) {
	var out bytes.Buffer

	handled, err := Dispatch([]string{"deprcate", "1"}, "/tmp", &out, subs("adr-lint create <title>"))
	if !handled {
		t.Fatal("expected handled=true so caller doesn't fall through to lint mode")
	}
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "deprcate") {
		t.Errorf("err = %q, want mention of typo", err.Error())
	}
}

func TestDispatch_FlagArgFallsThrough(t *testing.T) {
	var out bytes.Buffer

	handled, err := Dispatch([]string{"--branch"}, "/tmp", &out, subs(""))
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if handled {
		t.Error("flag args should fall through to default lint mode")
	}
}

func TestDispatch_EmptyStringFirstArgFallsThrough(t *testing.T) {
	var out bytes.Buffer

	handled, err := Dispatch([]string{""}, "/tmp", &out, subs(""))
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if handled {
		t.Error("empty-string arg should fall through, not be treated as a subcommand")
	}
}

func TestDispatch_EmptyArgsFallThrough(t *testing.T) {
	var out bytes.Buffer

	handled, err := Dispatch(nil, "/tmp", &out, subs(""))
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if handled {
		t.Error("no args should fall through to default lint mode")
	}
}

func TestDispatch_HelpListsAllSubcommands(t *testing.T) {
	cmds := map[string]Command{
		"create": {Run: noopRun, Usage: "adr-lint create <title>"},
		"list":   {Run: noopRun, Usage: "adr-lint list"},
	}
	var out bytes.Buffer

	handled, err := Dispatch([]string{"help"}, "/tmp", &out, cmds)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !handled {
		t.Fatal("help should be handled")
	}
	for _, want := range []string{"create", "list", "Usage"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("help output missing %q:\n%s", want, out.String())
		}
	}
}

func TestDispatch_SubcommandHelpFlagPrintsUsageAndSkipsRun(t *testing.T) {
	called := false
	cmds := map[string]Command{
		"create": {
			Run:   func(_ []string, _ string, _ io.Writer) error { called = true; return nil },
			Usage: "adr-lint create <title>",
		},
	}
	var out bytes.Buffer

	handled, err := Dispatch([]string{"create", "--help"}, "/tmp", &out, cmds)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !handled {
		t.Fatal("expected handled=true")
	}
	if called {
		t.Error("Run should not be called when --help is passed to a subcommand")
	}
	if !strings.Contains(out.String(), "adr-lint create <title>") {
		t.Errorf("expected usage in output; got:\n%s", out.String())
	}
}

func TestDispatch_SubcommandShortHelpFlagPrintsUsage(t *testing.T) {
	called := false
	cmds := map[string]Command{
		"create": {
			Run:   func(_ []string, _ string, _ io.Writer) error { called = true; return nil },
			Usage: "adr-lint create <title>",
		},
	}
	var out bytes.Buffer

	handled, err := Dispatch([]string{"create", "-h"}, "/tmp", &out, cmds)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !handled {
		t.Fatal("expected handled=true")
	}
	if called {
		t.Error("Run should not be called when -h is passed to a subcommand")
	}
	if !strings.Contains(out.String(), "adr-lint create <title>") {
		t.Errorf("expected usage in output; got:\n%s", out.String())
	}
}
