package responseschema

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestBuildLintResponseSchema_HasRequiredFields(t *testing.T) {
	schema := BuildLintResponseSchema()

	if schema["type"] != "object" {
		t.Errorf("schema.type = %v, want \"object\"", schema["type"])
	}
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatalf("schema.required is not []string: %T", schema["required"])
	}
	for _, want := range []string{"status", "confidence", "explanation"} {
		if !slices.Contains(required, want) {
			t.Errorf("required missing %q (got %v)", want, required)
		}
	}
}

func TestBuildLintResponseSchema_StatusEnumAndConfidenceEnum(t *testing.T) {
	schema := BuildLintResponseSchema()

	// Round-trip through JSON so we exercise the same shape the LLM sees.
	b, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)

	for _, want := range []string{
		`"status"`, `"PASS"`, `"FAIL"`,
		`"confidence"`, `"high"`, `"medium"`, `"low"`,
		`"explanation"`, `"violation"`, `"suggestion"`,
		`"reasoning"`, `"locations"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("marshalled schema missing %s", want)
		}
	}
}
