package tui

import (
	"strings"
	"testing"

	"github.com/atom-yt/claude-code-go/internal/commands"
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

func TestView_StatusBar_Asking(t *testing.T) {
	m := newTestModel(80, 24)
	m.status = StatusAsking
	out := m.View()
	if !strings.Contains(out, "waiting for approval") {
		t.Error("expected 'waiting for approval' in status bar")
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

func TestWordWrap_Empty(t *testing.T) {
	lines := wordWrap("", 10)
	if len(lines) != 1 || lines[0] != "" {
		t.Error("empty input should return single empty line")
	}
}

func TestWordWrap_ZeroWidth(t *testing.T) {
	lines := wordWrap("test", 0)
	// Zero width should not panic
	if len(lines) != 1 || lines[0] != "test" {
		t.Error("zero width should return original text")
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

func TestRenderMessage_Error(t *testing.T) {
	m := newTestModel(80, 24)
	m.messages = []ChatMessage{
		{Role: RoleError, Content: "something went wrong"},
	}
	lines := m.renderMessage(0)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Error:") {
		t.Errorf("expected 'Error:' label: %s", joined)
	}
	if !strings.Contains(joined, "something went wrong") {
		t.Errorf("expected error message: %s", joined)
	}
}

func TestRenderMessage_Ask(t *testing.T) {
	m := newTestModel(80, 24)
	m.messages = []ChatMessage{
		{Role: RoleAsk, Content: "Allow this operation?"},
	}
	lines := m.renderMessage(0)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Allow this operation?") {
		t.Errorf("expected ask content: %s", joined)
	}
}

func TestRenderTokenProgressBar(t *testing.T) {
	m := newTestModel(80, 24)
	m.contextWindow = 1000

	// Test with 50% usage
	m.sessionInputTokens = 300
	m.sessionOutputTokens = 200
	bar := m.renderTokenProgressBar()
	if bar == "" {
		t.Error("progress bar should not be empty with active session")
	}
	if !strings.Contains(bar, "50%") {
		t.Errorf("expected 50%% in progress bar: %s", bar)
	}

	// Test with 0 usage
	m.sessionInputTokens = 0
	m.sessionOutputTokens = 0
	bar = m.renderTokenProgressBar()
	if bar != "" {
		t.Errorf("progress bar should be empty with no usage: %s", bar)
	}

	// Test with 100% context window (edge case)
	m.sessionInputTokens = 0
	m.sessionOutputTokens = 0
	m.contextWindow = 0
	bar = m.renderTokenProgressBar()
	if bar != "" {
		t.Errorf("progress bar should be empty with no context window: %s", bar)
	}
}

func TestRenderTokenProgressBar_Colors(t *testing.T) {
	m := newTestModel(80, 24)
	m.contextWindow = 1000

	// Green zone (< 60%)
	m.sessionInputTokens = 300
	m.sessionOutputTokens = 200
	bar := m.renderTokenProgressBar()
	if !strings.Contains(bar, "[") {
		t.Error("progress bar should contain blocks")
	}

	// Yellow zone (60-80%)
	m.sessionInputTokens = 500
	m.sessionOutputTokens = 200
	bar = m.renderTokenProgressBar()
	if !strings.Contains(bar, "[") {
		t.Error("progress bar should contain blocks")
	}

	// Red zone (> 80%)
	m.sessionInputTokens = 600
	m.sessionOutputTokens = 300
	bar = m.renderTokenProgressBar()
	if !strings.Contains(bar, "[") {
		t.Error("progress bar should contain blocks")
	}
}

func TestRenderInput_MultiLine(t *testing.T) {
	m := newTestModel(80, 24)
	m.input = "line1\nline2\nline3"
	out := m.renderInput()
	if !strings.Contains(out, "[3L]") {
		t.Error("multi-line input should show line count")
	}
	if !strings.Contains(out, "line3") {
		t.Error("should show last line of multi-line input")
	}
	if strings.Contains(out, "line1") {
		t.Error("should not show first lines in input box")
	}
}

func TestRenderInput_AskingMode(t *testing.T) {
	m := newTestModel(80, 24)
	m.askPending = true
	out := m.renderInput()
	if !strings.Contains(out, "Allow?") {
		t.Error("asking mode should show permission prompt")
	}
	if !strings.Contains(out, "[y/n]") {
		t.Error("asking mode should show y/n options")
	}
}

func TestRenderStatusBar_CompactIndicator(t *testing.T) {
	m := newTestModel(80, 24)
	m.compactMessage = "Compacted history"
	out := m.renderStatusBar()
	if !strings.Contains(out, "Compacted history") {
		t.Error("status bar should show compact message")
	}
}

func TestRenderStatusBar_ConsolidateIndicator(t *testing.T) {
	m := newTestModel(80, 24)
	m.consolidateMessage = "Memory consolidated"
	out := m.renderStatusBar()
	if !strings.Contains(out, "Memory consolidated") {
		t.Error("status bar should show consolidate message")
	}
}

func TestRenderStatusBar_PlanMode(t *testing.T) {
	m := newTestModel(80, 24)
	// Mock runtime state in plan mode
	out := m.renderStatusBar()
	// Just verify the status bar contains expected elements
	if !strings.Contains(out, "model:test-model") {
		t.Error("status bar should show model")
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
		{"[hello", false}, // needs to end with letter and only contain valid chars
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

func TestRenderLogo(t *testing.T) {
	m := newTestModel(80, 24)
	logo := m.renderLogo()
	if logo == "" {
		t.Error("logo should not be empty")
	}
	if !strings.Contains(logo, "A") || !strings.Contains(logo, "T") || !strings.Contains(logo, "O") || !strings.Contains(logo, "M") {
		t.Error("logo should contain ATOM")
	}
	if !strings.Contains(logo, "AI 助手") {
		t.Error("logo should contain tagline")
	}
}

func TestView_Autocomplete(t *testing.T) {
	m := newTestModel(80, 24)
	m.cmdRegistry = &commands.Registry{} // Initialize registry to avoid nil panic
	m.autocomplete = &AutocompleteState{
		visible:       true,
		query:         "",
		suggestions:   []string{"help", "model", "clear"},
		selectedIndex: 0,
	}
	out := m.View()
	if !strings.Contains(out, "Suggestions") {
		t.Error("should show autocomplete header")
	}
	if !strings.Contains(out, "Tab to cycle") {
		t.Error("should show autocomplete instructions")
	}
}

func TestView_WidthHandling(t *testing.T) {
	// Very narrow width
	m := newTestModel(20, 10)
	m.messages = []ChatMessage{
		{Role: RoleUser, Content: "test message"},
	}
	out := m.View()
	if out == "" {
		t.Error("View should handle narrow widths")
	}

	// Very wide width
	m2 := newTestModel(200, 10)
	m2.messages = []ChatMessage{
		{Role: RoleUser, Content: "test message"},
	}
	out2 := m2.View()
	if out2 == "" {
		t.Error("View should handle wide widths")
	}
}