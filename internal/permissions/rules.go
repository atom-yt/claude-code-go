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
// Supports ** via filepath.Match after normalisation.
func globMatch(pattern, path string) bool {
	// Fast path: exact match
	if pattern == path {
		return true
	}
	// filepath.Match doesn't support **, so we split on /**/ and try each segment.
	matched, err := filepath.Match(pattern, path)
	if err == nil && matched {
		return true
	}
	// Also try matching just the base name.
	matched, err = filepath.Match(pattern, filepath.Base(path))
	return err == nil && matched
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
