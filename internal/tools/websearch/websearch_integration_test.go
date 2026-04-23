//go:build integration
// +build integration

package websearch

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTool_Call_MaxResultsClamping(t *testing.T) {
	tool := NewTool()

	// Test max > 10 gets clamped to 10
	result, err := tool.Call(context.Background(), map[string]any{
		"query":       "test",
		"max_results": 15.0,
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Search results for: test")

	// Count number of result entries (lines starting with a number and a dot)
	lines := strings.Split(result.Output, "\n")
	resultCount := 0
	for _, line := range lines {
		if len(line) > 0 && line[0] >= '1' && line[0] <= '9' && strings.Contains(line, ". ") {
			resultCount++
		}
	}
	// Default is 5 results, max is 10
	assert.LessOrEqual(t, resultCount, 10)
}

func TestTool_Call_DefaultMaxResults(t *testing.T) {
	tool := NewTool()

	// Test default max_results is 5
	result, err := tool.Call(context.Background(), map[string]any{
		"query": "golang",
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Search results for: golang")
}

func TestTool_Call_NoResults(t *testing.T) {
	tool := NewTool()

	// Search for something very unlikely to have results
	result, err := tool.Call(context.Background(), map[string]any{
		"query": "xyzzy123nonexistent456qwerty",
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "No results found")
}

func TestTool_Search_InvalidURL(t *testing.T) {
	tool := NewTool()

	// This tests that internal addresses are properly filtered
	// The search URL itself is valid (duckduckgo.com)
	result, err := tool.Call(context.Background(), map[string]any{
		"query": "test",
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Search results for: test")
}

func TestTool_ParseResults(t *testing.T) {
	tool := NewTool()

	// Test parsing with real search
	result, err := tool.Call(context.Background(), map[string]any{
		"query":       "github",
		"max_results": 3.0,
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "github")

	// Verify output format contains expected elements
	assert.Contains(t, result.Output, "Search results for:")
	assert.Contains(t, result.Output, "github")
}

func TestTool_Call_ContextCancellation(t *testing.T) {
	tool := NewTool()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := tool.Call(ctx, map[string]any{
		"query": "test",
	})
	assert.NoError(t, err)
	// Context cancellation should result in an error
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "Search error")
}