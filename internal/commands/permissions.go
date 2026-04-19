package commands

import (
	"context"
	"fmt"
	"strings"
)

// ---- /permissions ----

type permissionsCmd struct{}

func (c *permissionsCmd) Name() string      { return "permissions" }
func (c *permissionsCmd) Aliases() []string { return []string{"perms"} }
func (c *permissionsCmd) Description() string {
	return "Show current permission settings"
}

func (c *permissionsCmd) Execute(_ context.Context, args []string, cmdCtx *Context) (string, error) {
	if cmdCtx.GetConfig == nil {
		return "Config not available", nil
	}

	cfg := cmdCtx.GetConfig()
	perms, ok := cfg["permissions"].(map[string]any)
	if !ok {
		return "No permission configuration found", nil
	}

	lines := []string{"Permission settings:"}

	// Default mode
	defaultMode, _ := perms["defaultMode"].(string)
	if defaultMode == "" {
		defaultMode = "default"
	}
	modeDesc := map[string]string{
		"default":  "Read-only tools auto-allowed, mutating tools require ask",
		"manual":   "All tools require ask",
		"trust-all": "Skip all permission checks",
	}
	lines = append(lines, fmt.Sprintf("  Default mode: %s", defaultMode))
	if desc, ok := modeDesc[defaultMode]; ok {
		lines = append(lines, fmt.Sprintf("    %s", desc))
	}

	// Deny rules (highest priority)
	var denyRules, allowRules, askRules []any
	if deny, ok := perms["deny"].([]any); ok {
		denyRules = deny
		if len(deny) > 0 {
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("Deny rules (%d) - highest priority:", len(deny)))
			for i, r := range deny {
				if rule, ok := r.(map[string]any); ok {
					lines = append(lines, c.formatRule(i+1, rule))
				}
			}
		}
	}

	// Allow rules
	if allow, ok := perms["allow"].([]any); ok {
		allowRules = allow
		if len(allow) > 0 {
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("Allow rules (%d):", len(allow)))
			for i, r := range allow {
				if rule, ok := r.(map[string]any); ok {
					lines = append(lines, c.formatRule(i+1, rule))
				}
			}
		}
	}

	// Ask rules
	if ask, ok := perms["ask"].([]any); ok {
		askRules = ask
		if len(ask) > 0 {
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("Ask rules (%d):", len(ask)))
			for i, r := range ask {
				if rule, ok := r.(map[string]any); ok {
					lines = append(lines, c.formatRule(i+1, rule))
				}
			}
		}
	}

	if len(denyRules) == 0 && len(allowRules) == 0 && len(askRules) == 0 {
		lines = append(lines, "")
		lines = append(lines, "  No custom rules configured.")
	}

	// Built-in read-only tools
	lines = append(lines, "")
	lines = append(lines, "Read-only tools (auto-allowed in default mode):")
	lines = append(lines, "  Read, Glob, Grep, GrepFileContent")

	return "<!-- raw -->\n" + strings.Join(lines, "\n"), nil
}

func (c *permissionsCmd) formatRule(index int, rule map[string]any) string {
	parts := []string{fmt.Sprintf("%d.", index)}

	if tool, ok := rule["tool"].(string); ok && tool != "" {
		parts = append(parts, fmt.Sprintf("tool: %s", tool))
	}
	if path, ok := rule["path"].(string); ok && path != "" {
		parts = append(parts, fmt.Sprintf("path: %s", path))
	}
	if command, ok := rule["command"].(string); ok && command != "" {
		parts = append(parts, fmt.Sprintf("command: %s", command))
	}

	if len(parts) == 1 {
		return fmt.Sprintf("  %d. <empty rule>", index)
	}

	return "    " + strings.Join(parts, ", ")
}
