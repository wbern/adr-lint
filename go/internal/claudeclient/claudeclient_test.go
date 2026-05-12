package claudeclient

import (
	"testing"

	"github.com/wbern/adr-lint/go/internal/adr"
	"github.com/wbern/adr-lint/go/internal/types"
)

func sampleADR() adr.ADR {
	return adr.ADR{
		ID:          "0002",
		Title:       "Use Testify",
		Status:      adr.StatusAccepted,
		AppliesTo:   []string{"**/*_test.go"},
		Complexity:  adr.ComplexityUltralite,
		Decision:    "Check for gomock usage",
		FilePath:    "/test/adr.md",
		Content:     "Test content",
		DiffContext: true,
	}
}

func wrapCLIResponse(structured string) string {
	return `{"type":"result","result":"","structured_output":` + structured + `}`
}

func TestLintWithClaude_DoesNotPopulateTokenUsage(t *testing.T) {
	body := `{"status":"PASS","confidence":"high","explanation":"No violations found"}`
	envelope := `{"type":"result","result":"","structured_output":` + body +
		`,"usage":{"input_tokens":1500,"output_tokens":100}}`
	c := NewClient(func(args []string) (string, error) {
		return envelope, nil
	})

	got, err := c.Lint(sampleADR(), "+ vi.fn()")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if got.TokenUsage != nil {
		t.Errorf("TokenUsage = %+v, want nil", got.TokenUsage)
	}
}

func TestLintWithClaude_InvokesClaudeWithExpectedArgs(t *testing.T) {
	var captured []string
	c := NewClient(func(args []string) (string, error) {
		captured = append([]string(nil), args...)
		return wrapCLIResponse(`{"status":"PASS","confidence":"high","explanation":"OK"}`), nil
	})

	_, err := c.Lint(sampleADR(), "+ vi.fn()")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}

	requiredPairs := [][2]string{
		{"--output-format", "json"},
		{"--tools", ""},
	}
	for _, p := range requiredPairs {
		if !hasAdjacent(captured, p[0], p[1]) {
			t.Errorf("missing adjacent args %q %q in %v", p[0], p[1], captured)
		}
	}
	// `-p` must be followed by a prompt containing the ADR title.
	for i, a := range captured {
		if a == "-p" && i+1 < len(captured) && contains(captured[i+1], "Use Testify") {
			return
		}
	}
	t.Errorf("expected -p followed by prompt containing 'Use Testify', got %v", captured)
}

func hasAdjacent(args []string, a, b string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == a && args[i+1] == b {
			return true
		}
	}
	return false
}

func TestLintWithClaude_TimeoutReturnsERROR(t *testing.T) {
	c := NewClient(func(args []string) (string, error) {
		return "", ErrTimeout
	})

	got, err := c.Lint(sampleADR(), "+ code")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if got.Status != types.StatusERROR {
		t.Errorf("status = %q, want ERROR", got.Status)
	}
	if !contains(got.Explanation, "timeout") {
		t.Errorf("explanation = %q", got.Explanation)
	}
}

func TestLintWithClaude_CLINotFoundReturnsERROR(t *testing.T) {
	c := NewClient(func(args []string) (string, error) {
		return "", ErrCLINotFound
	})

	got, err := c.Lint(sampleADR(), "+ code")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if got.Status != types.StatusERROR {
		t.Errorf("status = %q, want ERROR", got.Status)
	}
	if !contains(got.Explanation, "Claude CLI not found") {
		t.Errorf("explanation = %q", got.Explanation)
	}
}

func TestLintWithClaude_FailResponseWithSuggestion(t *testing.T) {
	body := `{"status":"FAIL","confidence":"high","explanation":"Found gomock usage","violation":"gomock.NewController(t)","suggestion":"Use testify mocks instead"}`
	c := NewClient(func(args []string) (string, error) {
		return wrapCLIResponse(body), nil
	})

	got, err := c.Lint(sampleADR(), "+ gomock.NewController(t)")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if got.Status != types.StatusFAIL {
		t.Errorf("status = %q, want FAIL", got.Status)
	}
	if got.Explanation != "Found gomock usage" {
		t.Errorf("explanation = %q", got.Explanation)
	}
	if got.Suggestion == nil || *got.Suggestion != "Use testify mocks instead" {
		t.Errorf("suggestion = %v, want %q", got.Suggestion, "Use testify mocks instead")
	}
}

func TestLintWithClaude_PassResponse(t *testing.T) {
	body := `{"status":"PASS","confidence":"high","explanation":"No violations found"}`
	c := NewClient(func(args []string) (string, error) {
		return wrapCLIResponse(body), nil
	})

	got, err := c.Lint(sampleADR(), "+ vi.fn()")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if got.Status != types.StatusPASS {
		t.Errorf("status = %q, want PASS", got.Status)
	}
	if got.Explanation != "No violations found" {
		t.Errorf("explanation = %q", got.Explanation)
	}
}

func TestLintWithClaude_EmptyDiffReturnsSKIPPED(t *testing.T) {
	called := false
	c := NewClient(func(args []string) (string, error) {
		called = true
		return "", nil
	})

	got, err := c.Lint(sampleADR(), "")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if got.Status != types.StatusSKIPPED {
		t.Errorf("status = %q, want SKIPPED", got.Status)
	}
	if got.Explanation == "" || !contains(got.Explanation, "No changes") {
		t.Errorf("explanation = %q", got.Explanation)
	}
	if called {
		t.Error("runner should not be called for empty diff")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
