package mcp

import (
	"context"
	"fmt"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

// MCPCallTool defines the interface for MCP clients that can be used for tool registration.
type MCPCallTool interface {
	CallTool(ctx context.Context, name string, args map[string]any) (string, bool, error)
	Name() string
	TrustLevel() string
	GetTools() []ToolDef
	ListResources(ctx context.Context) ([]ResourceDef, error)
	ReadResource(ctx context.Context, uri string) (string, error)
}

// mcpTool wraps an MCP server tool as a tools.Tool.
type mcpTool struct {
	client MCPCallTool
	def    ToolDef
	name   string // prefixed: mcp__<server>__<name>
	trust  string // Trust level of the MCP server
}

var _ tools.Tool = (*mcpTool)(nil)

func (t *mcpTool) Name() string            { return t.name }
func (t *mcpTool) IsReadOnly() bool        { return false }
func (t *mcpTool) IsConcurrencySafe() bool { return false }
func (t *mcpTool) Description() string     { return t.def.Description }

// TrustLevel returns the trust level of the MCP server.
func (t *mcpTool) TrustLevel() string { return t.trust }

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
func RegisterTools(registry *tools.Registry, client MCPCallTool) {
	for _, def := range client.GetTools() {
		prefixed := fmt.Sprintf("mcp__%s__%s", client.Name(), def.Name)
		registry.Register(&mcpTool{
			client: client,
			def:    def,
			name:   prefixed,
			trust:  client.TrustLevel(),
		})
	}
}
