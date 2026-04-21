package api

import "testing"

func TestEstimateTokenCount(t *testing.T) {
	tests := []struct {
		text string
		want int // rough estimate, so we check within range
	}{
		{"", 0},
		{"hello", 2},                 // ~5 chars / 4 ≈ 1-2
		{"hello world", 3},           // ~11 chars / 4 ≈ 3
		{"你好", 2},                   // Chinese chars ~1.5 per char, so 2
		{"你好世界", 4},               // 4 Chinese chars ≈ 4-5 tokens
		{"The quick brown fox jumps over the lazy dog.", 10}, // ~43 chars / 4 ≈ 10-11
	}

	for _, tt := range tests {
		got := EstimateTokenCount(tt.text)
		// Allow 50% variance since estimation is approximate
		min := tt.want / 2
		max := tt.want * 3 / 2
		if got < min || got > max {
			t.Errorf("EstimateTokenCount(%q) = %d, want approx %d (%d-%d)", tt.text, got, tt.want, min, max)
		}
	}
}

func TestTokenBudget(t *testing.T) {
	tb := NewTokenBudget(100000)

	if tb.Total() != 100000 {
		t.Errorf("Total() = %d, want 100000", tb.Total())
	}

	if tb.Used() != 0 {
		t.Errorf("Used() = %d, want 0", tb.Used())
	}

	if tb.Remaining() != 75000 { // 100k - 25% reserve
		t.Errorf("Remaining() = %d, want 75000", tb.Remaining())
	}

	tb.AddUsage(5000)
	if tb.Used() != 5000 {
		t.Errorf("Used() = %d, want 5000", tb.Used())
	}

	if tb.Remaining() != 70000 {
		t.Errorf("Remaining() = %d, want 70000", tb.Remaining())
	}
}

func TestShouldCompact(t *testing.T) {
	tb := NewTokenBudget(100000)

	if tb.ShouldCompact(0.8) {
		t.Error("ShouldCompact(0.8) = true, want false at 0% usage")
	}

	tb.AddUsage(80000)
	if !tb.ShouldCompact(0.8) {
		t.Error("ShouldCompact(0.8) = false, want true at 80% usage")
	}

	if tb.ShouldCompact(0.9) {
		t.Error("ShouldCompact(0.9) = true, want false at 80% usage < 90%")
	}
}

func TestEstimateMessageTokens(t *testing.T) {
	msg := Message{
		Role: "user",
		Content: []ContentBlock{
			{Type: "text", Text: "hello world"},
		},
	}
	tokens := EstimateMessageTokens(msg)
	if tokens < 5 || tokens > 10 {
		t.Errorf("EstimateMessageTokens() = %d, want 5-10", tokens)
	}
}

func TestEstimateHistoryTokens(t *testing.T) {
	messages := []Message{
		TextMessage("user", "hello"),
		TextMessage("assistant", "hi there"),
		TextMessage("user", "how are you"),
	}
	tokens := EstimateHistoryTokens(messages)
	// Each message ~4-8 tokens + overhead
	if tokens < 12 || tokens > 40 {
		t.Errorf("EstimateHistoryTokens() = %d, want 12-40", tokens)
	}
}
