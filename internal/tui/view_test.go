package tui

import (
	"strings"
	"testing"

	"github.com/atom-yt/claude-code-go/internal/config"
)

func newTestModel(width, height int) Model {
	m := Model{
		cfg:        config.Settings{Model: "test-model"},
		status:     StatusReady,
		styles:     buildStyles(),
		historyIdx: -1,
		width:      width,
		height:     height,
	}
	return m
}

func TestView_Empty(t *testing.T) {
	m := newTestModel(80, 24)
	out := m.View()
	if out == "" {
		t.Error("View should return non-empty string")
	}
	if !strings.Contains(out, "test-model") {
		t.Error("status bar should show model name")
	}
}

func TestView_UserMessage(t *testing.T) {
	m := newTestModel(80, 24)
	m.messages = []ChatMessage{
		{Role: RoleUser, Content: "hello"},
	}
	out := m.View()
	if !strings.Contains(out, "You:") {
		t.Error("expected 'You:' label")
	}
	if !strings.Contains(out, "hello") {
		t.Error("expected message content")
	}
}

func TestView_ScrollHint(t *testing.T) {
	m := newTestModel(80, 10)
	// Fill with enough messages to enable scrolling.
	for i := 0; i < 20; i++ {
		m.messages = append(m.messages, ChatMessage{Role: RoleUser, Content: "line"})
	}
	m.scrollOffset = 5
	out := m.View()
	if !strings.Contains(out, "scrolled") {
		t.Error("expected scroll hint when scrolled up")
	}
}

func TestView_StatusBar_Thinking(t *testing.T) {
	m := newTestModel(80, 24)
	m.status = StatusThinking
	out := m.View()
	if !strings.Contains(out, "thinking") {
		t.Error("expected 'thinking' in status bar")
	}
}

func TestWordWrap(t *testing.T) {
	lines := wordWrap("hello world foo bar", 10)
	for _, l := range lines {
		if len(l) > 10 {
			t.Errorf("line %q exceeds width 10", l)
		}
	}
}

func TestRenderMessage_ToolProgress(t *testing.T) {
	m := newTestModel(80, 24)
	m.messages = []ChatMessage{
		{Role: RoleToolProgress, Content: "Running Bash(ls)", Streaming: false},
	}
	lines := m.renderMessage(0)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Running Bash") {
		t.Errorf("expected tool name in output: %s", joined)
	}
}

func TestMatchesBracketSequence(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Should match: ANSI CSI sequences
		{"[24;1R", true},
		{"[A", true},
		{"[K", true},
		{"[1;24r", true},
		{"[?1049l", true},
		{"[0m", true},
		{"[38;5;123m", true},

		// Should NOT match: normal input
		{"hello", false},
		{"[hello", false},  // needs to end with letter and only contain valid chars
		{"test", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := matchesBracketSequence(tt.input)
			if result != tt.expected {
				t.Errorf("matchesBracketSequence(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
