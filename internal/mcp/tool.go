package mcp

import (
	"context"
	"fmt"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

// mcpTool wraps an MCP server tool as a tools.Tool.
type mcpTool struct {
	client *Client
	def    ToolDef
	name   string // prefixed: mcp__<server>__<name>
}

var _ tools.Tool = (*mcpTool)(nil)

func (t *mcpTool) Name() string        { return t.name }
func (t *mcpTool) IsReadOnly() bool     { return false }
func (t *mcpTool) IsConcurrencySafe() bool { return false }
func (t *mcpTool) Description() string  { return t.def.Description }

func (t *mcpTool) InputSchema() map[string]any {
	if t.def.InputSchema != nil {
		return t.def.InputSchema
	}
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func (t *mcpTool) Call(ctx context.Context, input map[string]any) (tools.ToolResult, error) {
	output, isErr, err := t.client.CallTool(ctx, t.def.Name, input)
	if err != nil {
		return tools.ToolResult{IsError: true, Output: fmt.Sprintf("MCP error: %v", err)}, nil
	}
	return tools.ToolResult{Output: output, IsError: isErr}, nil
}

// RegisterTools adds all tools from the MCP client into the registry,
// prefixed with "mcp__<serverName>__".
func RegisterTools(registry *tools.Registry, client *Client) {
	for _, def := range client.Tools {
		prefixed := fmt.Sprintf("mcp__%s__%s", client.name, def.Name)
		registry.Register(&mcpTool{
			client: client,
			def:    def,
			name:   prefixed,
		})
	}
}
