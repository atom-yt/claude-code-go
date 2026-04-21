// Package session provides search and indexing capabilities for sessions.
package session

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/atom-yt/claude-code-go/internal/api"
)

// IndexEntry represents one indexed session.
type IndexEntry struct {
	SessionID    string
	Model        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Subject       string // Extracted from first user message
	TokenCount   int
	MessageCount  int
}

// Index provides search functionality over sessions.
type Index struct {
	mu     sync.RWMutex
	entries map[string]*IndexEntry // keyed by session ID
}

// NewIndex creates a new session index.
func NewIndex() *Index {
	return &Index{
		entries: make(map[string]*IndexEntry),
	}
}

// Build builds the index from provided session records.
func (idx *Index) Build(records []Record) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for _, rec := range records {
		// Extract subject from first user message
		subject := ""
		for _, msg := range rec.Messages {
			if msg.Role == "user" {
				// Extract text from ContentBlock array
				text := extractTextFromContent(msg.Content)
				// Take first ~100 chars as subject
				if len(text) > 100 {
					subject = strings.TrimSpace(text[:100])
				} else {
					subject = strings.TrimSpace(text)
				}
				break
			}
		}

		idx.entries[rec.ID] = &IndexEntry{
			SessionID:    rec.ID,
			Model:        rec.Model,
			CreatedAt:    rec.CreatedAt,
			UpdatedAt:    rec.UpdatedAt,
			Subject:       subject,
			TokenCount:   rec.InputTokens + rec.OutputTokens,
			MessageCount:  len(rec.Messages),
		}
	}

	return nil
}

// extractTextFromContent extracts plain text from ContentBlock array.
func extractTextFromContent(blocks []api.ContentBlock) string {
	var parts []string
	for _, block := range blocks {
		if block.Type == "text" && block.Text != "" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, "")
}

// Search performs full-text search across all sessions.
func (idx *Index) Search(query string) []IndexEntry {
	if query == "" {
		return []IndexEntry{}
	}

	query = strings.ToLower(query)

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var results []IndexEntry
	for _, entry := range idx.entries {
		if idx.matchesQuery(entry, query) {
			results = append(results, *entry)
		}
	}

	return results
}

// matchesQuery checks if an entry matches the search query.
func (idx *Index) matchesQuery(entry *IndexEntry, query string) bool {
	// Check subject
	if strings.Contains(strings.ToLower(entry.Subject), query) {
		return true
	}

	// Check session ID
	if strings.Contains(strings.ToLower(entry.SessionID), query) {
		return true
	}

	// Check model
	if strings.Contains(strings.ToLower(entry.Model), query) {
		return true
	}

	return false
}

// AddOrUpdate adds a new entry or updates an existing one.
func (idx *Index) AddOrUpdate(entry *IndexEntry) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if entry == nil {
		return nil
	}

	// Store a copy of the entry
	entryCopy := *entry
	idx.entries[entry.SessionID] = &entryCopy
	return nil
}

// Remove deletes an entry from the index.
func (idx *Index) Remove(sessionID string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	delete(idx.entries, sessionID)
}

// Get returns an index entry by session ID.
func (idx *Index) Get(sessionID string) *IndexEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.entries[sessionID]
}

// List returns all entries in the index.
func (idx *Index) List() []*IndexEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	entries := make([]*IndexEntry, 0, len(idx.entries))
	for _, entry := range idx.entries {
		entries = append(entries, entry)
	}
	return entries
}

// Count returns the number of entries in the index.
func (idx *Index) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.entries)
}

// Recent returns entries sorted by recency (most recent first).
func (idx *Index) Recent(limit int) []*IndexEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	entries := make([]*IndexEntry, 0, len(idx.entries))
	for _, entry := range idx.entries {
		entries = append(entries, entry)
	}

	// Sort by UpdatedAt descending (most recent first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].UpdatedAt.After(entries[j].UpdatedAt)
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries
}
