// Package read implements the Read file tool.
package read

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/baidu/claude-code-go/internal/tools"
)

const defaultMaxLines = 2000

// Tool reads file contents with optional line range.
type Tool struct{}

var _ tools.Tool = (*Tool)(nil)

func (t *Tool) Name() string        { return "Read" }
func (t *Tool) IsReadOnly() bool     { return true }
func (t *Tool) IsConcurrencySafe() bool { return true }

func (t *Tool) Description() string {
	return "Read the contents of a file. Returns file content with line numbers. " +
		"Use offset and limit to read a specific range of lines."
}

func (t *Tool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"file_path": map[string]any{
				"type":        "string",
				"description": "Absolute path to the file to read.",
			},
			"offset": map[string]any{
				"type":        "integer",
				"description": "Line number to start reading from (1-indexed). Defaults to 1.",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": fmt.Sprintf("Maximum number of lines to return. Defaults to %d.", defaultMaxLines),
			},
		},
		"required": []string{"file_path"},
	}
}

func (t *Tool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	path, _ := input["file_path"].(string)
	if path == "" {
		return tools.ToolResult{IsError: true, Output: "file_path is required"}, nil
	}

	offset := 1
	if v, ok := input["offset"]; ok {
		switch n := v.(type) {
		case float64:
			offset = int(n)
		case int:
			offset = n
		}
	}
	if offset < 1 {
		offset = 1
	}

	explicitLimit := false
	limit := defaultMaxLines
	if v, ok := input["limit"]; ok {
		switch n := v.(type) {
		case float64:
			limit = int(n)
			explicitLimit = true
		case int:
			limit = n
			explicitLimit = true
		}
	}
	if limit < 1 {
		limit = defaultMaxLines
	}

	f, err := os.Open(path)
	if err != nil {
		return tools.ToolResult{IsError: true, Output: fmt.Sprintf("cannot open file: %v", err)}, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var sb strings.Builder
	lineNum := 0
	written := 0
	truncated := false

	for scanner.Scan() {
		lineNum++
		if lineNum < offset {
			continue
		}
		if written >= limit {
			truncated = true
			break
		}
		fmt.Fprintf(&sb, "%d\t%s\n", lineNum, scanner.Text())
		written++
	}

	if err := scanner.Err(); err != nil {
		return tools.ToolResult{IsError: true, Output: fmt.Sprintf("read error: %v", err)}, nil
	}

	if written == 0 {
		return tools.ToolResult{Output: "(empty file or offset beyond end of file)"}, nil
	}

	if truncated && !explicitLimit {
		fmt.Fprintf(&sb, "\n[Output truncated: showing lines %d–%d. Use offset/limit to read more.]\n",
			offset, offset+written-1)
	}

	return tools.ToolResult{Output: sb.String()}, nil
}
