// Package websearch implements a web search tool using DuckDuckGo HTML.
package websearch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTool_Name(t *testing.T) {
	tool := NewTool()
	assert.Equal(t, "WebSearch", tool.Name())
}

func TestTool_IsReadOnly(t *testing.T) {
	tool := NewTool()
	assert.True(t, tool.IsReadOnly())
}

func TestTool_IsConcurrencySafe(t *testing.T) {
	tool := NewTool()
	assert.True(t, tool.IsConcurrencySafe())
}

func TestTool_Description(t *testing.T) {
	tool := NewTool()
	desc := tool.Description()
	assert.Contains(t, desc, "web")
	assert.Contains(t, desc, "search")
	assert.Contains(t, desc, "real-time")
}

func TestTool_InputSchema(t *testing.T) {
	tool := NewTool()
	schema := tool.InputSchema()
	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema["type"])

	props := schema["properties"].(map[string]any)
	assert.Contains(t, props, "query")
	assert.Contains(t, props, "max_results")

	required := schema["required"].([]string)
	assert.Contains(t, required, "query")
}

func TestTool_Call_MissingQuery(t *testing.T) {
	tool := NewTool()
	result, err := tool.Call(context.Background(), map[string]any{})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "query parameter is required")
}

func TestTool_Call_EmptyQuery(t *testing.T) {
	tool := NewTool()
	result, err := tool.Call(context.Background(), map[string]any{"query": ""})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "query parameter is required")
}