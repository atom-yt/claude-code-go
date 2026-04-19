package commands

import (
	"context"
	"fmt"
	"strings"
)

// ---- /mcp ----

type mcpCmd struct{}

func (c *mcpCmd) Name() string      { return "mcp" }
func (c *mcpCmd) Aliases() []string { return nil }
func (c *mcpCmd) Description() string {
	return "Show MCP server configuration"
}

func (c *mcpCmd) Execute(_ context.Context, _ []string, cmdCtx *Context) (string, error) {
	if cmdCtx.GetConfig == nil {
		return "Config not available", nil
	}

	cfg := cmdCtx.GetConfig()
	mcpServers, ok := cfg["mcpServers"].(map[string]any)
	if !ok || len(mcpServers) == 0 {
		return "No MCP servers configured.\n\nConfigure MCP servers in ~/.claude/settings.json or .claude/settings.json:\n\n  \"mcpServers\": {\n    \"my-server\": {\n      \"type\": \"stdio\",\n      \"command\": \"npx\",\n      \"args\": [\"-y\", \"@my/mcp-server\"]\n    }\n  }", nil
	}

	lines := []string{fmt.Sprintf("MCP servers configured (%d):", len(mcpServers))}
	lines = append(lines, "")

	for name, server := range mcpServers {
		if serverMap, ok := server.(map[string]any); ok {
			lines = append(lines, c.formatServer(name, serverMap))
		}
	}

	lines = append(lines, "")
	lines = append(lines, "MCP tools are automatically registered and available for use.")
	lines = append(lines, "Look for tools prefixed with 'mcp__' when the AI uses tools.")

	return "<!-- raw -->\n" + strings.Join(lines, "\n"), nil
}

func (c *mcpCmd) formatServer(name string, server map[string]any) string {
	lines := []string{fmt.Sprintf("%s:", name)}

	serverType, _ := server["type"].(string)
	if serverType == "" {
		serverType = "stdio"
	}
	lines = append(lines, fmt.Sprintf("  Type: %s", serverType))

	if cmd, ok := server["command"].(string); ok {
		lines = append(lines, fmt.Sprintf("  Command: %s", cmd))
	}

	if args, ok := server["args"].([]any); ok && len(args) > 0 {
		argsStr := fmt.Sprintf("%v", args)
		lines = append(lines, fmt.Sprintf("  Args: %s", argsStr))
	}

	if url, ok := server["url"].(string); ok && url != "" {
		lines = append(lines, fmt.Sprintf("  URL: %s", url))
	}

	return strings.Join(lines, "\n")
}
