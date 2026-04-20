package providers

import "strings"

// ProviderInfo describes one known provider endpoint.
type ProviderInfo struct {
	BaseURL      string
	Protocol     string
	DefaultModel string
}

const (
	ProtocolAnthropic = "anthropic"
	ProtocolOpenAI    = "openai"
)

var knownProviders = map[string]ProviderInfo{
	"openai":        {BaseURL: "https://api.openai.com/v1", Protocol: ProtocolOpenAI, DefaultModel: "gpt-4o"},
	"kimi":          {BaseURL: "https://api.moonshot.cn/v1", Protocol: ProtocolOpenAI, DefaultModel: "moonshot-v1-8k"},
	"moonshot":      {BaseURL: "https://api.moonshot.cn/v1", Protocol: ProtocolOpenAI, DefaultModel: "moonshot-v1-8k"},
	"deepseek":      {BaseURL: "https://api.deepseek.com/v1", Protocol: ProtocolOpenAI, DefaultModel: "deepseek-chat"},
	"qwen":          {BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Protocol: ProtocolOpenAI, DefaultModel: "qwen-plus"},
	"codex":         {BaseURL: "https://coder.api.visioncoder.cn/v1", Protocol: ProtocolOpenAI},
	"ark":           {BaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3", Protocol: ProtocolOpenAI, DefaultModel: "ark-code-latest"},
	"ark-openai":    {BaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3", Protocol: ProtocolOpenAI, DefaultModel: "ark-code-latest"},
	"ark-anthropic": {BaseURL: "https://ark.cn-beijing.volces.com/api/coding", Protocol: ProtocolAnthropic, DefaultModel: "ark-code-latest"},
	"anthropic":     {BaseURL: "https://api.anthropic.com", Protocol: ProtocolAnthropic, DefaultModel: "claude-sonnet-4-6"},
}

var modelContextWindows = map[string]int{
	"claude-sonnet-4-6": 200000,
	"claude-opus-4-6":   200000,
	"claude-3-opus":     200000,
	"claude-3-sonnet":   200000,
	"claude-3-haiku":    200000,
	"claude-3.5-sonnet": 200000,
	"claude-3.5-haiku":  200000,
	"gpt-4o":            128000,
	"gpt-4o-mini":       128000,
	"gpt-4-turbo":       128000,
	"gpt-4":             8192,
	"gpt-3.5-turbo":     16385,
	"o1":                200000,
	"o1-mini":           128000,
	"o1-preview":        128000,
	"o3":                200000,
	"o3-mini":           200000,
	"deepseek-chat":     128000,
	"deepseek-coder":    128000,
	"qwen":              128000,
	"qwen2":             128000,
	"qwen-max":          32768,
	"moonshot-v1":       128000,
}

// Lookup returns known metadata for the provider.
func Lookup(provider string) (ProviderInfo, bool) {
	info, ok := knownProviders[strings.ToLower(provider)]
	return info, ok
}

// ResolveBaseURL resolves the effective base URL for a provider.
func ResolveBaseURL(provider, explicit string) string {
	if explicit != "" {
		return explicit
	}
	if info, ok := Lookup(provider); ok {
		return info.BaseURL
	}
	return explicit
}

// ResolveProtocol resolves the wire protocol for a provider.
func ResolveProtocol(provider string) string {
	if info, ok := Lookup(provider); ok && info.Protocol != "" {
		return info.Protocol
	}
	if strings.EqualFold(provider, "anthropic") || provider == "" {
		return ProtocolAnthropic
	}
	return ProtocolOpenAI
}

// ContextWindow returns the context window size for a model.
func ContextWindow(model string) int {
	for knownModel, window := range modelContextWindows {
		if model == knownModel || len(model) > len(knownModel) && model[:len(knownModel)] == knownModel {
			return window
		}
	}
	return 128000
}
