package memory

import (
	"os"
	"path/filepath"
	"strings"
)

// MemoryRootDir returns the path to ~/.claude/projects/<sanitized-git-root>/memory/.
// If no git root is found, returns empty string.
func MemoryRootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	gitRoot, err := findGitRoot()
	if err != nil {
		return "", err
	}

	sanitized := sanitizeGitRoot(gitRoot)
	return filepath.Join(home, ".claude", "projects", sanitized, "memory"), nil
}

// MemoryIndexPath returns the path to MEMORY.md in the memory directory.
func MemoryIndexPath() (string, error) {
	root, err := MemoryRootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "MEMORY.md"), nil
}

// LockFilePath returns the path to the .consolidate-lock file.
func LockFilePath() (string, error) {
	root, err := MemoryRootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".consolidate-lock"), nil
}

// RunningLockFilePath returns the path to the .consolidate-lock.running file.
func RunningLockFilePath() (string, error) {
	root, err := MemoryRootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".consolidate-lock.running"), nil
}

// findGitRoot finds the .git directory by walking up from current directory.
func findGitRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

// sanitizeGitRoot converts a git root path to a safe filename.
// Replaces "/" with "-" and prefixes with "-".
// Example: "/Users/tong/project" -> "-Users-tong-project"
func sanitizeGitRoot(path string) string {
	path = strings.ReplaceAll(path, "/", "-")
	if !strings.HasPrefix(path, "-") {
		path = "-" + path
	}
	return path
}
