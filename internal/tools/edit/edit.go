// Package edit implements the Edit file tool.
package edit

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

// Tool performs exact string replacements in files.
type Tool struct{}

var _ tools.Tool = (*Tool)(nil)

func (t *Tool) Name() string            { return "Edit" }
func (t *Tool) IsReadOnly() bool         { return false }
func (t *Tool) IsConcurrencySafe() bool  { return false }

func (t *Tool) Description() string {
	return "Replace exact occurrences of old_string with new_string in a file. " +
		"By default fails if old_string is not unique in the file. " +
		"Set replace_all=true to replace every occurrence."
}

func (t *Tool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"file_path": map[string]any{
				"type":        "string",
				"description": "Absolute path to the file to edit.",
			},
			"old_string": map[string]any{
				"type":        "string",
				"description": "The exact text to find and replace.",
			},
			"new_string": map[string]any{
				"type":        "string",
				"description": "The replacement text.",
			},
			"replace_all": map[string]any{
				"type":        "boolean",
				"description": "If true, replace all occurrences. Defaults to false.",
			},
		},
		"required": []string{"file_path", "old_string", "new_string"},
	}
}

func (t *Tool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	path, _ := input["file_path"].(string)
	oldStr, _ := input["old_string"].(string)
	newStr, _ := input["new_string"].(string)
	replaceAll, _ := input["replace_all"].(bool)

	if path == "" {
		return tools.ToolResult{IsError: true, Output: "file_path is required"}, nil
	}
	if oldStr == "" {
		return tools.ToolResult{IsError: true, Output: "old_string is required"}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return tools.ToolResult{IsError: true, Output: fmt.Sprintf("cannot read file: %v", err)}, nil
	}

	content := string(data)
	count := strings.Count(content, oldStr)

	if count == 0 {
		return tools.ToolResult{IsError: true, Output: "old_string not found in file"}, nil
	}

	if count > 1 && !replaceAll {
		return tools.ToolResult{
			IsError: true,
			Output: fmt.Sprintf("old_string appears %d times — use replace_all=true to replace all, "+
				"or provide more context to make it unique", count),
		}, nil
	}

	var result string
	if replaceAll {
		result = strings.ReplaceAll(content, oldStr, newStr)
	} else {
		result = strings.Replace(content, oldStr, newStr, 1)
	}

	if err := os.WriteFile(path, []byte(result), 0o644); err != nil {
		return tools.ToolResult{IsError: true, Output: fmt.Sprintf("cannot write file: %v", err)}, nil
	}

	replaced := 1
	if replaceAll {
		replaced = count
	}
	return tools.ToolResult{
		Output: fmt.Sprintf("Replaced %d occurrence(s) in %s", replaced, path),
	}, nil
}
