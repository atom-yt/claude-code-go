package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/atom-yt/claude-code-go/internal/agent"
	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/memory"
)

type mockCompactStreamer struct {
	lastReq api.MessagesRequest
	events  []api.APIEvent
}

func (m *mockCompactStreamer) StreamMessages(_ context.Context, req api.MessagesRequest) <-chan api.APIEvent {
	m.lastReq = req
	ch := make(chan api.APIEvent, len(m.events))
	for _, ev := range m.events {
		ch <- ev
	}
	close(ch)
	return ch
}

func TestCompactHistoryRewritesUIAndAgentHistory(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	repoDir := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	oldWD, _ := os.Getwd()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("chdir repo: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	streamer := &mockCompactStreamer{
		events: []api.APIEvent{
			{Type: api.EventTextDelta, Text: "- preserved goal\n- unresolved issue"},
			{Type: api.EventMessageStop, StopReason: "end_turn"},
		},
	}

	m := newTestModel(80, 24)
	m.compactKeepRecent = 1
	m.sessionID = "session-test-1"
	m.sessionInputTokens = 123
	m.sessionOutputTokens = 45
	m.ag = agent.New(streamer, "claude-test", "anthropic", nil, nil, nil)
	m.ag.SetSystemPrompt("project rules")
	m.ag.SetHistory([]api.Message{
		api.TextMessage(api.RoleUser, "first task"),
		api.TextMessage(api.RoleAssistant, "first answer"),
		api.TextMessage(api.RoleUser, "recent user"),
	})
	m.messages = []ChatMessage{
		{Role: RoleUser, Content: "first task"},
		{Role: RoleAssistant, Content: "first answer"},
		{Role: RoleUser, Content: "recent user"},
	}

	if err := m.compactHistory(context.Background()); err != nil {
		t.Fatalf("compactHistory: %v", err)
	}

	history := m.ag.History()
	if len(history) != 2 {
		t.Fatalf("want 2 history entries after compact, got %d", len(history))
	}
	if got := history[0].Content[0].Text; got != "## Conversation Summary\n\n- preserved goal\n- unresolved issue" {
		t.Fatalf("unexpected summary history: %q", got)
	}
	if got := history[1].Content[0].Text; got != "recent user" {
		t.Fatalf("want recent tail preserved, got %q", got)
	}

	if len(m.messages) != 2 {
		t.Fatalf("want 2 UI messages after compact, got %d", len(m.messages))
	}
	if m.messages[0].Role != RoleAssistant {
		t.Fatalf("want summary UI role assistant, got %q", m.messages[0].Role)
	}
	if m.sessionInputTokens != 0 || m.sessionOutputTokens != 0 {
		t.Fatalf("want session token counters reset, got %d/%d", m.sessionInputTokens, m.sessionOutputTokens)
	}
	if m.lastCompactTime.IsZero() || time.Since(m.lastCompactTime) > time.Minute {
		t.Fatal("want compact timestamp updated")
	}
	if streamer.lastReq.System == nil {
		t.Fatal("want compact request to include system prompt")
	}

	memDir, err := memory.MemoryRootDir()
	if err != nil {
		t.Fatalf("MemoryRootDir: %v", err)
	}
	indexData, err := os.ReadFile(filepath.Join(memDir, "MEMORY.md"))
	if err != nil {
		t.Fatalf("read MEMORY.md: %v", err)
	}
	if got := string(indexData); got == "" || !containsAll(got, "## Session Summaries", "session-test-1.md") {
		t.Fatalf("unexpected MEMORY.md contents:\n%s", got)
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
