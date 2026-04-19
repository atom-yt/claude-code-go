package permissions

import (
	"path/filepath"
	"strings"
)

// matchRule returns true if the rule matches the given tool call.
// An empty rule field is a wildcard that matches anything.
func matchRule(rule Rule, toolName string, input map[string]any) bool {
	// Tool name match
	if rule.Tool != "" && !strings.EqualFold(rule.Tool, toolName) {
		return false
	}

	// Path match: applies to "file_path" or "path" input fields
	if rule.Path != "" {
		matched := false
		for _, key := range []string{"file_path", "path"} {
			if v, ok := input[key]; ok {
				if s, ok := v.(string); ok && s != "" {
					if globMatch(rule.Path, s) {
						matched = true
						break
					}
				}
			}
		}
		if !matched {
			return false
		}
	}

	// Command match: prefix match on the "command" input field
	if rule.Command != "" {
		cmd, _ := input["command"].(string)
		if !matchCommandPrefix(rule.Command, cmd) {
			return false
		}
	}

	return true
}

// globMatch matches a file path against a glob pattern.
// Supports ** for recursive directory matching.
func globMatch(pattern, path string) bool {
	// Fast path: exact match
	if pattern == path {
		return true
	}

	// Try direct match (Go 1.16+ supports ** in filepath.Match)
	matched, err := filepath.Match(pattern, path)
	if err == nil && matched {
		return true
	}

	// If pattern doesn't contain **, we're done
	if !strings.Contains(pattern, "**") {
		// Fallback: try matching just the base name
		matched, err = filepath.Match(pattern, filepath.Base(path))
		return err == nil && matched
	}

	// Handle ** with custom logic
	return matchDoubleStar(pattern, path)
}

// matchDoubleStar handles ** wildcard patterns for recursive directory matching.
// Pattern like "src/**/*.go" should match "src/a.go", "src/dir/b.go", "src/dir/subdir/c.go"
func matchDoubleStar(pattern, path string) bool {
	// Clean both paths
	pattern = filepath.Clean(pattern)
	path = filepath.Clean(path)

	// Split pattern by **
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		// Multiple **, treat as single *
		pattern = strings.ReplaceAll(pattern, "**", "*")
		matched, err := filepath.Match(pattern, path)
		return err == nil && matched
	}

	prefix := parts[0]
	suffix := parts[1]

	// Handle empty prefix (pattern starts with **)
	if prefix == "" {
		return matchDoubleStarSuffixOnly(suffix, path)
	}

	// Handle empty suffix (pattern ends with **)
	if suffix == "" {
		// prefix/** matches prefix and everything under it
		// Need to handle case where path equals prefix (even without separator)
		// and where path is under prefix
		trimmedPrefix := strings.TrimSuffix(prefix, string(filepath.Separator))
		if path == prefix || path == trimmedPrefix {
			return true
		}
		// Check if path is under prefix (with separator)
		return strings.HasPrefix(path, trimmedPrefix+string(filepath.Separator))
	}

	// Pattern is prefix/**/suffix
	// First, try using Go's Match - it handles some ** cases
	matched, _ := filepath.Match(pattern, path)
	if matched {
		return true
	}

	// Custom logic for full ** support
	// Trim separator from prefix if present
	if len(prefix) > 0 && prefix[len(prefix)-1] == filepath.Separator {
		prefix = strings.TrimSuffix(prefix, string(filepath.Separator))
	}
	// Trim separator from suffix if present
	if len(suffix) > 0 && suffix[0] == filepath.Separator {
		suffix = strings.TrimPrefix(suffix, string(filepath.Separator))
	}

	// Check if path starts with prefix (exactly)
	if !strings.HasPrefix(path, prefix) {
		return false
	}

	// Get the part after prefix
	afterPrefix := path[len(prefix):]

	// If afterPrefix starts with separator, skip it
	if len(afterPrefix) > 0 && afterPrefix[0] == filepath.Separator {
		afterPrefix = afterPrefix[1:]
	}

	// Now check if the remaining part matches suffix
	// The remaining part can have multiple directories, but the last part must match suffix

	// Check if the base name matches suffix (handles src/**/*.go matching src/file.go)
	base := filepath.Base(afterPrefix)
	matched, _ = filepath.Match(suffix, base)
	if matched {
		return true
	}

	// Also check if the whole afterPrefix matches suffix (handles src/**/*.go matching src/sub/file.go)
	matched, _ = filepath.Match(suffix, afterPrefix)
	if matched {
		return true
	}

	// If suffix contains path separators, try to match the end of afterPrefix
	if strings.Contains(suffix, string(filepath.Separator)) {
		// For prefix/**/dir/*.go, we need to find "dir/" in afterPrefix and check if the rest matches "*.go"
		suffixFirstPart := strings.SplitN(suffix, string(filepath.Separator), 2)[0]
		// Find the last occurrence of the first part
		idx := strings.LastIndex(afterPrefix, suffixFirstPart)
		if idx >= 0 {
			remaining := afterPrefix[idx:]
			matched, _ := filepath.Match(suffix, remaining)
			if matched {
				return true
			}
		}
	}

	return false
}

// matchDoubleStarSuffixOnly handles patterns like **/suffix
func matchDoubleStarSuffixOnly(suffix, path string) bool {
	if suffix == "" {
		return true // ** matches everything
	}

	// Trim leading separator from suffix if present
	trimmedSuffix := strings.TrimPrefix(suffix, string(filepath.Separator))

	// For **/*.go:
	// - Check if base name matches *.go
	matched, _ := filepath.Match(trimmedSuffix, filepath.Base(path))
	if matched {
		return true
	}

	// If suffix contains path separators, we need to find it in the path
	if strings.Contains(trimmedSuffix, string(filepath.Separator)) {
		// Find the suffix pattern anywhere in the path
		pathParts := strings.Split(path, string(filepath.Separator))
		suffixParts := strings.Split(trimmedSuffix, string(filepath.Separator))

		// Try to find suffix starting at each position in path
		for i := 0; i <= len(pathParts)-len(suffixParts); i++ {
			match := true
			for j, suffixPart := range suffixParts {
				pathPart := pathParts[i+j]
				matched, err := filepath.Match(suffixPart, pathPart)
				if err != nil || !matched {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}

	return false
}

// matchCommandPrefix returns true if cmd starts with the pattern prefix,
// supporting simple shell-style wildcards via filepath.Match on each word.
func matchCommandPrefix(pattern, cmd string) bool {
	if pattern == "" {
		return true
	}
	if cmd == "" {
		return false
	}
	// Exact prefix match first
	if strings.HasPrefix(cmd, pattern) {
		return true
	}
	// Glob match on the whole command string
	matched, err := filepath.Match(pattern, cmd)
	return err == nil && matched
}
