package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSummaryStoreWriteSessionSummary(t *testing.T) {
	root := t.TempDir()
	store := newSummaryStoreWithRoot(root)
	store.nowFn = func() time.Time {
		return time.Date(2026, 4, 20, 13, 0, 0, 0, time.UTC)
	}

	path, err := store.WriteSessionSummary("session-123", "claude-sonnet-4-6", "## Conversation Summary\n\n- goal\n- next step")
	if err != nil {
		t.Fatalf("WriteSessionSummary: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read summary file: %v", err)
	}
	content := string(data)
	for _, want := range []string{
		"# Session Summary",
		"- Session: session-123",
		"- Model: claude-sonnet-4-6",
		"## Conversation Summary",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("summary file missing %q:\n%s", want, content)
		}
	}

	indexPath := filepath.Join(root, "MEMORY.md")
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read MEMORY.md: %v", err)
	}
	index := string(indexData)
	if !strings.Contains(index, "## Session Summaries") {
		t.Fatalf("MEMORY.md missing session summaries section:\n%s", index)
	}
	if !strings.Contains(index, "(sessions/session-123.md)") {
		t.Fatalf("MEMORY.md missing summary link:\n%s", index)
	}
}
