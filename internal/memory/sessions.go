package memory

import (
	"os"
	"path/filepath"
	"time"

	"github.com/atom-yt/claude-code-go/internal/session"
)

// SessionSummary contains a brief summary of a session for consolidation.
type SessionSummary struct {
	ID        string
	Model     string
	CreatedAt time.Time
	UpdatedAt time.Time
	// We could add more fields like topic, tools used, etc.
}

// GetSessionsSince returns all sessions updated after the given time.
func GetSessionsSince(since time.Time) ([]SessionSummary, error) {
	all, err := session.List()
	if err != nil {
		return nil, err
	}

	var summaries []SessionSummary
	for _, r := range all {
		if r.UpdatedAt.After(since) {
			summaries = append(summaries, SessionSummary{
				ID:        r.ID,
				Model:     r.Model,
				CreatedAt: r.CreatedAt,
				UpdatedAt: r.UpdatedAt,
			})
		}
	}
	return summaries, nil
}

// CountSessionsSince returns the count of sessions updated after the given time.
func CountSessionsSince(since time.Time) (int, error) {
	all, err := session.List()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, r := range all {
		if r.UpdatedAt.After(since) {
			count++
		}
	}
	return count, nil
}

// sessionsDir returns the path to ~/.claude/sessions/.
func sessionsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "sessions"), nil
}