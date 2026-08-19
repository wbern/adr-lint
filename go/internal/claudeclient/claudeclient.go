// Package claudeclient invokes the `claude` CLI to lint a diff
// against a single ADR. The shell-out is abstracted behind a Runner
// so tests inject deterministic responses; NewDefaultClient wires the
// production path to os/exec with a 120s timeout.
package claudeclient

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
	"time"

	"github.com/wbern/adr-lint/go/internal/adr"
	"github.com/wbern/adr-lint/go/internal/promptbuilder"
	"github.com/wbern/adr-lint/go/internal/responseparser"
	"github.com/wbern/adr-lint/go/internal/responseschema"
	"github.com/wbern/adr-lint/go/internal/types"
)

// ClaudeTimeout is the wall-clock budget for one `claude` invocation.
const ClaudeTimeout = 120 * time.Second

// Runner sentinel errors. NewDefaultClient's runner wraps real exec
// errors with these so the Lint error-classifier does not need to
// reach into os/exec internals.
var (
	ErrCLINotFound = errors.New("claude CLI not found")
	ErrTimeout     = errors.New("claude CLI timeout")
)

//go:embed complexity-models.json
var complexityModelsJSON []byte

type complexityConfig struct {
	Model             string `json:"model"`
	MaxTokensPerChunk int    `json:"maxTokensPerChunk"`
	MaxOutputTokens   int    `json:"maxOutputTokens"`
}

var claudeModels map[adr.Complexity]complexityConfig

func init() {
	var raw struct {
		Claude map[string]complexityConfig `json:"claude"`
	}
	if err := json.Unmarshal(complexityModelsJSON, &raw); err != nil {
		panic("claudeclient: cannot parse complexity-models.json: " + err.Error())
	}
	claudeModels = make(map[adr.Complexity]complexityConfig, len(raw.Claude))
	for k, v := range raw.Claude {
		claudeModels[adr.Complexity(k)] = v
	}
}

type cliResponse struct {
	Type             string          `json:"type"`
	Result           string          `json:"result"`
	StructuredOutput json.RawMessage `json:"structured_output,omitempty"`
	Usage            *cliUsage       `json:"usage,omitempty"`
}

// cliUsage is the token accounting the Claude CLI already reports and this
// client used to throw away. A pointer, so "the CLI said nothing" stays
// distinguishable from "the CLI said zero" — a zeroed cost reads as a free
// review, which is worse than an absent one.
type cliUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

// toTokenUsage folds the CLI's four counters into the shape the rest of the
// tool reports. Cache reads and cache writes are counted as prompt tokens
// because they ARE input the model processed; omitting them under-reports
// spend precisely on the runs that repeat most.
func (u *cliUsage) toTokenUsage(model string) *types.TokenUsage {
	if u == nil {
		return nil
	}
	prompt := u.InputTokens + u.CacheReadInputTokens + u.CacheCreationInputTokens
	tu := &types.TokenUsage{
		PromptTokens:     prompt,
		CompletionTokens: u.OutputTokens,
		TotalTokens:      prompt + u.OutputTokens,
		Model:            model,
	}
	if u.CacheReadInputTokens > 0 {
		cached := u.CacheReadInputTokens
		tu.CachedTokens = &cached
	}
	return tu
}

// Runner shells out to claude with the given argv and returns its
// stdout. On failure, returns a non-nil error whose Error() string is
// surfaced to the user as the lint explanation.
type Runner func(args []string) (string, error)

// Client carries the claude Runner.
type Client struct {
	run Runner
}

// NewClient builds a Client backed by run.
func NewClient(run Runner) *Client {
	return &Client{run: run}
}

// NewDefaultClient builds a Client that shells out to `claude` via
// os/exec with a 120s timeout. Wraps exec.ErrNotFound as
// ErrCLINotFound and deadline-exceeded as ErrTimeout so the error
// classifier can produce stable user-facing messages.
func NewDefaultClient() *Client {
	return NewClient(func(args []string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), ClaudeTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, "claude", args...)
		out, err := cmd.Output()
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				return "", ErrCLINotFound
			}
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return "", ErrTimeout
			}
			return string(out), err
		}
		return string(out), nil
	})
}

// Lint runs claude over the diff for adr and returns a LintResult.
// Empty diffs short-circuit to a SKIPPED result without calling the
// runner.
func (c *Client) Lint(a adr.ADR, diff string) (types.LintResult, error) {
	if strings.TrimSpace(diff) == "" {
		return types.LintResult{
			ADR:         a,
			Status:      types.StatusSKIPPED,
			Explanation: "No changes to lint",
		}, nil
	}

	prompt := promptbuilder.BuildPrompt(a, diff)
	schema := responseschema.BuildLintResponseSchema()
	model := claudeModels[a.Complexity].Model
	schemaJSON, _ := json.Marshal(schema)

	args := []string{
		"-p", prompt,
		"--output-format", "json",
		"--model", model,
		"--tools", "",
		"--json-schema", string(schemaJSON),
	}

	out, err := c.run(args)
	if err != nil {
		return c.classifyRunnerError(a, err), nil
	}

	var cli cliResponse
	if err := json.Unmarshal([]byte(out), &cli); err != nil {
		return types.LintResult{
			ADR:         a,
			Status:      types.StatusERROR,
			Explanation: "Claude CLI error: " + err.Error(),
		}, nil
	}

	responseText := cli.Result
	if len(cli.StructuredOutput) > 0 {
		responseText = string(cli.StructuredOutput)
	}

	result := responseparser.ParseResponse(a, responseText, cli.Usage.toTokenUsage(model))
	result.PromptBytes = len(prompt)
	return result, nil
}

func (c *Client) classifyRunnerError(a adr.ADR, err error) types.LintResult {
	if errors.Is(err, ErrCLINotFound) {
		return types.LintResult{
			ADR:         a,
			Status:      types.StatusERROR,
			Explanation: "Claude CLI not found. Install with: npm install -g @anthropic-ai/claude-code",
		}
	}
	if errors.Is(err, ErrTimeout) {
		return types.LintResult{
			ADR:         a,
			Status:      types.StatusERROR,
			Explanation: "Claude CLI timeout after 120s",
		}
	}
	return types.LintResult{
		ADR:         a,
		Status:      types.StatusERROR,
		Explanation: "Claude CLI error: " + err.Error(),
	}
}
