// Package pathutil provides path validation and security utilities.
package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Sensitive file/directory names that should be protected.
var sensitiveNames = map[string]bool{
	".env":           true,
	".env.local":     true,
	".env.production": true,
	".env.development": true,
	"credentials":    true,
	"credentials.json": true,
	".credentials":   true,
	".gitconfig":     true,
	".netrc":         true,
	".ssh":           true,
	"id_rsa":         true,
	"id_ed25519":     true,
	"secrets":        true,
	"secrets.json":   true,
	".secrets":       true,
}

// Validator checks paths for security issues.
type Validator struct {
	workspaceRoot string
	allowOutside  bool
}

// NewValidator creates a path validator.
// workspaceRoot is the allowed directory boundary (empty = no boundary check).
func NewValidator(workspaceRoot string) *Validator {
	return &Validator{
		workspaceRoot: workspaceRoot,
		allowOutside:  false,
	}
}

// SetAllowOutside controls whether paths outside workspace are allowed.
func (v *Validator) SetAllowOutside(allow bool) {
	v.allowOutside = allow
}

// Validate checks a path for security issues.
// Returns the normalized absolute path and any validation error.
func (v *Validator) Validate(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}

	// Normalize the path (resolve symlinks where possible)
	normalized := absPath
	if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
		normalized = resolved
	}

	// Check workspace boundary
	if v.workspaceRoot != "" && !v.allowOutside {
		workspaceAbs, err := filepath.Abs(v.workspaceRoot)
		if err != nil {
			return "", fmt.Errorf("cannot resolve workspace root: %w", err)
		}

		workspaceClean := filepath.Clean(workspaceAbs)
		normalizedClean := filepath.Clean(normalized)

		// First try exact match after cleaning
		if normalizedClean == workspaceClean {
			return normalized, nil
		}

		// Then check if it's a subdirectory (prefix match with separator)
		if strings.HasPrefix(normalizedClean+string(filepath.Separator), workspaceClean+string(filepath.Separator)) {
			return normalized, nil
		}

		// Also try with EvalSymlinks for comparison
		workspaceResolved := workspaceClean
		if resolved, err := filepath.EvalSymlinks(workspaceAbs); err == nil {
			workspaceResolved = filepath.Clean(resolved)
		}

		if normalizedClean == workspaceResolved ||
			strings.HasPrefix(normalizedClean+string(filepath.Separator), workspaceResolved+string(filepath.Separator)) {
			return normalized, nil
		}

		return "", fmt.Errorf("path %q is outside workspace %q", path, v.workspaceRoot)
	}

	return normalized, nil
}

// IsSensitive checks if a path points to a sensitive file or contains a sensitive directory.
func IsSensitive(path string) bool {
	// Check if any component is sensitive
	parts := strings.Split(path, string(filepath.Separator))
	for _, part := range parts {
		if sensitiveNames[strings.ToLower(part)] {
			return true
		}
	}
	return false
}

// IsWithinDir checks if a path is within a directory.
func IsWithinDir(path, dir string) bool {
	path = filepath.Clean(path)
	dir = filepath.Clean(dir)

	if !filepath.IsAbs(path) || !filepath.IsAbs(dir) {
		return false
	}

	return strings.HasPrefix(path+string(filepath.Separator), dir+string(filepath.Separator)) ||
		path == dir
}

// ExpandHome expands ~ to the user's home directory.
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}