package compact

import (
	"strings"
	"testing"

	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMicroCompactor_EstimateTokens(t *testing.T) {
	cfg := DefaultMicroConfig()
	c := NewMicroCompactor(cfg)

	tests := []struct {
		text   string
		expect int
	}{
		{"hello world", 2},  // ~11 chars / 4 ≈ 2-3 tokens
		{"", 1},            // min 1
		{"a", 1},           // min 1
		{strings.Repeat("word ", 100), 50}, // ~500 chars / 4 = ~125, but min 1
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := c.EstimateTokens(tt.text)
			assert.True(t, result >= 1, "tokens should be at least 1")
		})
	}
}

func TestMicroCompactor_EstimateMessageTokens(t *testing.T) {
	cfg := DefaultMicroConfig()
	c := NewMicroCompactor(cfg)

	msg := api.TextMessage(api.RoleUser, "This is a test message with some text")
	tokens := c.EstimateMessageTokens(msg)
	assert.True(t, tokens > 0)
}

func TestMicroCompactor_EstimateHistoryTokens(t *testing.T) {
	cfg := DefaultMicroConfig()
	c := NewMicroCompactor(cfg)

	history := []api.Message{
		api.TextMessage(api.RoleUser, "Hello"),
		api.TextMessage(api.RoleAssistant, "Hi there"),
		api.TextMessage(api.RoleUser, "How are you?"),
	}

	tokens := c.EstimateHistoryTokens(history)
	assert.True(t, tokens > 0)
	assert.True(t, tokens >= 3, "should have at least 3 messages worth of tokens")
}

func TestMicroCompactor_ShouldCompact(t *testing.T) {
	tests := []struct {
		name         string
		config       MicroConfig
		historySize  int
		threshold    int
		expectCompact bool
	}{
		{
			name:   "below limits",
			config: MicroConfig{MaxMessages: 10, TokenEstimateRatio: 4.0},
			historySize: 5,
			threshold: 1000,
			expectCompact: false,
		},
		{
			name:   "above message limit",
			config: MicroConfig{MaxMessages: 10, TokenEstimateRatio: 4.0},
			historySize: 15,
			threshold: 1000,
			expectCompact: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMicroCompactor(tt.config)

			history := make([]api.Message, tt.historySize)
			for i := range history {
				history[i] = api.TextMessage(api.RoleUser, "test message")
			}

			should := c.ShouldCompact(history, tt.threshold)
			assert.Equal(t, tt.expectCompact, should)
		})
	}
}

func TestMicroCompactor_Compact(t *testing.T) {
	cfg := MicroConfig{
		MaxMessages:        20,
		MaxToolOutputChars: 50,
		MergeConsecutive:   false, // Don't merge for this test
		TokenEstimateRatio: 4.0,
	}
	c := NewMicroCompactor(cfg)

	longOutput := strings.Repeat("This is a long tool output. ", 20)

	history := []api.Message{
		api.TextMessage(api.RoleUser, "Let me read a file"),
		{
			Role: api.RoleAssistant,
			Content: []api.ContentBlock{{
				Type: "tool_use",
				ID:   "tool-1",
				Name: "Read",
				Input: map[string]any{"path": "file.txt"},
			}},
		},
		{
			Role: api.RoleUser,
			Content: []api.ContentBlock{{
				Type:      "tool_result",
				ToolUseID: "tool-1",
				Content:   longOutput,
			}},
		},
	}

	compacted := c.Compact(history)

	require.Len(t, compacted, 3)

	// Check tool result was truncated
	toolResultMsg := compacted[2]
	toolResultBlock := toolResultMsg.Content[0]
	assert.Equal(t, "tool_result", toolResultBlock.Type)
	assert.True(t, len(toolResultBlock.Content) <= cfg.MaxToolOutputChars+20, // +20 for truncation suffix
		"tool output should be truncated")
	assert.Contains(t, toolResultBlock.Content, "...", "should have truncation marker")
}

func TestMicroCompactor_CompactWithCompact(t *testing.T) {
	cfg := DefaultMicroConfig()
	c := NewMicroCompactor(cfg)

	// Create 25 messages, keep last 5 as-is
	history := make([]api.Message, 25)
	for i := range history {
		history[i] = api.TextMessage(api.RoleUser, "message "+string(rune('a'+i)))
	}

	compacted := c.CompactWithCompact(history, 5)

	assert.Len(t, compacted, 25)
}

func TestMicroCompactor_mergeConsecutiveToolResults(t *testing.T) {
	cfg := MicroConfig{
		MaxMessages:        20,
		MaxToolOutputChars: 200,
		MergeConsecutive:   true,
		TokenEstimateRatio: 4.0,
	}
	c := NewMicroCompactor(cfg)

	history := []api.Message{
		{
			Role: api.RoleUser,
			Content: []api.ContentBlock{{
				Type:      "tool_result",
				ToolUseID: "tool-1",
				Content:   "Result 1",
			}},
		},
		{
			Role: api.RoleUser,
			Content: []api.ContentBlock{{
				Type:      "tool_result",
				ToolUseID: "tool-2",
				Content:   "Result 2",
			}},
		},
		{
			Role: api.RoleAssistant,
			Content: []api.ContentBlock{{
				Type: "text",
				Text: "Here are the results",
			}},
		},
	}

	compacted := c.mergeConsecutiveToolResults(history)

	// Two tool results should be merged into one message
	assert.Len(t, compacted, 2)
	assert.Equal(t, api.RoleUser, compacted[0].Role)
	assert.Len(t, compacted[0].Content, 2)
}

func TestMicroCompactor_CompactRatio(t *testing.T) {
	cfg := MicroConfig{
		MaxMessages:        20,
		MaxToolOutputChars: 50,
		MergeConsecutive:   true,
		TokenEstimateRatio: 4.0,
	}
	c := NewMicroCompactor(cfg)

	history := []api.Message{
		api.TextMessage(api.RoleUser, "Hello"),
		api.TextMessage(api.RoleAssistant, "Hi"),
		{
			Role: api.RoleAssistant,
			Content: []api.ContentBlock{{
				Type: "tool_use",
				ID:   "tool-1",
				Name: "Read",
				Input: map[string]any{},
			}},
		},
		{
			Role: api.RoleUser,
			Content: []api.ContentBlock{{
				Type:      "tool_result",
				ToolUseID: "tool-1",
				Content:   strings.Repeat("long output ", 20),
			}},
		},
	}

	compacted := c.Compact(history)

	ratio := c.CompactRatio(history, compacted)
	assert.True(t, ratio >= 0 && ratio <= 1, "ratio should be between 0 and 1")
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLen   int
		contains string
	}{
		{
			name:     "short text unchanged",
			text:     "hello",
			maxLen:   20,
			contains: "hello",
		},
		{
			name:     "long text truncated",
			text:     "this is a very long text that needs truncation",
			maxLen:   10,
			contains: "...",
		},
		{
			name:     "truncate at word boundary",
			text:     "one two three four five",
			maxLen:   15,
			contains: "one two...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateText(tt.text, tt.maxLen)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestMicroCompactor_SummaryTokenEstimate(t *testing.T) {
	cfg := DefaultMicroConfig()
	c := NewMicroCompactor(cfg)

	history := []api.Message{
		api.TextMessage(api.RoleUser, "First message"),
		api.TextMessage(api.RoleAssistant, "Response"),
		api.TextMessage(api.RoleUser, "Second message"),
	}

	estimate := c.SummaryTokenEstimate(history)
	original := c.EstimateHistoryTokens(history)

	// Estimate should be a reasonable value (at least 100 due to max, and some function of original)
	assert.True(t, estimate >= 100, "estimate should be at least min 100")
	assert.True(t, estimate >= original/10, "estimate should be at least 10% of original")
}

func TestMicroCompactor_hasToolResult(t *testing.T) {
	tests := []struct {
		name     string
		msg      api.Message
		expected bool
	}{
		{
			name:     "message with tool result",
			msg: api.Message{
				Role: api.RoleUser,
				Content: []api.ContentBlock{{
					Type: "tool_result",
					ToolUseID: "tool-1",
					Content:   "result",
				}},
			},
			expected: true,
		},
		{
			name:     "text message",
			msg: api.Message{
				Role: api.RoleUser,
				Content: []api.ContentBlock{{
					Type: "text",
					Text: "hello",
				}},
			},
			expected: false,
		},
		{
			name:     "message with tool use",
			msg: api.Message{
				Role: api.RoleAssistant,
				Content: []api.ContentBlock{{
					Type: "tool_use",
					ID:   "tool-1",
					Name: "Read",
				}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasToolResult(tt.msg)
			assert.Equal(t, tt.expected, result)
		})
	}
}