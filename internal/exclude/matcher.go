package exclude

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type rule struct {
	value     string
	isPattern bool
}

type Matcher struct {
	sourceRoot string
	rules      []rule
}

func New(rules []string, sourceRoot string) (*Matcher, error) {
	absSourceRoot, err := filepath.Abs(sourceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	compiled := make([]rule, 0, len(rules))
	for _, raw := range rules {
		normalized := filepath.ToSlash(raw)
		if !doublestar.ValidatePattern(normalized) {
			return nil, fmt.Errorf("%w: exclude rule %q is invalid", ErrInvalidPattern, raw)
		}

		compiled = append(compiled, rule{
			value:     normalized,
			isPattern: hasMeta(normalized),
		})
	}

	return &Matcher{
		sourceRoot: absSourceRoot,
		rules:      compiled,
	}, nil
}

func (m *Matcher) Match(filePath string) bool {
	matched, _ := m.MatchRule(filePath)
	return matched
}

func (m *Matcher) MatchRule(filePath string) (bool, string) {
	candidate := filePath
	if filepath.IsAbs(candidate) {
		if rel, err := filepath.Rel(m.sourceRoot, candidate); err == nil {
			candidate = rel
		}
	}
	normalized := path.Clean(filepath.ToSlash(candidate))

	for _, r := range m.rules {
		if !r.isPattern {
			if normalized == r.value || strings.HasPrefix(normalized, r.value+"/") {
				return true, r.value
			}

			continue
		}

		// Try direct match first
		matched, matchErr := doublestar.Match(r.value, normalized)
		if matchErr == nil && matched {
			return true, r.value
		}

		// If direct match fails, try matching against path prefixes
		// This handles cases like:
		// - pattern "build/*" matching "build/out/main.bin"
		// - pattern "**/node_modules" matching "web/node_modules/react/index.js"
		parts := strings.Split(normalized, "/")
		for i := 1; i <= len(parts); i++ {
			prefix := path.Join(parts[:i]...)
			matched, matchErr = doublestar.Match(r.value, prefix)
			if matchErr == nil && matched {
				return true, r.value
			}
		}
	}

	return false, ""
}

func hasMeta(pattern string) bool {
	return strings.ContainsAny(pattern, "*?{")
}
