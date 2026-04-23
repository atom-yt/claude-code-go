package compact

import (
	"strings"

	"github.com/atom-yt/claude-code-go/internal/api"
)

// Config for micro-compact operations.
type MicroConfig struct {
	MaxMessages        int     // Max messages before micro-compact triggers (default: 20)
	MaxToolOutputChars int     // Max chars per tool output before truncation (default: 200)
	MergeConsecutive   bool    // Merge consecutive tool results from same tool
	TokenEstimateRatio float64 // Char to token ratio for estimation (default: 4)
}

// DefaultMicroConfig returns sensible defaults.
func DefaultMicroConfig() MicroConfig {
	return MicroConfig{
		MaxMessages:        20,
		MaxToolOutputChars: 200,
		MergeConsecutive:   true,
		TokenEstimateRatio: 4.0,
	}
}

// MicroCompactor performs lightweight, immediate compression of conversation history.
// Unlike the full Compact service which uses LLM, this uses heuristics for speed.
type MicroCompactor struct {
	config MicroConfig
}

// NewMicroCompactor creates a new micro-compactor.
func NewMicroCompactor(cfg MicroConfig) *MicroCompactor {
	if cfg.TokenEstimateRatio <= 0 {
		cfg.TokenEstimateRatio = 4.0
	}
	return &MicroCompactor{config: cfg}
}

// EstimateTokens estimates token count from character count using a ratio.
// This is a rough approximation (typically 1 token ≈ 4 chars for English).
func (m *MicroCompactor) EstimateTokens(text string) int {
	// Quick approximation: chars / ratio
	chars := len(text)
	tokens := int(float64(chars) / m.config.TokenEstimateRatio)
	if tokens < 1 {
		return 1
	}
	return tokens
}

// EstimateMessageTokens estimates total tokens in a message.
func (m *MicroCompactor) EstimateMessageTokens(msg api.Message) int {
	total := 0
	for _, block := range msg.Content {
		if block.Type == "text" {
			total += m.EstimateTokens(block.Text)
		} else if block.Type == "tool_result" {
			total += m.EstimateTokens(block.Content)
		}
		// Add small overhead for tool_use blocks
		if block.Type == "tool_use" {
			total += 10 // Rough overhead
		}
	}
	return total
}

// EstimateHistoryTokens estimates total tokens in history.
func (m *MicroCompactor) EstimateHistoryTokens(history []api.Message) int {
	total := 0
	for _, msg := range history {
		total += m.EstimateMessageTokens(msg)
	}
	return total
}

// ShouldCompact determines if micro-compact should be triggered.
func (m *MicroCompactor) ShouldCompact(history []api.Message, thresholdTokens int) bool {
	if len(history) >= m.config.MaxMessages {
		return true
	}
	if thresholdTokens > 0 && m.EstimateHistoryTokens(history) >= thresholdTokens {
		return true
	}
	return false
}

// Compact performs micro-compaction on history.
// It truncates tool outputs and merges consecutive tool results.
func (m *MicroCompactor) Compact(history []api.Message) []api.Message {
	if len(history) == 0 {
		return history
	}

	compacted := make([]api.Message, 0, len(history))

	for _, msg := range history {
		compactedMsg := m.compactMessage(msg)
		compacted = append(compacted, compactedMsg)
	}

	// Optionally merge consecutive tool result messages
	if m.config.MergeConsecutive {
		compacted = m.mergeConsecutiveToolResults(compacted)
	}

	return compacted
}

// CompactWithCompact performs micro-compaction with a soft message limit.
// It keeps the most recent `keep` messages and micro-compacts older ones.
func (m *MicroCompactor) CompactWithCompact(history []api.Message, keep int) []api.Message {
	if len(history) <= keep {
		return m.Compact(history)
	}

	// Keep most recent messages as-is
	recent := history[len(history)-keep:]

	// Micro-compact older messages
	older := history[:len(history)-keep]
	compactedOlder := m.Compact(older)

	return append(compactedOlder, recent...)
}

// compactMessage compresses a single message.
func (m *MicroCompactor) compactMessage(msg api.Message) api.Message {
	compacted := msg
	compacted.Content = make([]api.ContentBlock, len(msg.Content))

	for i, block := range msg.Content {
		if block.Type == "tool_result" && len(block.Content) > m.config.MaxToolOutputChars {
			// Truncate tool output
			truncated := block.Content[:m.config.MaxToolOutputChars]
			truncated += "... (truncated)"
			compacted.Content[i] = block
			compacted.Content[i].Content = truncated
		} else {
			compacted.Content[i] = block
		}
	}

	return compacted
}

// mergeConsecutiveToolResults merges consecutive user messages containing tool results.
func (m *MicroCompactor) mergeConsecutiveToolResults(history []api.Message) []api.Message {
	if len(history) < 2 {
		return history
	}

	merged := make([]api.Message, 0, len(history))

	for i := 0; i < len(history); i++ {
		msg := history[i]

		// Only merge user messages with tool results
		if msg.Role != api.RoleUser || !hasToolResult(msg) {
			merged = append(merged, msg)
			continue
		}

		// Look ahead to see if next message is also user with tool results
		j := i + 1
		for j < len(history) && history[j].Role == api.RoleUser && hasToolResult(history[j]) {
			j++
		}

		// If we found consecutive tool results, merge them
		if j > i+1 {
			mergedMsg := msg
			mergedMsg.Content = append([]api.ContentBlock(nil), msg.Content...)
			for k := i + 1; k < j; k++ {
				mergedMsg.Content = append(mergedMsg.Content, history[k].Content...)
			}
			merged = append(merged, mergedMsg)
			i = j - 1 // Skip ahead
		} else {
			merged = append(merged, msg)
		}
	}

	return merged
}

// hasToolResult checks if a message contains tool result content blocks.
func hasToolResult(msg api.Message) bool {
	for _, block := range msg.Content {
		if block.Type == "tool_result" {
			return true
		}
	}
	return false
}

// TruncateText safely truncates text to a maximum length.
// The maxLen includes the ellipsis "..." when truncating.
func TruncateText(text string, maxLen int) string {
	if maxLen <= 3 {
		return "..." // At minimum return ellipsis
	}
	if len(text) <= maxLen {
		return text
	}

	// Reserve space for ellipsis
	truncateTo := maxLen - 3

	// Try to truncate at a word boundary
	truncated := text[:truncateTo]
	lastSpace := strings.LastIndex(truncated, " ")
	lastNewline := strings.LastIndex(truncated, "\n")
	lastBreak := max(lastSpace, lastNewline)

	if lastBreak > truncateTo/2 {
		// If we found a good break point, use it
		truncated = text[:lastBreak]
	}

	return truncated + "..."
}

// DeduplicateSimilarToolResults removes duplicate tool results from the same tool.
func DeduplicateSimilarToolResults(history []api.Message, similarityThreshold float64) []api.Message {
	// This is a placeholder for more sophisticated deduplication
	// For now, just return history as-is
	return history
}

// SummaryTokenEstimate estimates tokens needed for a summary of given history.
// This uses a rough heuristic: ~10% of original tokens for summary.
func (m *MicroCompactor) SummaryTokenEstimate(history []api.Message) int {
	originalTokens := m.EstimateHistoryTokens(history)
	// Summary typically needs ~10-15% of original context
	return max(100, originalTokens/10)
}

// CompactRatio returns the compression ratio (after / before).
// Lower values mean more compression.
func (m *MicroCompactor) CompactRatio(original, compacted []api.Message) float64 {
	if len(original) == 0 {
		return 0
	}
	origTokens := m.EstimateHistoryTokens(original)
	compTokens := m.EstimateHistoryTokens(compacted)
	if origTokens == 0 {
		return 0
	}
	return float64(compTokens) / float64(origTokens)
}