package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const sessionSummariesDirName = "sessions"

// SummaryStore writes compacted session summaries into durable memory storage.
type SummaryStore struct {
	rootDir string
	nowFn   func() time.Time
}

// NewSummaryStore creates a new summary store using the project memory root.
func NewSummaryStore() (*SummaryStore, error) {
	root, err := MemoryRootDir()
	if err != nil {
		return nil, err
	}
	return &SummaryStore{
		rootDir: root,
		nowFn:   time.Now,
	}, nil
}

func newSummaryStoreWithRoot(root string) *SummaryStore {
	return &SummaryStore{
		rootDir: root,
		nowFn:   time.Now,
	}
}

// WriteSessionSummary stores one session summary as a markdown artifact and
// ensures the memory index points to it.
func (s *SummaryStore) WriteSessionSummary(sessionID, model, summary string) (string, error) {
	if strings.TrimSpace(summary) == "" {
		return "", fmt.Errorf("summary is empty")
	}
	if sessionID == "" {
		return "", fmt.Errorf("session ID is required")
	}

	if err := os.MkdirAll(filepath.Join(s.rootDir, sessionSummariesDirName), 0o755); err != nil {
		return "", err
	}

	filename := filepath.Join(sessionSummariesDirName, sanitizeSessionFilename(sessionID)+".md")
	content := formatSessionSummaryMarkdown(sessionID, model, summary, s.nowFn())
	fullPath := filepath.Join(s.rootDir, filename)
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return "", err
	}

	desc := fmt.Sprintf("Session summary for %s", sessionID)
	if model != "" {
		desc += fmt.Sprintf(" (%s)", model)
	}
	if err := s.updateMemoryIndex(filename, desc); err != nil {
		return "", err
	}
	return fullPath, nil
}

func sanitizeSessionFilename(sessionID string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-", ":", "-")
	return replacer.Replace(sessionID)
}

func formatSessionSummaryMarkdown(sessionID, model, summary string, now time.Time) string {
	var parts []string
	parts = append(parts, "# Session Summary")
	parts = append(parts, fmt.Sprintf("- Session: %s", sessionID))
	if model != "" {
		parts = append(parts, fmt.Sprintf("- Model: %s", model))
	}
	parts = append(parts, fmt.Sprintf("- Updated: %s", now.Format(time.RFC3339)))
	parts = append(parts, "")
	parts = append(parts, strings.TrimSpace(summary))
	return strings.Join(parts, "\n")
}

func (s *SummaryStore) updateMemoryIndex(relativePath, description string) error {
	if err := os.MkdirAll(s.rootDir, 0o755); err != nil {
		return err
	}

	memIndexPath := filepath.Join(s.rootDir, "MEMORY.md")
	existingContent, _ := os.ReadFile(memIndexPath)
	existing := string(existingContent)
	if existing == "" {
		existing = "# MEMORY\n\n## Session Summaries\n"
	} else if !strings.Contains(existing, "## Session Summaries") {
		existing = strings.TrimRight(existing, "\n") + "\n\n## Session Summaries\n"
	}

	entry := fmt.Sprintf("- [%s](%s) - %s", filepath.Base(relativePath), relativePath, description)
	if strings.Contains(existing, entry) || strings.Contains(existing, "("+relativePath+")") {
		return os.WriteFile(memIndexPath, []byte(existing), 0o644)
	}

	updated := strings.TrimRight(existing, "\n") + "\n" + entry + "\n"
	return os.WriteFile(memIndexPath, []byte(updated), 0o644)
}
