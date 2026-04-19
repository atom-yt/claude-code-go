package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const defaultModel = "claude-sonnet-4-6"

// CLIFlags carries values supplied on the command line.
type CLIFlags struct {
	Model    string
	APIKey   string
	Provider string
	BaseURL  string
	Verbose  bool
}

// Load resolves configuration from (highest to lowest priority):
//  1. CLI flags
//  2. Environment variables
//  3. Project-level .claude/settings.json (cwd → parents)
//  4. User-level ~/.claude/settings.json
//  5. Built-in defaults
func Load(flags CLIFlags) Settings {
	user := loadFile(userSettingsPath())
	project := loadFile(findProjectSettings())

	// Merge: project overrides user for shared fields.
	model := firstNonEmpty(flags.Model, os.Getenv("CLAUDE_MODEL"), project.Model, user.Model, defaultModel)
	provider := firstNonEmpty(flags.Provider, os.Getenv("CLAUDE_PROVIDER"), project.Provider, user.Provider)

	// Provider-specific API key resolution.
	// ANTHROPIC_API_KEY is only used for anthropic/ark-anthropic providers.
	// Other providers use OPENAI_API_KEY or their own env var.
	var apiKey string
	switch strings.ToLower(provider) {
	case "", "anthropic", "ark-anthropic":
		apiKey = firstNonEmpty(flags.APIKey,
			os.Getenv("ANTHROPIC_API_KEY"),
			project.APIKey, user.APIKey)
	default:
		apiKey = firstNonEmpty(flags.APIKey,
			providerEnvKey(provider),
			os.Getenv("OPENAI_API_KEY"),
			project.APIKey, user.APIKey)
	}
	baseURL := firstNonEmpty(flags.BaseURL,
		providerEnvURL(provider),
		os.Getenv("CLAUDE_BASE_URL"), project.BaseURL, user.BaseURL)

	// Inject env vars from settings file.
	for k, v := range user.Env {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
	for k, v := range project.Env {
		os.Setenv(k, v) // project env overrides user env
	}

	perms := mergePermissions(user.Permissions, project.Permissions)
	hooksCfg := mergeHooks(user.Hooks, project.Hooks)
	mcpServers := mergeMCPServers(user.MCPServers, project.MCPServers)

	// Load compact configuration with project override
	autoCompact := true
	compactThreshold := 0.8
	compactCooldown := 5
	compactKeepRecent := 10
	contextWindow := 0

	// Load auto-dream configuration with project override
	autoDreamEnabled := false
	autoMemoryDirectory := ""
	minConsolidateHours := 24
	minConsolidateSessions := 5

	if project.AutoCompact != false { // Explicit false check
		autoCompact = project.AutoCompact
	}
	if user.AutoCompact != false {
		autoCompact = user.AutoCompact
	}

	if project.CompactThreshold > 0 {
		compactThreshold = project.CompactThreshold
	}
	if user.CompactThreshold > 0 {
		compactThreshold = user.CompactThreshold
	}

	if project.CompactCooldown > 0 {
		compactCooldown = project.CompactCooldown
	}
	if user.CompactCooldown > 0 {
		compactCooldown = user.CompactCooldown
	}

	if project.CompactKeepRecent > 0 {
		compactKeepRecent = project.CompactKeepRecent
	}
	if user.CompactKeepRecent > 0 {
		compactKeepRecent = user.CompactKeepRecent
	}

	if project.ContextWindow > 0 {
		contextWindow = project.ContextWindow
	}
	if user.ContextWindow > 0 {
		contextWindow = user.ContextWindow
	}

	// Load auto-dream configuration
	if project.AutoDreamEnabled {
		autoDreamEnabled = project.AutoDreamEnabled
	} else if user.AutoDreamEnabled {
		autoDreamEnabled = user.AutoDreamEnabled
	}

	if project.AutoMemoryDirectory != "" {
		autoMemoryDirectory = project.AutoMemoryDirectory
	} else if user.AutoMemoryDirectory != "" {
		autoMemoryDirectory = user.AutoMemoryDirectory
	}

	if project.MinConsolidateHours > 0 {
		minConsolidateHours = project.MinConsolidateHours
	} else if user.MinConsolidateHours > 0 {
		minConsolidateHours = user.MinConsolidateHours
	}

	if project.MinConsolidateSessions > 0 {
		minConsolidateSessions = project.MinConsolidateSessions
	} else if user.MinConsolidateSessions > 0 {
		minConsolidateSessions = user.MinConsolidateSessions
	}

	return Settings{
		Model:                  model,
		APIKey:                 apiKey,
		Provider:               provider,
		BaseURL:                baseURL,
		Verbose:                flags.Verbose,
		Permissions:            perms,
		Hooks:                  hooksCfg,
		MCPServers:             mcpServers,
		AutoCompact:            autoCompact,
		CompactThreshold:       compactThreshold,
		CompactCooldown:        compactCooldown,
		CompactKeepRecent:      compactKeepRecent,
		ContextWindow:          contextWindow,
		AutoDreamEnabled:       autoDreamEnabled,
		AutoMemoryDirectory:    autoMemoryDirectory,
		MinConsolidateHours:    minConsolidateHours,
		MinConsolidateSessions: minConsolidateSessions,
	}
}

// mergePermissions combines user and project permission configs.
// Project rules take precedence for DefaultMode; lists are concatenated.
func mergePermissions(user, project PermissionsConfig) PermissionsConfig {
	mode := user.DefaultMode
	if project.DefaultMode != "" {
		mode = project.DefaultMode
	}
	if mode == "" {
		mode = "default"
	}
	return PermissionsConfig{
		DefaultMode: mode,
		Allow:       append(user.Allow, project.Allow...),
		Deny:        append(user.Deny, project.Deny...),
		Ask:         append(user.Ask, project.Ask...),
	}
}

// userSettingsPath returns the path to ~/.claude/settings.json.
func userSettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "settings.json")
}

// findProjectSettings walks up from cwd looking for .claude/settings.json.
func findProjectSettings() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := cwd
	for {
		candidate := filepath.Join(dir, ".claude", "settings.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// loadFile reads and parses a settings JSON file.
// Returns an empty settingsFile on any error.
func loadFile(path string) settingsFile {
	if path == "" {
		return settingsFile{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return settingsFile{}
	}
	var sf settingsFile
	_ = json.Unmarshal(data, &sf)
	return sf
}

// mergeMCPServers merges server maps; project entries override user entries.
func mergeMCPServers(user, project map[string]MCPServerConfig) map[string]MCPServerConfig {
	out := make(map[string]MCPServerConfig)
	for k, v := range user {
		out[k] = v
	}
	for k, v := range project {
		out[k] = v
	}
	return out
}

// mergeHooks concatenates matchers from user and project configs per event.
func mergeHooks(user, project map[string][]HookMatcherConfig) map[string][]HookMatcherConfig {
	out := make(map[string][]HookMatcherConfig)
	for k, v := range user {
		out[k] = append(out[k], v...)
	}
	for k, v := range project {
		out[k] = append(out[k], v...)
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// providerEnvKey returns the provider-specific API key from env vars.
func providerEnvKey(provider string) string {
	switch strings.ToLower(provider) {
	case "codex":
		return os.Getenv("CODEX_API_KEY")
	}
	return ""
}

// providerEnvURL returns the provider-specific base URL from env vars.
func providerEnvURL(provider string) string {
	switch strings.ToLower(provider) {
	case "codex":
		return os.Getenv("CODEX_BASE_URL")
	}
	return ""
}
