package syntheticdiff

import (
	"strings"
	"testing"
)

func TestGenerateSyntheticDiff_ValidGitFormat(t *testing.T) {
	content := "line 1\nline 2\nline 3"
	diff := GenerateSyntheticDiff("pkg/example.go", content)

	for _, want := range []string{
		"diff --git a/pkg/example.go b/pkg/example.go",
		"new file mode 100644",
		"--- /dev/null",
		"+++ b/pkg/example.go",
		"+line 1",
		"+line 2",
		"+line 3",
	} {
		if !strings.Contains(diff, want) {
			t.Errorf("diff missing %q\n--- diff ---\n%s", want, diff)
		}
	}
}

func TestGenerateSyntheticDiff_SingleLine(t *testing.T) {
	diff := GenerateSyntheticDiff("README.md", "single line")
	if !strings.Contains(diff, "+single line") {
		t.Errorf("diff missing +single line: %q", diff)
	}
	if !strings.Contains(diff, "@@ -0,0 +1,1 @@") {
		t.Errorf("diff missing hunk header @@ -0,0 +1,1 @@: %q", diff)
	}
}

func TestGenerateSyntheticDiff_EmptyFile(t *testing.T) {
	diff := GenerateSyntheticDiff("empty.txt", "")
	if !strings.Contains(diff, "diff --git a/empty.txt b/empty.txt") {
		t.Errorf("diff missing header for empty.txt: %q", diff)
	}
	// Empty file should produce no "+" content lines (only headers).
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			t.Errorf("empty file produced added line: %q", line)
		}
	}
}

func TestGenerateSyntheticDiff_SpecialCharsInPath(t *testing.T) {
	diff := GenerateSyntheticDiff("pkg/[id]/page.go", "test content")
	want := "diff --git a/pkg/[id]/page.go b/pkg/[id]/page.go"
	if !strings.Contains(diff, want) {
		t.Errorf("diff missing %q\n--- diff ---\n%s", want, diff)
	}
}

func TestGenerateSyntheticDiff_PreservesIndentation(t *testing.T) {
	content := "func test() {\n\tif ok {\n\t\treturn \"indented\"\n\t}\n}"
	diff := GenerateSyntheticDiff("test.go", content)
	for _, want := range []string{
		"+func test() {",
		"+\tif ok {",
		"+\t\treturn \"indented\"",
	} {
		if !strings.Contains(diff, want) {
			t.Errorf("diff missing %q", want)
		}
	}
}
