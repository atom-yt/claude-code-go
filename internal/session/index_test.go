package session

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/atom-yt/claude-code-go/internal/api"
)

func TestIndexBuild(t *testing.T) {
	// Create mock records
	now := time.Now()
	records := []Record{
		{
			ID:        "session-1",
			Model:      "test-model",
			Messages: []api.Message{
				{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: "Hello World Test"}}},
			},
			InputTokens:  50,
			OutputTokens:  30,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:        "session-2",
			Model:      "test-model-2",
			Messages: []api.Message{
				{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: "Another Test"}}},
			},
			InputTokens:  60,
			OutputTokens: 40,
			CreatedAt: now.Add(-time.Hour),
			UpdatedAt: now.Add(-time.Hour),
		},
	}

	idx := NewIndex()
	err := idx.Build(records)
	assert.NoError(t, err)
	assert.Equal(t, 2, idx.Count())

	// Verify entries were built
	entry1 := idx.Get("session-1")
	assert.NotNil(t, entry1)
	assert.Equal(t, "Hello World Test", entry1.Subject)

	entry2 := idx.Get("session-2")
	assert.NotNil(t, entry2)
	assert.Equal(t, "Another Test", entry2.Subject)
}

func TestIndexAddOrUpdate(t *testing.T) {
	idx := NewIndex()

	entry1 := IndexEntry{
		SessionID:    "test-session-1",
		Model:        "test-model",
		Subject:       "test subject",
		TokenCount:   100,
		MessageCount:  5,
	}

	err := idx.AddOrUpdate(&entry1)
	assert.NoError(t, err)
	assert.Equal(t, 1, idx.Count())

	// Update should replace
	entry1.MessageCount = 10
	err = idx.AddOrUpdate(&entry1)
	assert.NoError(t, err)
	assert.Equal(t, 1, idx.Count())
}

func TestIndexSearch(t *testing.T) {
	idx := NewIndex()

	entry1 := IndexEntry{
		SessionID:    "session-1",
		Model:        "model-a",
		Subject:       "Hello World Test",
		TokenCount:   100,
		MessageCount:  5,
	}

	entry2 := IndexEntry{
		SessionID:    "session-2",
		Model:        "model-b",
		Subject:       "Another Test Message",
		TokenCount:   200,
		MessageCount:  10,
	}

	idx.AddOrUpdate(&entry1)
	idx.AddOrUpdate(&entry2)

	// Search for "test"
	results := idx.Search("test")
	assert.Equal(t, 2, len(results))

	// Search for "model"
	results = idx.Search("model")
	assert.Equal(t, 2, len(results))

	// Search for "session" - should match all due to ID prefix
	results = idx.Search("session")
	assert.Equal(t, 2, len(results))

	// Empty search
	results = idx.Search("nonexistent")
	assert.Equal(t, 0, len(results))
}

func TestIndexRemove(t *testing.T) {
	idx := NewIndex()

	entry := IndexEntry{
		SessionID:    "session-1",
		Model:        "test-model",
		Subject:       "test subject",
		TokenCount:   100,
		MessageCount:  5,
	}

	err := idx.AddOrUpdate(&entry)
	assert.NoError(t, err)

	idx.Remove("session-1")
	assert.Equal(t, 0, idx.Count())

	// Removing non-existent should not error
	idx.Remove("nonexistent")
}

func TestIndexGet(t *testing.T) {
	idx := NewIndex()

	entry := IndexEntry{
		SessionID:    "session-1",
		Model:        "test-model",
		Subject:       "test subject",
		TokenCount:   100,
		MessageCount:  5,
	}

	err := idx.AddOrUpdate(&entry)
	assert.NoError(t, err)

	retrieved := idx.Get("session-1")
	assert.NotNil(t, retrieved)
	assert.Equal(t, "session-1", retrieved.SessionID)
	assert.Equal(t, "test subject", retrieved.Subject)

	// Get non-existent
	retrieved = idx.Get("nonexistent")
	assert.Nil(t, retrieved)
}

func TestIndexRecent(t *testing.T) {
	idx := NewIndex()

	now := time.Now()
	for i := 0; i < 5; i++ {
		entry := IndexEntry{
			SessionID:    fmt.Sprintf("session-%d", i),
			UpdatedAt:    now.Add(-time.Duration(i) * time.Hour),
		}
		idx.AddOrUpdate(&entry)
	}

	recent := idx.Recent(3)
	assert.Equal(t, 3, len(recent))

	// Verify order (most recent first)
	assert.True(t, recent[0].UpdatedAt.After(recent[1].UpdatedAt))
	assert.True(t, recent[1].UpdatedAt.After(recent[2].UpdatedAt))

	// Request more than available
	recent = idx.Recent(10)
	assert.Equal(t, 5, len(recent))
}

func TestIndexCount(t *testing.T) {
	idx := NewIndex()

	entry := IndexEntry{
		SessionID:    "test-session-1",
		Model:        "test-model",
		Subject:       "test subject",
		TokenCount:   100,
		MessageCount:  5,
	}

	assert.Equal(t, 0, idx.Count())

	err := idx.AddOrUpdate(&entry)
	assert.NoError(t, err)

	assert.Equal(t, 1, idx.Count())
}

func TestMatchesQuery(t *testing.T) {
	idx := NewIndex()

	// Create test entries as variables
	e1 := IndexEntry{Subject: "Hello World"}
	e2 := IndexEntry{Subject: "HELLO WORLD"}
	e3 := IndexEntry{Model: "gpt-4"}
	e4 := IndexEntry{SessionID: "session-abc-123"}
	e5 := IndexEntry{Subject: "Hello World"}

	tests := []struct {
		name     string
		entry    *IndexEntry
		query       string
		shouldMatch bool
	}{
		{"subject match", &e1, "world", true},
		{"subject match - case insensitive", &e2, "hello", true},
		{"model match", &e3, "gpt", true},
		{"session ID match", &e4, "abc", true},
		{"no match", &e5, "xyz", false},
		{"empty query", &e5, "", true},
	}

	for _, tt := range tests {
		match := idx.matchesQuery(tt.entry, tt.query)
		assert.Equal(t, tt.shouldMatch, match, tt.name)
	}
}