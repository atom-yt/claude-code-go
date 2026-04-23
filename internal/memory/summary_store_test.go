package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create test store in temp directory
func newTestStore(t *testing.T) (*SummaryStore, string) {
	tmpDir := t.TempDir()
	memDir := filepath.Join(tmpDir, "memory")
	require.NoError(t, os.MkdirAll(memDir, 0755))

	storeFile := filepath.Join(memDir, "summaries.json")

	store := &SummaryStore{
		file:     storeFile,
		summaries: make(map[string]*Summary),
		nextID:   1,
	}

	return store, tmpDir
}

// Legacy test kept for backward compatibility reference
// Note: This tests the old file-based design (now replaced by JSON store)
func TestSummaryStoreWriteSessionSummary_Legacy(t *testing.T) {
	// This test is kept for reference but is not executable with new design
	// The new SummaryStore uses JSON persistence, not individual markdown files
	t.Skip("Legacy test - new SummaryStore uses JSON persistence")
}

// New comprehensive tests for the updated SummaryStore

func TestSummaryStore_Save(t *testing.T) {
	store, _ := newTestStore(t)

	summary := &Summary{
		SessionID:   "session-123",
		Model:       "claude-sonnet-4-6",
		Summary:     "Implemented feature X with tests",
		OpenLoops:   []string{"Need to handle edge case Y"},
		PendingTasks: []string{"task-1", "task-2"},
		Tags:        []string{"feature", "testing"},
		Metadata: map[string]string{
			"project": "my-project",
		},
	}

	err := store.Save(summary)
	assert.NoError(t, err)
	assert.NotEmpty(t, summary.ID)
	assert.NotZero(t, summary.CreatedAt)
	assert.NotZero(t, summary.UpdatedAt)
}

func TestSummaryStore_Get(t *testing.T) {
	store, _ := newTestStore(t)

	original := &Summary{
		SessionID: "session-456",
		Model:     "claude-opus-4-6",
		Summary:   "Refactored module A",
	}

	require.NoError(t, store.Save(original))

	retrieved, err := store.Get(original.ID)
	require.NoError(t, err)

	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.SessionID, retrieved.SessionID)
	assert.Equal(t, original.Model, retrieved.Model)
	assert.Equal(t, original.Summary, retrieved.Summary)
}

func TestSummaryStore_Get_NotFound(t *testing.T) {
	store, _ := newTestStore(t)

	_, err := store.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSummaryStore_GetBySessionID(t *testing.T) {
	store, _ := newTestStore(t)

	sessionID := "session-789"

	summary1 := &Summary{
		SessionID: sessionID,
		Model:     "claude-sonnet-4-6",
		Summary:   "Initial work",
	}
	summary2 := &Summary{
		SessionID: sessionID,
		Model:     "claude-sonnet-4-6",
		Summary:   "Follow-up work",
	}

	require.NoError(t, store.Save(summary1))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, store.Save(summary2))

	retrieved, err := store.GetBySessionID(sessionID)
	require.NoError(t, err)

	assert.Equal(t, summary2.Summary, retrieved.Summary)
}

func TestSummaryStore_List(t *testing.T) {
	store, _ := newTestStore(t)

	summaries := []Summary{
		{SessionID: "s1", Model: "m1", Summary: "Summary 1"},
		{SessionID: "s2", Model: "m2", Summary: "Summary 2"},
	}

	for i := range summaries {
		require.NoError(t, store.Save(&summaries[i]))
		time.Sleep(5 * time.Millisecond)
	}

	listed, err := store.List()
	require.NoError(t, err)
	assert.Len(t, listed, 2)
}

func TestSummaryStore_Search(t *testing.T) {
	store, _ := newTestStore(t)

	summaries := []Summary{
		{SessionID: "s1", Model: "m1", Summary: "Fix authentication bug", Tags: []string{"auth", "bug"}},
		{SessionID: "s2", Model: "m2", Summary: "Add user profile feature", Tags: []string{"feature"}},
	}

	for i := range summaries {
		require.NoError(t, store.Save(&summaries[i]))
	}

	results, err := store.Search("authentication")
	require.NoError(t, err)
	assert.Len(t, results, 1)

	// Case insensitive
	results, err = store.Search("AUTHENTICATION")
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSummaryStore_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	memDir := filepath.Join(tmpDir, "memory")
	require.NoError(t, os.MkdirAll(memDir, 0755))

	storeFile := filepath.Join(memDir, "summaries.json")

	// Create and save
	store := &SummaryStore{
		file:     storeFile,
		summaries: make(map[string]*Summary),
		nextID:   1,
	}

	summary := &Summary{
		SessionID: "persist-test",
		Model:     "claude-sonnet-4-6",
		Summary:   "Persistent summary",
	}

	require.NoError(t, store.Save(summary))

	// Create new store instance to load from disk
	store2 := &SummaryStore{
		file:     storeFile,
		summaries: make(map[string]*Summary),
		nextID:   1,
	}

	require.NoError(t, store2.load())

	retrieved, err := store2.Get(summary.ID)
	require.NoError(t, err)
	assert.Equal(t, summary.Summary, retrieved.Summary)
}
