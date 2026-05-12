package filefilter

import (
	"reflect"
	"slices"
	"testing"
)

func TestFilterExcludedFiles_PnpmLock(t *testing.T) {
	got := FilterExcludedFiles([]string{"pkg/index.go", "pnpm-lock.yaml"})
	want := []string{"pkg/index.go"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFilterExcludedFiles_PackageLock(t *testing.T) {
	got := FilterExcludedFiles([]string{"pkg/index.go", "package-lock.json"})
	want := []string{"pkg/index.go"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFilterExcludedFiles_YarnLock(t *testing.T) {
	got := FilterExcludedFiles([]string{"pkg/index.go", "yarn.lock"})
	want := []string{"pkg/index.go"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFilterExcludedFiles_AnyLockFile(t *testing.T) {
	got := FilterExcludedFiles([]string{"pkg/index.go", "composer.lock", "Gemfile.lock"})
	want := []string{"pkg/index.go"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFilterExcludedFiles_GeneratedFiles(t *testing.T) {
	got := FilterExcludedFiles([]string{
		"pkg/index.go",
		"pkg/types.generated.go",
		"api.generated.js",
	})
	want := []string{"pkg/index.go"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFilterExcludedFiles_GeneratedDirectories(t *testing.T) {
	got := FilterExcludedFiles([]string{
		"pkg/index.go",
		"internal/api/generated/example/organizations/v1alpha/organizations.pb.go",
		"internal/api/generated/google/protobuf/timestamp.pb.go",
	})
	want := []string{"pkg/index.go"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFilterExcludedFiles_DistDirectory(t *testing.T) {
	got := FilterExcludedFiles([]string{
		"pkg/index.go",
		"dist/bundle.js",
		"dist/types/index.d.go",
	})
	want := []string{"pkg/index.go"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFilterExcludedFiles_NodeModules(t *testing.T) {
	got := FilterExcludedFiles([]string{
		"pkg/index.go",
		"node_modules/lodash/index.js",
	})
	want := []string{"pkg/index.go"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFilterExcludedFiles_KeepsNormalFiles(t *testing.T) {
	input := []string{
		"pkg/index.go",
		"pkg/utils/helpers.go",
		"go.mod",
		"README.md",
	}
	got := FilterExcludedFiles(input)
	if !reflect.DeepEqual(got, input) {
		t.Errorf("got %v, want %v", got, input)
	}
}

func TestFilterExcludedFiles_AllExcludedReturnsEmpty(t *testing.T) {
	got := FilterExcludedFiles([]string{"pnpm-lock.yaml", "package-lock.json"})
	if len(got) != 0 {
		t.Errorf("got %v, want empty slice", got)
	}
}

func TestGetExcludePatterns(t *testing.T) {
	patterns := GetExcludePatterns()
	for _, want := range []string{"pnpm-lock.yaml", "package-lock.json", "*.lock"} {
		if !slices.Contains(patterns, want) {
			t.Errorf("GetExcludePatterns() missing %q (got %v)", want, patterns)
		}
	}
}
