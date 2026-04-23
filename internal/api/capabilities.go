package api

// Capabilities declares what features a provider/model combination supports.
// The agent loop can use this to degrade gracefully (e.g. skip tool specs if
// ToolUse is false, avoid parallel calls if ParallelToolCalls is false).
type Capabilities struct {
	ToolUse           bool
	ParallelToolCalls bool
	Vision            bool
	Streaming         bool
	Reasoning         bool // For o1/o3-style reasoning models
}

// ModelCapabilities maps model names to their specific capabilities.
// This allows fine-grained capability control per model, not just per provider.
var ModelCapabilities = map[string]Capabilities{
	// Claude models
	"claude-sonnet-4-6":     {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},
	"claude-opus-4-6":       {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},
	"claude-3-opus":         {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},
	"claude-3-sonnet":       {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},
	"claude-3-haiku":        {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},
	"claude-3.5-sonnet":     {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},
	"claude-3.5-haiku":      {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},

	// OpenAI models
	"gpt-4o":                {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},
	"gpt-4o-mini":           {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},
	"gpt-4-turbo":           {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},
	"gpt-4":                 {ToolUse: true, ParallelToolCalls: false, Vision: true, Streaming: true, Reasoning: false},
	"gpt-3.5-turbo":         {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},

	// OpenAI reasoning models
	"o1":                     {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: false, Reasoning: true},
	"o1-mini":                {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: false, Reasoning: true},
	"o1-preview":             {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: false, Reasoning: true},
	"o3":                     {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: false, Reasoning: true},
	"o3-mini":                {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: false, Reasoning: true},

	// DeepSeek models
	"deepseek-chat":          {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true, Reasoning: false},
	"deepseek-coder":         {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true, Reasoning: false},

	// Moonshot models
	"moonshot-v1":           {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true, Reasoning: false},
	"moonshot-v1-8k":        {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true, Reasoning: false},
	"moonshot-v1-32k":       {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true, Reasoning: false},
	"moonshot-v1-128k":      {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true, Reasoning: false},

	// Qwen models
	"qwen-plus":             {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},
	"qwen-max":              {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},
	"qwen":                  {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},
	"qwen2":                 {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},

	// GLM models
	"glm-5":                 {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},
	"glm-4":                 {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},

	// Ark models
	"ark-code-latest":        {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},
}

// ProviderCapabilities maps provider names to their default capabilities.
// Used as fallback when model-specific capabilities are not available.
// Unknown providers default to a conservative set (ToolUse + Streaming only).
var ProviderCapabilities = map[string]Capabilities{
	"anthropic":     {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},
	"openai":        {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true, Reasoning: false},
	"codex":         {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},
	"deepseek":      {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true, Reasoning: false},
	"kimi":          {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true, Reasoning: false},
	"moonshot":      {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true, Reasoning: false},
	"qwen":          {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},
	"ark":           {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},
	"ark-openai":    {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},
	"ark-anthropic": {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true, Reasoning: false},
}

// DefaultCapabilities is used for unknown providers.
var DefaultCapabilities = Capabilities{
	ToolUse:           true,
	ParallelToolCalls: false,
	Vision:            false,
	Streaming:         true,
	Reasoning:         false,
}

// GetCapabilities returns the capabilities for the given model name.
// First checks model-specific capabilities, then falls back to provider defaults.
func GetCapabilities(provider, model string) Capabilities {
	// Try model-specific capabilities first
	if c, ok := ModelCapabilities[model]; ok {
		return c
	}

	// Fall back to provider capabilities
	if c, ok := ProviderCapabilities[provider]; ok {
		return c
	}

	return DefaultCapabilities
}

// HasCapability returns true if the model has the requested capability.
func HasCapability(provider, model string, capability string) bool {
	caps := GetCapabilities(provider, model)
	switch capability {
	case "tooluse", "tool_use":
		return caps.ToolUse
	case "parallel", "parallel_tool_calls":
		return caps.ParallelToolCalls
	case "vision":
		return caps.Vision
	case "streaming":
		return caps.Streaming
	case "reasoning":
		return caps.Reasoning
	default:
		return false
	}
}
