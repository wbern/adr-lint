// Package adr parses Architecture Decision Records (ADRs) from markdown
// files with optional YAML frontmatter.
package adr

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Status is the lifecycle state of an ADR.
type Status string

const (
	StatusAccepted   Status = "accepted"
	StatusProposed   Status = "proposed"
	StatusDeprecated Status = "deprecated"
	StatusSuperseded Status = "superseded"
)

// Complexity selects the model tier used to lint against this ADR.
type Complexity string

const (
	ComplexityUltralite Complexity = "ultralite"
	ComplexityLite      Complexity = "lite"
	ComplexityStandard  Complexity = "standard"
	ComplexityComplex   Complexity = "complex"
)

// ADR is the parsed representation of a single ADR file. PreFilter is
// always a slice; a single-string YAML value is normalized to a
// one-element slice so the "any pattern matches" semantics hold.
// EnforcedBy is *string to distinguish "missing" from "empty".
type ADR struct {
	ID          string
	Title       string
	Status      Status
	AppliesTo   []string
	Complexity  Complexity
	Decision    string
	FilePath    string
	Content     string
	PreFilter   []string
	EnforcedBy  *string
	DiffContext bool
}

// frontmatter is the raw shape of the YAML block; field types accept the
// union shapes the original parser handled.
type frontmatter struct {
	Status      string      `yaml:"status"`
	Date        string      `yaml:"date"`
	AppliesTo   []string    `yaml:"applies_to"`
	Complexity  string      `yaml:"complexity"`
	PreFilter   interface{} `yaml:"pre_filter"`
	EnforcedBy  string      `yaml:"enforced_by"`
	DiffContext *bool       `yaml:"diff_context"`
}

// titleRe matches the H1 header that carries the ADR's number and title,
// e.g. "# 12. Use Const Assertions".
var titleRe = regexp.MustCompile(`(?m)^# (\d+)\. (.+)$`)

// frontmatterRe matches a leading YAML frontmatter block delimited by "---".
var frontmatterRe = regexp.MustCompile(`(?s)\A---\n(.*?)\n---\n`)

// adrFileRe filters parseADRs to numbered ADR files (0001-foo.md).
var adrFileRe = regexp.MustCompile(`^\d{4}-.*\.md$`)

// ParseADRs reads every numbered ADR file in dir (0001-*.md ... 9999-*.md),
// parses each, and returns those eligible for AI linting:
// non-empty Decision section, status not deprecated/superseded, and no
// `enforced_by` set (those are handled by other tooling).
func ParseADRs(dir string) ([]ADR, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read adr dir %q: %w", dir, err)
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() || !adrFileRe.MatchString(e.Name()) {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
	}
	sort.Strings(paths)

	out := make([]ADR, 0, len(paths))
	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("read adr %q: %w", p, err)
		}
		adr := ParseADR(string(content), p)
		if strings.TrimSpace(adr.Decision) == "" {
			continue
		}
		if adr.Status == StatusDeprecated || adr.Status == StatusSuperseded {
			continue
		}
		if adr.EnforcedBy != nil {
			continue
		}
		out = append(out, adr)
	}
	return out, nil
}

// ParseADR parses a single ADR's raw markdown content. filePath is recorded
// on the ADR but not read; pass the path the content came from.
func ParseADR(content, filePath string) ADR {
	fm, body := extractFrontmatter(content)

	id, title := parseTitle(body, filePath)

	status := parseStatusValue(strOr(fm, func(f *frontmatter) string { return f.Status }))

	var appliesTo []string
	if fm != nil && len(fm.AppliesTo) > 0 {
		appliesTo = fm.AppliesTo
	} else {
		appliesTo = parseAppliesToFromSection(body)
	}

	complexity := parseComplexityValue(strOr(fm, func(f *frontmatter) string { return f.Complexity }))
	if complexity == "" {
		complexity = parseComplexityFromSection(body)
	}

	decision := ExtractSection(body, "Decision")

	preFilter := normalizePreFilter(fm)

	var enforcedBy *string
	if fm != nil && fm.EnforcedBy != "" {
		v := fm.EnforcedBy
		enforcedBy = &v
	}

	diffContext := true
	if fm != nil && fm.DiffContext != nil {
		diffContext = *fm.DiffContext
	}

	return ADR{
		ID:          id,
		Title:       title,
		Status:      status,
		AppliesTo:   appliesTo,
		Complexity:  complexity,
		Decision:    decision,
		FilePath:    filePath,
		Content:     body,
		PreFilter:   preFilter,
		EnforcedBy:  enforcedBy,
		DiffContext: diffContext,
	}
}

func parseTitle(body, filePath string) (id, title string) {
	if m := titleRe.FindStringSubmatch(body); m != nil {
		return m[1], m[2]
	}
	base := filepath.Base(filePath)
	base = strings.TrimSuffix(base, ".md")
	if idx := strings.Index(base, "-"); idx >= 0 {
		return base[:idx], "Unknown"
	}
	return base, "Unknown"
}

func parseStatusValue(v string) Status {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "accepted", "proposed", "deprecated", "superseded":
		return Status(strings.ToLower(strings.TrimSpace(v)))
	}
	return StatusAccepted
}

func parseComplexityValue(v string) Complexity {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "ultralite", "lite", "standard", "complex":
		return Complexity(strings.ToLower(strings.TrimSpace(v)))
	}
	return ""
}

func parseAppliesToFromSection(body string) []string {
	section := ExtractSection(body, "Applies To")
	if section == "" {
		return []string{"**/*"}
	}
	var out []string
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		pat := strings.TrimSpace(trimmed[strings.Index(trimmed, "-")+1:])
		pat = strings.TrimPrefix(pat, "`")
		pat = strings.TrimSuffix(pat, "`")
		if pat != "" {
			out = append(out, pat)
		}
	}
	return out
}

func parseComplexityFromSection(body string) Complexity {
	section := ExtractSection(body, "Complexity")
	if section == "" {
		return ComplexityStandard
	}
	if c := parseComplexityValue(section); c != "" {
		return c
	}
	return ComplexityStandard
}

// ExtractSection returns the trimmed body of the `## <name>` section.
// Lines are collected until the next `## ` header is encountered.
// Returns "" if the section is absent.
func ExtractSection(content, name string) string {
	lines := strings.Split(content, "\n")
	header := "## " + name
	headerIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			headerIdx = i
			break
		}
	}
	if headerIdx == -1 {
		return ""
	}
	var collected []string
	for i := headerIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			break
		}
		collected = append(collected, lines[i])
	}
	return strings.TrimSpace(strings.Join(collected, "\n"))
}

func extractFrontmatter(content string) (*frontmatter, string) {
	m := frontmatterRe.FindStringSubmatchIndex(content)
	if m == nil {
		return nil, content
	}
	yamlBlock := content[m[2]:m[3]]
	body := content[m[1]:]
	var fm frontmatter
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return nil, content
	}
	return &fm, body
}

// strOr returns the result of f(fm) when fm is non-nil, otherwise "".
// Lets parseStatusValue / parseComplexityValue stay nil-safe without
// repeated guards at every call site.
func strOr(fm *frontmatter, f func(*frontmatter) string) string {
	if fm == nil {
		return ""
	}
	return f(fm)
}

// normalizePreFilter folds the `string | []string` YAML union into a slice.
// Returns nil if the field is absent or empty so callers can distinguish
// "unset" from "set to empty list" via len().
func normalizePreFilter(fm *frontmatter) []string {
	if fm == nil || fm.PreFilter == nil {
		return nil
	}
	switch v := fm.PreFilter.(type) {
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}
	return nil
}
