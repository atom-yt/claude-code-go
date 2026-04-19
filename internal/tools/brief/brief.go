// Package brief implements the Brief tool for batch file upload and summarization.
package brief

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

// Tool implements the Brief tool.
type Tool struct{}

var _ tools.Tool = (*Tool)(nil)

func (t *Tool) Name() string            { return "Brief" }
func (t *Tool) IsReadOnly() bool        { return true }
func (t *Tool) IsConcurrencySafe() bool { return true }

func (t *Tool) Description() string {
	return "Upload and summarize multiple files at once. " +
		"Useful for sharing context about a project or codebase."
}

func (t *Tool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"paths": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "List of file paths to upload and summarize.",
			},
			"max_lines_per_file": map[string]any{
				"type":        "integer",
				"description": "Maximum lines to read from each file. Default: 100",
			},
			"include_summary": map[string]any{
				"type":        "boolean",
				"description": "Whether to include a summary of each file. Default: true",
			},
		},
		"required": []string{"paths"},
	}
}

func (t *Tool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	paths, _ := input["paths"].([]any)
	if len(paths) == 0 {
		return tools.ToolResult{IsError: true, Output: "paths is required and must be a non-empty array"}, nil
	}

	maxLines := 100
	if v, ok := input["max_lines_per_file"].(float64); ok {
		maxLines = int(v)
	}

	includeSummary := true
	if v, ok := input["include_summary"].(bool); ok {
		includeSummary = v
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Brief: %d file(s)\n", len(paths)))
	sb.WriteString(strings.Repeat("-", 60))
	sb.WriteString("\n\n")

	for i, p := range paths {
		filePath, _ := p.(string)
		if filePath == "" {
			continue
		}

		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, filePath))

		content, err := readFile(filePath, maxLines)
		if err != nil {
			sb.WriteString(fmt.Sprintf("   Error: %v\n\n", err))
			continue
		}

		if includeSummary {
			summary := summarizeContent(content)
			sb.WriteString(fmt.Sprintf("   Summary: %s\n", summary))
		}

		lines := strings.Count(content, "\n") + 1
		sb.WriteString(fmt.Sprintf("   Lines: %d\n", lines))
		sb.WriteString(fmt.Sprintf("   Size: %d bytes\n\n", len(content)))

		// Include first few lines as preview
		previewLines := 5
		preview := getPreview(content, previewLines)
		sb.WriteString("   Preview:\n")
		for _, line := range strings.Split(preview, "\n") {
			if line != "" {
				sb.WriteString(fmt.Sprintf("   > %s\n", line))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("-", 60))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Total files: %d\n", len(paths)))

	return tools.ToolResult{Output: sb.String()}, nil
}

// readFile reads a file and returns its content, limited to maxLines.
func readFile(path string, maxLines int) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content := string(data)
	if maxLines > 0 {
		lines := strings.Split(content, "\n")
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			content = strings.Join(lines, "\n") + "\n... (truncated)"
		}
	}

	return content, nil
}

// summarizeContent creates a brief summary of the content.
func summarizeContent(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return "empty"
	}

	// Count non-empty lines
	nonEmpty := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
	}

	// Detect programming language based on file extension
	// (simplified detection)
	language := "text"

	// Estimate complexity based on unique lines
	uniqueLines := make(map[string]bool)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			uniqueLines[trimmed] = true
		}
	}

	complexity := "simple"
	if len(uniqueLines) > 50 {
		complexity = "moderate"
	}
	if len(uniqueLines) > 100 {
		complexity = "complex"
	}

	return fmt.Sprintf("%s, %s (%d unique lines)", language, complexity, len(uniqueLines))
}

// getPreview returns the first n lines of content.
func getPreview(content string, n int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= n {
		return content
	}
	return strings.Join(lines[:n], "\n")
}
