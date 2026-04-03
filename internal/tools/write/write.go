// Package write implements the Write file tool.
package write

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

// Tool creates or overwrites files.
type Tool struct{}

var _ tools.Tool = (*Tool)(nil)

func (t *Tool) Name() string            { return "Write" }
func (t *Tool) IsReadOnly() bool         { return false }
func (t *Tool) IsConcurrencySafe() bool  { return false }

func (t *Tool) Description() string {
	return "Write content to a file, creating it (and any missing parent directories) if it doesn't exist, " +
		"or overwriting it if it does."
}

func (t *Tool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"file_path": map[string]any{
				"type":        "string",
				"description": "Absolute path of the file to write.",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Content to write to the file.",
			},
		},
		"required": []string{"file_path", "content"},
	}
}

func (t *Tool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	path, _ := input["file_path"].(string)
	content, _ := input["content"].(string)

	if path == "" {
		return tools.ToolResult{IsError: true, Output: "file_path is required"}, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return tools.ToolResult{IsError: true, Output: fmt.Sprintf("cannot create directories: %v", err)}, nil
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return tools.ToolResult{IsError: true, Output: fmt.Sprintf("cannot write file: %v", err)}, nil
	}

	return tools.ToolResult{Output: fmt.Sprintf("File written successfully: %s", path)}, nil
}
