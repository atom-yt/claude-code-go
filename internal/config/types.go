// Package config handles settings loading and merging.
package config

// Settings holds fully resolved runtime configuration.
type Settings struct {
	Model       string
	APIKey      string
	Provider    string // "anthropic" (default) | "openai" | "kimi" | custom
	BaseURL     string // overrides the provider's default base URL
	Verbose     bool
	Permissions PermissionsConfig
	Hooks       map[string][]HookMatcherConfig // keyed by event name
	MCPServers  map[string]MCPServerConfig

	// Auto-compact configuration
	AutoCompact       bool    // Enable/disable auto-compact (default: true)
	CompactThreshold  float64 // Percentage threshold (0.0-1.0, default: 0.8)
	CompactCooldown   int     // Cooldown time in minutes (default: 5)
	CompactKeepRecent int     // Number of recent messages to keep (default: 10)
	ContextWindow     int     // Override context window size in tokens

	// Auto-dream configuration
	AutoDreamEnabled       bool   // Enable/disable auto-dream (default: false)
	AutoMemoryDirectory    string // Optional custom memory directory path
	MinConsolidateHours    int    // Min hours since last consolidation (default: 24)
	MinConsolidateSessions int    // Min sessions to trigger consolidation (default: 5)
}

// PermissionsConfig mirrors the permissions block in settings.json.
type PermissionsConfig struct {
	DefaultMode string       `json:"defaultMode"`
	Allow       []RuleConfig `json:"allow"`
	Deny        []RuleConfig `json:"deny"`
	Ask         []RuleConfig `json:"ask"`
}

// RuleConfig is one permission rule as stored in settings.json.
type RuleConfig struct {
	Tool    string `json:"tool"`
	Path    string `json:"path"`
	Command string `json:"command"`
}

// HookCommandConfig is one hook step as stored in settings.json.
type HookCommandConfig struct {
	Type    string            `json:"type"`
	Command string            `json:"command"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Timeout int               `json:"timeout"`
}

// HookMatcherConfig is one hook matcher entry as stored in settings.json.
type HookMatcherConfig struct {
	Matcher string              `json:"matcher"`
	Hooks   []HookCommandConfig `json:"hooks"`
}

// MCPServerConfig describes how to connect to one MCP server.
type MCPServerConfig struct {
	Type    string            `json:"type"`    // "stdio" | "sse" (sse not implemented yet)
	Command string            `json:"command"` // for stdio: executable
	Args    []string          `json:"args"`    // for stdio: arguments
	Env     []string          `json:"env"`     // extra env vars ("KEY=VALUE")
	URL     string            `json:"url"`     // for sse
	Headers map[string]string `json:"headers"` // for sse
	Trust   string            `json:"trust"`   // "full" | "limited" | "untrusted" (default: "untrusted")
}

// Trust levels for MCP servers.
const (
	TrustFull     = "full"     // No permission prompts for MCP tools
	TrustLimited  = "limited"  // Prompt for potentially dangerous operations
	TrustUntrusted = "untrusted" // Prompt for all operations (default)
)

// settingsFile mirrors the full JSON structure of ~/.claude/settings.json.
type settingsFile struct {
	Model       string                         `json:"model"`
	APIKey      string                         `json:"apiKey"`
	Provider    string                         `json:"provider"`
	BaseURL     string                         `json:"baseURL"`
	Permissions PermissionsConfig              `json:"permissions"`
	Hooks       map[string][]HookMatcherConfig `json:"hooks"`
	MCPServers  map[string]MCPServerConfig     `json:"mcpServers"`
	Env         map[string]string              `json:"env"`

	// Auto-compact configuration
	AutoCompact       bool    `json:"autoCompact"`
	CompactThreshold  float64 `json:"compactThreshold"`
	CompactCooldown   int     `json:"compactCooldown"`
	CompactKeepRecent int     `json:"compactKeepRecent"`
	ContextWindow     int     `json:"contextWindow"`

	// Auto-dream configuration
	AutoDreamEnabled       bool   `json:"autoDreamEnabled"`
	AutoMemoryDirectory    string `json:"autoMemoryDirectory"`
	MinConsolidateHours    int    `json:"minConsolidateHours"`
	MinConsolidateSessions int    `json:"minConsolidateSessions"`
}
