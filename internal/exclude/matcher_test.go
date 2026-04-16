package exclude_test

import (
	"path/filepath"
	"testing"

	"github.com/capcom6/sftp-sync/internal/exclude"
)

func TestNewRejectsInvalidPattern(t *testing.T) {
	t.Parallel()

	_, err := exclude.New([]string{"**/["}, ".")
	if err == nil {
		t.Fatal("expected invalid pattern error")
	}
}

func TestMatcherLiteralPathMatching(t *testing.T) {
	t.Parallel()

	matcher, err := exclude.New([]string{".git", "vendor/cache"}, "/repo")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "exact dir", path: "./.git", want: true},
		{name: "descendant", path: "./.git/config", want: true},
		{name: "other path", path: "./pkg/main.go", want: false},
		{name: "nested descendant", path: "vendor/cache/tmp.txt", want: true},
		{name: "meta-like characters stay literal", path: "vendor/cache[file].txt", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := matcher.Match(tt.path); got != tt.want {
				t.Fatalf("Match(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestMatcherLiteralBackwardCompatibilityWithMetaCharacters(t *testing.T) {
	t.Parallel()

	matcher, err := exclude.New([]string{"cache[file].txt"}, "/repo")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := matcher.Match("cache[file].txt"); !got {
		t.Fatalf("literal path containing [] should still match exactly")
	}

	if got := matcher.Match("cachef.txt"); got {
		t.Fatalf("literal-first behavior expected; glob expansion should not alter matching")
	}
}

func TestMatcherPatternMatching(t *testing.T) {
	t.Parallel()

	matcher, err := exclude.New([]string{"**/*.tmp", "build/*", "**/node_modules"}, "/repo")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "recursive extension", path: "a/b/c.tmp", want: true},
		{name: "single segment", path: "build/main.bin", want: true},
		{name: "single segment nested via matched ancestor", path: "build/out/main.bin", want: true},
		{name: "pattern matches ancestor directory", path: "web/node_modules/react/index.js", want: true},
		{name: "non match", path: "src/main.go", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := matcher.Match(tt.path); got != tt.want {
				t.Fatalf("Match(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestMatcherAbsolutePathAndRootBoundary(t *testing.T) {
	t.Parallel()

	root := filepath.FromSlash("/repo")
	inside := filepath.Join(root, "dist", "bundle.js")
	outside := filepath.FromSlash("/other/dist/bundle.js")

	matcher, err := exclude.New([]string{"dist/**"}, root)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := matcher.Match(inside); !got {
		t.Fatalf("Match(%q) = %v, want true", inside, got)
	}

	if got := matcher.Match(outside); got {
		t.Fatalf("Match(%q) = %v, want false", outside, got)
	}
}
