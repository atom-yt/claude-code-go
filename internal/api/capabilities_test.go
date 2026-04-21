package api

import "testing"

func TestGetCapabilities(t *testing.T) {
	tests := []struct {
		provider string
		model    string
		want     Capabilities
	}{
		{"anthropic", "claude-sonnet-4-6", Capabilities{ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false}},
		{"openai", "gpt-4o", Capabilities{ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false}},
		{"openai", "o1", Capabilities{ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: false, Reasoning: true}},
		{"openai", "o3-mini", Capabilities{ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: false, Reasoning: true}},
		{"deepseek", "deepseek-chat", Capabilities{ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true, Reasoning: false}},
		{"unknown", "unknown-model", Capabilities{ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true, Reasoning: false}},
	}

	for _, tt := range tests {
		got := GetCapabilities(tt.provider, tt.model)
		if got != tt.want {
			t.Errorf("GetCapabilities(%q, %q) = %+v, want %+v", tt.provider, tt.model, got, tt.want)
		}
	}
}

func TestHasCapability(t *testing.T) {
	tests := []struct {
		provider    string
		model       string
		capability  string
		want        bool
	}{
		{"anthropic", "claude-sonnet-4-6", "vision", true},
		{"anthropic", "claude-sonnet-4-6", "reasoning", false},
		{"openai", "o1", "reasoning", true},
		{"openai", "o1", "streaming", false},
		{"openai", "gpt-4o", "parallel_tool_calls", true},
		{"deepseek", "deepseek-chat", "parallel", false},
		{"unknown", "unknown", "tooluse", true},
		{"unknown", "unknown", "vision", false},
	}

	for _, tt := range tests {
		got := HasCapability(tt.provider, tt.model, tt.capability)
		if got != tt.want {
			t.Errorf("HasCapability(%q, %q, %q) = %v, want %v", tt.provider, tt.model, tt.capability, got, tt.want)
		}
	}
}