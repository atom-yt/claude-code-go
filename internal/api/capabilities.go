package api

// Capabilities declares what features a provider/model combination supports.
// The agent loop can use this to degrade gracefully (e.g. skip tool specs if
// ToolUse is false, avoid parallel calls if ParallelToolCalls is false).
type Capabilities struct {
	ToolUse           bool
	ParallelToolCalls bool
	Vision            bool
	Streaming         bool
}

// ProviderCapabilities maps provider names to their default capabilities.
// Unknown providers default to a conservative set (ToolUse + Streaming only).
var ProviderCapabilities = map[string]Capabilities{
	"anthropic":     {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true},
	"openai":        {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true},
	"codex":         {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true},
	"deepseek":      {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true},
	"kimi":          {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true},
	"moonshot":      {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true},
	"qwen":          {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true},
	"ark":           {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true},
	"ark-openai":    {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true},
	"ark-anthropic": {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true},
}

// DefaultCapabilities is used for unknown providers.
var DefaultCapabilities = Capabilities{
	ToolUse:           true,
	ParallelToolCalls: false,
	Vision:            false,
	Streaming:         true,
}

// GetCapabilities returns the capabilities for the given provider name.
func GetCapabilities(provider string) Capabilities {
	if c, ok := ProviderCapabilities[provider]; ok {
		return c
	}
	return DefaultCapabilities
}
