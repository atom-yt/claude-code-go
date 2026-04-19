package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ---- /config ----

type configCmd struct{}

func (c *configCmd) Name() string      { return "config" }
func (c *configCmd) Aliases() []string { return []string{"cfg"} }
func (c *configCmd) Description() string {
	return "Show current configuration"
}

func (c *configCmd) Execute(_ context.Context, args []string, cmdCtx *Context) (string, error) {
	if cmdCtx.GetConfig == nil {
		return "Config not available", nil
	}

	cfg := cmdCtx.GetConfig()

	var showAll bool
	if len(args) > 0 && args[0] == "all" {
		showAll = true
	}

	lines := []string{"Current configuration:"}

	// Basic settings
	model := fmt.Sprintf("%v", cfg["model"])
	provider := fmt.Sprintf("%v", cfg["provider"])
	if provider == "" {
		provider = "anthropic"
	}
	lines = append(lines, fmt.Sprintf("  Model:    %s/%s", provider, model))

	if baseURL, ok := cfg["baseURL"].(string); ok && baseURL != "" {
		lines = append(lines, fmt.Sprintf("  Base URL: %s", baseURL))
	}

	// Auto-compact settings
	if cfg["autoCompact"].(bool) {
		lines = append(lines, "")
		lines = append(lines, "Auto-compact:")
		lines = append(lines, fmt.Sprintf("  Enabled:    %v", cfg["autoCompact"]))
		lines = append(lines, fmt.Sprintf("  Threshold:  %.0f%%", cfg["compactThreshold"].(float64)*100))
		lines = append(lines, fmt.Sprintf("  Cooldown:    %d min", cfg["compactCooldown"].(int)))
		lines = append(lines, fmt.Sprintf("  Keep recent: %d messages", cfg["compactKeepRecent"].(int)))

		if cw, ok := cfg["contextWindow"].(int); ok && cw > 0 {
			lines = append(lines, fmt.Sprintf("  Context window: %d tokens", cw))
		}
	}

	// Auto-dream settings
	if cfg["autoDreamEnabled"].(bool) {
		lines = append(lines, "")
		lines = append(lines, "Auto-dream:")
		lines = append(lines, fmt.Sprintf("  Enabled:          %v", cfg["autoDreamEnabled"]))
		if md, ok := cfg["autoMemoryDirectory"].(string); ok && md != "" {
			lines = append(lines, fmt.Sprintf("  Memory directory: %s", md))
		}
		lines = append(lines, fmt.Sprintf("  Min hours:        %d", cfg["minConsolidateHours"].(int)))
		lines = append(lines, fmt.Sprintf("  Min sessions:     %d", cfg["minConsolidateSessions"].(int)))
	}

	// MCP servers
	if mcpServers, ok := cfg["mcpServers"].(map[string]any); ok && len(mcpServers) > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("MCP servers: (%d configured)", len(mcpServers)))
		for name, server := range mcpServers {
			if showAll {
				serverJSON, _ := json.MarshalIndent(server, "    ", "  ")
				lines = append(lines, fmt.Sprintf("  %s:", name))
				lines = append(lines, string(serverJSON))
			} else {
				lines = append(lines, fmt.Sprintf("  - %s", name))
			}
		}
	}

	// Permissions
	if showAll {
		if perms, ok := cfg["permissions"].(map[string]any); ok {
			lines = append(lines, "")
			lines = append(lines, "Permissions:")
			if defaultMode, ok := perms["defaultMode"].(string); ok {
				lines = append(lines, fmt.Sprintf("  Default mode: %s", defaultMode))
			}
			if allow, ok := perms["allow"].([]any); ok && len(allow) > 0 {
				lines = append(lines, fmt.Sprintf("  Allow: %d rules", len(allow)))
			}
			if deny, ok := perms["deny"].([]any); ok && len(deny) > 0 {
				lines = append(lines, fmt.Sprintf("  Deny:  %d rules", len(deny)))
			}
			if ask, ok := perms["ask"].([]any); ok && len(ask) > 0 {
				lines = append(lines, fmt.Sprintf("  Ask:   %d rules", len(ask)))
			}
		}
	}

	// Hooks
	if showAll {
		if hooks, ok := cfg["hooks"].(map[string]any); ok && len(hooks) > 0 {
			lines = append(lines, "")
			lines = append(lines, "Hooks:")
			for event, matchers := range hooks {
				if m, ok := matchers.([]any); ok {
					lines = append(lines, fmt.Sprintf("  %s: %d matchers", event, len(m)))
				}
			}
		}
	}

	if !showAll {
		lines = append(lines, "")
		lines = append(lines, "Use /config all to show full configuration including permissions and hooks.")
	}

	return "<!-- raw -->\n" + strings.Join(lines, "\n"), nil
}
