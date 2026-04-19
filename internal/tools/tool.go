// Package tools defines the Tool interface and registry.
package tools

import "context"

// ToolResult is the output of a tool call.
type ToolResult struct {
	Output  string // text to return to Claude as tool_result content
	IsError bool   // if true, the output describes an error
}

// Tool is the interface every tool must implement.
type Tool interface {
	// Name is the unique identifier used in API tool specs and tool_use calls.
	Name() string
	// Description is shown to Claude to explain what the tool does.
	Description() string
	// InputSchema returns a JSON Schema object describing accepted parameters.
	InputSchema() map[string]any
	// Call executes the tool with parsed input and returns its result.
	Call(ctx context.Context, input map[string]any) (ToolResult, error)
	// IsReadOnly returns true if the tool never modifies files or state.
	IsReadOnly() bool
	// IsConcurrencySafe returns true if the tool can run in parallel with others.
	IsConcurrencySafe() bool
}
