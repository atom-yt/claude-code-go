package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const defaultModel = "claude-sonnet-4-6"

// CLIFlags carries values supplied on the command line.
type CLIFlags struct {
	Model   string
	APIKey  string
	Verbose bool
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
	apiKey := firstNonEmpty(flags.APIKey, os.Getenv("ANTHROPIC_API_KEY"), project.APIKey, user.APIKey)

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

	return Settings{
		Model:       model,
		APIKey:      apiKey,
		Verbose:     flags.Verbose,
		Permissions: perms,
		Hooks:       hooksCfg,
		MCPServers:  mcpServers,
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
