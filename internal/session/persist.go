// Package session manages conversation persistence.
package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/atom-yt/claude-code-go/internal/api"
)

// Record is one persisted session.
type Record struct {
	ID           string        `json:"id"`
	Model        string        `json:"model"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Messages     []api.Message `json:"messages"`
	InputTokens  int           `json:"input_tokens"`
	OutputTokens int           `json:"output_tokens"`
}

// sessionsDir returns the path to ~/.claude/sessions/.
func sessionsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".claude", "sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// Save writes a session record to disk.
func Save(rec Record) error {
	dir, err := sessionsDir()
	if err != nil {
		return err
	}
	rec.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, rec.ID+".json")
	return os.WriteFile(path, data, 0o644)
}

// Load reads a session by ID.
func Load(id string) (Record, error) {
	dir, err := sessionsDir()
	if err != nil {
		return Record{}, err
	}
	data, err := os.ReadFile(filepath.Join(dir, id+".json"))
	if err != nil {
		return Record{}, err
	}
	var rec Record
	return rec, json.Unmarshal(data, &rec)
}

// Latest returns the most recently updated session, or an error if none exist.
func Latest() (Record, error) {
	all, err := List()
	if err != nil || len(all) == 0 {
		return Record{}, fmt.Errorf("no sessions found")
	}
	return all[0], nil
}

// List returns all sessions sorted by UpdatedAt descending.
func List() ([]Record, error) {
	dir, err := sessionsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var records []Record
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var rec Record
		if err := json.Unmarshal(data, &rec); err != nil {
			continue
		}
		records = append(records, rec)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].UpdatedAt.After(records[j].UpdatedAt)
	})
	return records, nil
}

// NewID generates a session ID from the current time.
func NewID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixMilli())
}
