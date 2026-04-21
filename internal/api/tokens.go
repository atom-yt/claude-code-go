package api

import (
	"unicode/utf8"
)

// EstimateTokenCount provides a rough token count estimation.
// For most models, tokens ≈ characters / 4 for English, and fewer for Chinese.
// This is a simplified approximation suitable for budgeting, not exact counts.
func EstimateTokenCount(text string) int {
	if text == "" {
		return 0
	}

	// Count characters first
	charCount := utf8.RuneCountInString(text)

	// Estimate based on character patterns
	// Chinese/Japanese characters typically use ~1.5-2 tokens per character
	// ASCII text uses ~4 characters per token
	var chineseCount int
	var asciiCount int
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF || // CJK Unified Ideographs
			r >= 0x3400 && r <= 0x4DBF || // CJK Unified Ideographs Extension A
			r >= 0x20000 && r <= 0x2A6DF || // CJK Unified Ideographs Extension B
			r >= 0xF900 && r <= 0xFAFF || // CJK Compatibility Ideographs
			r >= 0x3040 && r <= 0x30FF || // Hiragana/Katakana
			r >= 0x31F0 && r <= 0x31FF || // Katakana Phonetic Extensions
			r >= 0xAC00 && r <= 0xD7AF { // Hangul Syllables
			chineseCount++
		} else if r < 128 {
			asciiCount++
		}
	}

	// Estimate tokens: ~4 chars/token for ASCII, ~1.5 chars/token for CJK
	tokens := (asciiCount / 4) + (chineseCount * 2 / 3)
	remaining := charCount - chineseCount - asciiCount
	tokens += remaining / 3 // Other unicode chars

	if tokens < 1 {
		return 1
	}
	return tokens
}

// EstimateMessageTokens estimates tokens for a single message.
func EstimateMessageTokens(msg Message) int {
	total := EstimateTokenCount(string(msg.Role))
	total += EstimateTokenCountFromBlocks(msg.Content)

	// Small overhead for message structure
	return total + 4
}

// EstimateTokenCountFromBlocks estimates tokens from ContentBlock array.
func EstimateTokenCountFromBlocks(blocks []ContentBlock) int {
	var total int
	for _, block := range blocks {
		switch block.Type {
		case "text":
			total += EstimateTokenCount(block.Text)
		case "image":
			// Images typically use ~85-258 tokens depending on detail
			// Use a conservative estimate
			total += 85
		}
	}
	return total
}

// EstimateToolUseTokens estimates tokens for a tool_use content block.
func EstimateToolUseTokens(toolName string, inputJSON string) int {
	// Tool name and input JSON
	tokens := EstimateTokenCount(toolName) + EstimateTokenCount(inputJSON)

	// Overhead for tool structure
	return tokens + 10
}

// EstimateToolResultTokens estimates tokens for a tool_result content block.
func EstimateToolResultTokens(toolID, output string, isError bool) int {
	tokens := EstimateTokenCount(toolID) + EstimateTokenCount(output)

	// Add a bit more for error metadata if present
	if isError {
		tokens += 2
	}

	// Overhead for result structure
	return tokens + 5
}

// EstimateHistoryTokens estimates total tokens in message history.
func EstimateHistoryTokens(messages []Message) int {
	var total int
	for _, msg := range messages {
		total += EstimateMessageTokens(msg)
	}
	return total
}

// EstimateSystemPromptTokens estimates tokens for a system prompt.
// Adds overhead for system instruction metadata.
func EstimateSystemPromptTokens(system string) int {
	tokens := EstimateTokenCount(system)
	// System prompts have some metadata overhead
	return tokens + 10
}

// TokenBudget manages token budget for a conversation.
type TokenBudget struct {
	contextWindow int
	usedTokens   int
	reserved      int // Reserved for response (typically model-specific)
}

// NewTokenBudget creates a new token budget manager.
func NewTokenBudget(contextWindow int) *TokenBudget {
	return &TokenBudget{
		contextWindow: contextWindow,
		reserved:      contextWindow / 4, // Reserve 25% for model response
	}
}

// SetReserved sets the number of tokens to reserve for the model's response.
func (tb *TokenBudget) SetReserved(tokens int) {
	tb.reserved = tokens
}

// AddUsage adds estimated token usage to the budget.
func (tb *TokenBudget) AddUsage(tokens int) {
	tb.usedTokens += tokens
}

// Remaining returns the number of tokens available for input.
func (tb *TokenBudget) Remaining() int {
	available := tb.contextWindow - tb.usedTokens - tb.reserved
	if available < 0 {
		return 0
	}
	return available
}

// Used returns the number of tokens currently used.
func (tb *TokenBudget) Used() int {
	return tb.usedTokens
}

// Total returns the total context window size.
func (tb *TokenBudget) Total() int {
	return tb.contextWindow
}

// CapacityUsage returns the percentage of capacity used (0.0 to 1.0).
func (tb *TokenBudget) CapacityUsage() float64 {
	if tb.contextWindow == 0 {
		return 0
	}
	return float64(tb.usedTokens) / float64(tb.contextWindow)
}

// ShouldCompact returns true if the history should be compacted based on usage.
func (tb *TokenBudget) ShouldCompact(threshold float64) bool {
	return tb.CapacityUsage() >= threshold
}

// EstimateAvailableForText estimates how many more tokens of text can be added.
func (tb *TokenBudget) EstimateAvailableForText() int {
	remaining := tb.Remaining()
	// Rough estimate: ~4 chars per token
	return remaining * 4
}

// EstimateMaxHistorySize estimates the maximum number of messages that can fit
// given an average message size in tokens.
func (tb *TokenBudget) EstimateMaxHistorySize(avgMsgTokens int) int {
	if avgMsgTokens == 0 {
		avgMsgTokens = 100 // Conservative default
	}
	available := tb.Remaining()
	if available <= 0 {
		return 0
	}
	return available / avgMsgTokens
}
