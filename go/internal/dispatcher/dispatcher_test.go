package dispatcher

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func noopRun(_ []string, _ string, _ io.Writer) error { return nil }

func TestDispatch_UnknownSubcommandErrors(t *testing.T) {
	subs := map[string]Func{"create": noopRun}
	var out bytes.Buffer

	handled, err := Dispatch([]string{"deprcate", "1"}, "/tmp", &out, subs)
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
	subs := map[string]Func{"create": noopRun}
	var out bytes.Buffer

	handled, err := Dispatch([]string{"--branch"}, "/tmp", &out, subs)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if handled {
		t.Error("flag args should fall through to default lint mode")
	}
}

func TestDispatch_EmptyArgsFallThrough(t *testing.T) {
	subs := map[string]Func{"create": noopRun}
	var out bytes.Buffer

	handled, err := Dispatch(nil, "/tmp", &out, subs)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if handled {
		t.Error("no args should fall through to default lint mode")
	}
}

func TestDispatch_HelpListsAllSubcommands(t *testing.T) {
	subs := map[string]Func{"create": noopRun, "list": noopRun}
	var out bytes.Buffer

	handled, err := Dispatch([]string{"help"}, "/tmp", &out, subs)
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
