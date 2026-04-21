// Package mcpresource implements MCP resource access tools.
package mcpresource

import (
	"context"
	"fmt"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/mcp"
	"github.com/atom-yt/claude-code-go/internal/tools"
)

// ListMcpResourcesTool implements ListMcpResources tool.
type ListMcpResourcesTool struct {
	clients *map[string]mcp.MCPCallTool
}

var _ tools.Tool = (*ListMcpResourcesTool)(nil)

func (t *ListMcpResourcesTool) Name() string            { return "ListMcpResources" }
func (t *ListMcpResourcesTool) IsReadOnly() bool        { return true }
func (t *ListMcpResourcesTool) IsConcurrencySafe() bool { return true }

func (t *ListMcpResourcesTool) Description() string {
	return "List all MCP resources available from connected servers"
}

func (t *ListMcpResourcesTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{},
	}
}

func (t *ListMcpResourcesTool) Call(ctx context.Context, _ map[string]any) (tools.ToolResult, error) {
	if t.clients == nil {
		return tools.ToolResult{
			Output:  "No MCP servers connected",
			IsError: true,
		}, nil
	}

	var sb strings.Builder
	sb.WriteString("MCP Resources:\n")
	sb.WriteString("────────────────────────────────────────────\n")

	// Collect resources from all servers
	var allResources []mcp.ResourceDef
	for name, client := range *t.clients {
		resources, err := client.ListResources(ctx)
		if err != nil {
			sb.WriteString(fmt.Sprintf("[%s] Error: %v\n", name, err))
			continue
		}

		if len(resources) == 0 {
			sb.WriteString(fmt.Sprintf("[%s] No resources available\n", name))
			continue
		}

		sb.WriteString(fmt.Sprintf("\n--- %s (%s, %s) ---\n", name, client.Name(), client.TrustLevel()))

		for _, res := range resources {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", res.Name, res.URI))
			if res.Description != "" {
				sb.WriteString(fmt.Sprintf("    %s\n", res.Description))
			}
		}

		allResources = append(allResources, resources...)
	}

	if len(allResources) == 0 {
		sb.WriteString("No MCP resources available from any server.\n")
	}

	sb.WriteString("────────────────────────────────────────────\n")

	return tools.ToolResult{
		Output: sb.String(),
		IsError: false,
	}, nil
}

// ReadMcpResourceTool implements ReadMcpResource tool.
type ReadMcpResourceTool struct {
	clients *map[string]mcp.MCPCallTool
}

var _ tools.Tool = (*ReadMcpResourceTool)(nil)

func (t *ReadMcpResourceTool) Name() string            { return "ReadMcpResource" }
func (t *ReadMcpResourceTool) IsReadOnly() bool        { return true }
func (t *ReadMcpResourceTool) IsConcurrencySafe() bool { return true }

func (t *ReadMcpResourceTool) Description() string {
	return "Read content of an MCP resource by URI"
}

func (t *ReadMcpResourceTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"uri": map[string]any{
				"type":        "string",
				"description": "URI of the resource to read (e.g., file://path or mcp://server/resource)",
			},
		},
		"required": []string{"uri"},
	}
}

func (t *ReadMcpResourceTool) Call(ctx context.Context, input map[string]any) (tools.ToolResult, error) {
	uri, ok := input["uri"].(string)
	if !ok || uri == "" {
		return tools.ToolResult{
			Output:  "Error: uri parameter is required",
			IsError: true,
		}, nil
	}

	if t.clients == nil {
		return tools.ToolResult{
			Output:  "No MCP servers connected",
			IsError: true,
		}, nil
	}

	// Parse URI to find server prefix: mcp://<server>/<resource>
	serverName := ""
	if len(uri) > 6 && uri[:6] == "mcp://" {
		// Extract server name from URI
		remaining := uri[6:]
		if idx := findChar(remaining, '/'); idx > 0 {
			serverName = remaining[:idx]
		}
	}

	client, hasClient := (*t.clients)[serverName]
	if !hasClient {
		return tools.ToolResult{
			Output: fmt.Sprintf("Error: MCP server '%s' not connected or not found in URI: %s", serverName, uri),
			IsError: true,
		}, nil
	}

	content, err := client.ReadResource(ctx, uri)
	if err != nil {
		return tools.ToolResult{
			Output: fmt.Sprintf("Error reading resource: %v", err),
			IsError: true,
		}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Resource: %s\n", uri))
	sb.WriteString(fmt.Sprintf("Server: %s (%s, %s)\n", serverName, client.Name(), client.TrustLevel()))
	sb.WriteString("────────────────────────────────────────────\n")
	sb.WriteString(content)
	sb.WriteString("────────────────────────────────────────────\n")

	return tools.ToolResult{
		Output: sb.String(),
		IsError: false,
	}, nil
}

func findChar(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// SetClients sets the MCP clients map for tool access.
func (t *ListMcpResourcesTool) SetClients(clients *map[string]mcp.MCPCallTool) {
	t.clients = clients
}

func (t *ReadMcpResourceTool) SetClients(clients *map[string]mcp.MCPCallTool) {
	t.clients = clients
}
