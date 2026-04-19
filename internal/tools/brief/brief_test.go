package brief

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTool_Name(t *testing.T) {
	tool := &Tool{}
	assert.Equal(t, "Brief", tool.Name())
	assert.True(t, tool.IsReadOnly())
	assert.True(t, tool.IsConcurrencySafe())
}

func TestTool_InputSchema(t *testing.T) {
	tool := &Tool{}
	schema := tool.InputSchema()

	assert.Equal(t, "object", schema["type"])
	props := schema["properties"].(map[string]any)
	assert.Contains(t, props, "paths")
	assert.Contains(t, props, "max_lines_per_file")
	assert.Contains(t, props, "include_summary")
}

func TestTool_Call_MissingPaths(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{})

	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "paths is required")
}

func TestTool_Call_EmptyPaths(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"paths": []any{},
	})

	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "non-empty array")
}

func TestTool_Call_ValidFiles(t *testing.T) {
	// Create temporary files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test1.txt")
	file2 := filepath.Join(tmpDir, "test2.txt")

	os.WriteFile(file1, []byte("Hello\nWorld\n"), 0644)
	os.WriteFile(file2, []byte("Another file\nWith content\n"), 0644)

	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"paths": []any{file1, file2},
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Brief: 2 file(s)")
	assert.Contains(t, result.Output, file1)
	assert.Contains(t, result.Output, file2)
}

func TestTool_Call_NonExistentFile(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"paths": []any{"/nonexistent/file.txt"},
	})

	assert.NoError(t, err)
	assert.Contains(t, result.Output, "Error:")
}