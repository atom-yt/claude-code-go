package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Summary represents a comprehensive session summary.
type Summary struct {
	ID          string            `json:"id"`
	SessionID   string            `json:"sessionId"`
	Model       string            `json:"model"`
	Summary     string            `json:"summary"`      // AI-generated session summary
	OpenLoops   []string          `json:"openLoops"`   // Unresolved questions/issues
	PendingTasks []string          `json:"pendingTasks"` // Task IDs still pending
	Tags        []string          `json:"tags"`        // Topic tags for discovery
	Metadata    map[string]string `json:"metadata"`    // Custom metadata
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// SummaryStore manages durable storage of session summaries.
type SummaryStore struct {
	mu       sync.RWMutex
	file     string
	summaries map[string]*Summary
	nextID   int
}

// NewSummaryStore creates a new summary store.
// The store will load existing summaries from .claude/summaries.json.
func NewSummaryStore() (*SummaryStore, error) {
	return newSummaryStoreImpl("")
}

// NewSummaryStoreWithRoot creates a new summary store with explicit workspace root.
// The store will load existing summaries from .claude/summaries.json.
// If workspaceRoot is empty, it will be auto-detected.
func NewSummaryStoreWithRoot(workspaceRoot string) (*SummaryStore, error) {
	return newSummaryStoreImpl(workspaceRoot)
}

// newSummaryStoreImpl is the internal implementation.
func newSummaryStoreImpl(workspaceRoot string) (*SummaryStore, error) {
	memDir, err := MemoryRootDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory directory: %w", err)
	}

	file := filepath.Join(memDir, "summaries.json")

	s := &SummaryStore{
		file:     file,
		summaries: make(map[string]*Summary),
		nextID:   1,
	}

	if err := s.load(); err != nil {
		return nil, fmt.Errorf("failed to load summaries: %w", err)
	}

	return s, nil
}

// load loads summaries from storage file.
func (s *SummaryStore) load() error {
	data, err := os.ReadFile(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No existing file
		}
		return err
	}

	var fileData struct {
		Summaries map[string]*Summary `json:"summaries"`
		NextID    int               `json:"nextId"`
	}

	if err := json.Unmarshal(data, &fileData); err != nil {
		return err
	}

	s.summaries = fileData.Summaries
	if fileData.NextID > 0 {
		s.nextID = fileData.NextID
	}

	return nil
}

// save persists summaries to storage file.
func (s *SummaryStore) save() error {
	dir := filepath.Dir(s.file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(struct {
		Summaries map[string]*Summary `json:"summaries"`
		NextID    int               `json:"nextId"`
	}{
		Summaries: s.summaries,
		NextID:    s.nextID,
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.file, data, 0644)
}

// Save creates or updates a summary.
func (s *SummaryStore) Save(summary *Summary) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if summary.ID == "" {
		summary.ID = s.generateID()
	}
	summary.UpdatedAt = time.Now()
	if summary.CreatedAt.IsZero() {
		summary.CreatedAt = time.Now()
	}

	s.summaries[summary.ID] = summary
	return s.save()
}

// Get retrieves a summary by ID.
func (s *SummaryStore) Get(id string) (*Summary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summary, exists := s.summaries[id]
	if !exists {
		return nil, fmt.Errorf("summary not found: %s", id)
	}
	// Return a copy to avoid concurrent modification
	copy := *summary
	return &copy, nil
}

// GetBySessionID retrieves the most recent summary for a session.
func (s *SummaryStore) GetBySessionID(sessionID string) (*Summary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latest *Summary
	for _, summary := range s.summaries {
		if summary.SessionID == sessionID {
			if latest == nil || summary.UpdatedAt.After(latest.UpdatedAt) {
				latest = summary
			}
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no summary found for session: %s", sessionID)
	}
	copy := *latest
	return &copy, nil
}

// List returns all summaries sorted by UpdatedAt descending.
func (s *SummaryStore) List() ([]Summary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]Summary, 0, len(s.summaries))
	for _, summary := range s.summaries {
		summaries = append(summaries, *summary)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].UpdatedAt.After(summaries[j].UpdatedAt)
	})

	return summaries, nil
}

// ListSince returns summaries updated after the given time.
func (s *SummaryStore) ListSince(since time.Time) ([]Summary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var summaries []Summary
	for _, summary := range s.summaries {
		if summary.UpdatedAt.After(since) {
			summaries = append(summaries, *summary)
		}
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].UpdatedAt.After(summaries[j].UpdatedAt)
	})

	return summaries, nil
}

// ListByTag returns summaries with matching tags.
func (s *SummaryStore) ListByTag(tag string) ([]Summary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var summaries []Summary
	for _, summary := range s.summaries {
		for _, t := range summary.Tags {
			if t == tag {
				summaries = append(summaries, *summary)
				break
			}
		}
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].UpdatedAt.After(summaries[j].UpdatedAt)
	})

	return summaries, nil
}

// Delete removes a summary by ID.
func (s *SummaryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.summaries[id]; !exists {
		return fmt.Errorf("summary not found: %s", id)
	}

	delete(s.summaries, id)
	return s.save()
}

// DeleteBySessionID removes all summaries for a session.
func (s *SummaryStore) DeleteBySessionID(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	deleted := false
	for id, summary := range s.summaries {
		if summary.SessionID == sessionID {
			delete(s.summaries, id)
			deleted = true
		}
	}

	if !deleted {
		return fmt.Errorf("no summaries found for session: %s", sessionID)
	}

	return s.save()
}

// UpdatePendingTasks updates the pending task list for a session.
func (s *SummaryStore) UpdatePendingTasks(sessionID string, taskIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var latest *Summary
	for _, summary := range s.summaries {
		if summary.SessionID == sessionID {
			if latest == nil || summary.UpdatedAt.After(latest.UpdatedAt) {
				latest = summary
			}
		}
	}

	if latest == nil {
		return fmt.Errorf("no summary found for session: %s", sessionID)
	}

	latest.PendingTasks = taskIDs
	latest.UpdatedAt = time.Now()
	return s.save()
}

// generateID creates a new unique summary ID.
func (s *SummaryStore) generateID() string {
	id := fmt.Sprintf("summary-%d", s.nextID)
	s.nextID++
	return id
}

// Search performs full-text search across summary content.
func (s *SummaryStore) Search(query string) ([]Summary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Simple case-insensitive substring search
	var results []Summary
	queryLower := lower(query)

	for _, summary := range s.summaries {
		if contains(lower(summary.Summary), queryLower) {
			results = append(results, *summary)
			continue
		}
		for _, tag := range summary.Tags {
			if contains(lower(tag), queryLower) {
				results = append(results, *summary)
				break
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].UpdatedAt.After(results[j].UpdatedAt)
	})

	return results, nil
}

// Helper functions for case-insensitive search
func lower(s string) string {
	// Simple lower - for better Unicode support, use strings.ToLower from stdlib
	low := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			low += string(r + ('a' - 'A'))
		} else {
			low += string(r)
		}
	}
	return low
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// sanitizeSessionFilename converts sessionID to a safe filename.
func sanitizeSessionFilename(sessionID string) string {
	result := make([]byte, len(sessionID))
	for i, c := range sessionID {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' {
			result[i] = byte(c)
		} else {
			result[i] = '-'
		}
	}
	return string(result)
}

// formatSessionSummaryMarkdown formats a session summary as markdown.
func formatSessionSummaryMarkdown(sessionID, model, summary string, now time.Time) string {
	var lines []string
	lines = append(lines, "# Session Summary")
	lines = append(lines, fmt.Sprintf("- Session: %s", sessionID))
	if model != "" {
		lines = append(lines, fmt.Sprintf("- Model: %s", model))
	}
	lines = append(lines, fmt.Sprintf("- Updated: %s", now.Format(time.RFC3339)))
	lines = append(lines, "")
	lines = append(lines, summary)
	return strings.Join(lines, "\n")
}

// updateMemoryIndex adds an entry to MEMORY.md.
func (s *SummaryStore) updateMemoryIndex(relativePath, description string) error {
	memDir, err := MemoryRootDir()
	if err != nil {
		return err
	}
	memIndexPath := filepath.Join(memDir, "MEMORY.md")

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

// WriteSessionSummary is a legacy-compatible method for storing session summaries.
// It creates a Summary struct and saves it to the store.
// For backward compatibility, it also creates a markdown file and updates MEMORY.md.
// Returns the summary ID (previously returned file path for backward compatibility).
func (s *SummaryStore) WriteSessionSummary(sessionID, model, summary string) (string, error) {
	if summary == "" {
		return "", fmt.Errorf("summary is empty")
	}
	if sessionID == "" {
		return "", fmt.Errorf("session ID is required")
	}

	// Save to JSON store
	sum := &Summary{
		SessionID: sessionID,
		Model:     model,
		Summary:   summary,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.Save(sum); err != nil {
		return "", fmt.Errorf("failed to save summary: %w", err)
	}

	// For backward compatibility, also write markdown file and update MEMORY.md
	memDir, err := MemoryRootDir()
	if err != nil {
		// If memory directory not available, just return the ID (JSON storage worked)
		return sum.ID, nil
	}

	// Create markdown file in sessions directory
	sessionsDir := filepath.Join(memDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		return sum.ID, nil
	}

	filename := filepath.Join(sessionsDir, sanitizeSessionFilename(sessionID)+".md")
	content := formatSessionSummaryMarkdown(sessionID, model, summary, time.Now())
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		return sum.ID, nil
	}

	// Update MEMORY.md index
	desc := fmt.Sprintf("Session summary for %s", sessionID)
	if model != "" {
		desc += fmt.Sprintf(" (%s)", model)
	}
	_ = s.updateMemoryIndex(filename, desc)

	return sum.ID, nil
}