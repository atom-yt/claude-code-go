package providers

import "testing"

func TestResolveProtocol(t *testing.T) {
	if got := ResolveProtocol("ark-anthropic"); got != ProtocolAnthropic {
		t.Fatalf("want anthropic protocol, got %q", got)
	}
	if got := ResolveProtocol("qwen"); got != ProtocolOpenAI {
		t.Fatalf("want openai protocol, got %q", got)
	}
}

func TestContextWindowFallback(t *testing.T) {
	if got := ContextWindow("claude-sonnet-4-6-20250514"); got != 200000 {
		t.Fatalf("want 200000, got %d", got)
	}
	if got := ContextWindow("unknown-model"); got != 128000 {
		t.Fatalf("want fallback 128000, got %d", got)
	}
}
