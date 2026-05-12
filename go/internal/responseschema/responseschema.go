// Package responseschema defines the JSON schema the LLM is asked to
// emit for each ADR/diff lint result. Field order matters: critical
// fields come first so that truncation at the JSON tail loses optional
// fields (reasoning, locations) before required ones.
package responseschema

// Schema is the marshallable shape of the lint-response schema. Using
// map[string]any rather than a struct preserves the load-bearing field
// order and lets callers inject provider-specific keys without
// changing this package.
type Schema map[string]any

// BuildLintResponseSchema returns a fresh copy of the schema each call.
// Callers may mutate the returned map without affecting future callers.
func BuildLintResponseSchema() Schema {
	return Schema{
		"type": "object",
		"properties": map[string]any{
			"status": map[string]any{
				"type":        "string",
				"enum":        []string{"PASS", "FAIL"},
				"description": "Whether the code change passes or fails the lint check",
			},
			"confidence": map[string]any{
				"type":        "string",
				"enum":        []string{"high", "medium", "low"},
				"description": "Confidence level in the result",
			},
			"explanation": map[string]any{
				"type":        "string",
				"description": "PASS: max 80 chars. FAIL: max 200 chars explaining the violation.",
			},
			"violation": map[string]any{
				"type":        "string",
				"description": "Max 100 chars. The specific code that violates. FAIL only, omit for PASS.",
			},
			"suggestion": map[string]any{
				"type":        "string",
				"description": "Max 150 chars. Brief fix recommendation. FAIL only, omit for PASS.",
			},
			"reasoning": map[string]any{
				"type":        "string",
				"description": "Max 300 chars. FAIL only - brief analysis. Omit for PASS.",
			},
			"locations": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "FAIL only. File:line references extracted from diff, e.g. ['pkg/foo.go:42'].",
			},
		},
		"required": []string{"status", "confidence", "explanation"},
	}
}
