// Package grep implements the Grep content search tool.
package grep

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

const maxMatchLines = 500

// Tool searches file content with a regular expression.
type Tool struct{}

var _ tools.Tool = (*Tool)(nil)

func (t *Tool) Name() string            { return "Grep" }
func (t *Tool) IsReadOnly() bool        { return true }
func (t *Tool) IsConcurrencySafe() bool { return true }

func (t *Tool) Description() string {
	return "Search for a regular expression pattern in files. " +
		"Returns matching lines with file name and line number. " +
		"Use include to filter by filename pattern and context for surrounding lines."
}

func (t *Tool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Regular expression to search for.",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "File or directory to search. Defaults to current working directory.",
			},
			"include": map[string]any{
				"type":        "string",
				"description": "Glob pattern to filter files (e.g. \"*.go\", \"*.{ts,tsx}\").",
			},
			"context": map[string]any{
				"type":        "integer",
				"description": "Number of context lines before and after each match.",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *Tool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	pattern, _ := input["pattern"].(string)
	if pattern == "" {
		return tools.ToolResult{IsError: true, Output: "pattern is required"}, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return tools.ToolResult{IsError: true, Output: fmt.Sprintf("invalid regex: %v", err)}, nil
	}

	searchPath, _ := input["path"].(string)
	if searchPath == "" {
		searchPath, _ = os.Getwd()
	}

	include, _ := input["include"].(string)

	contextLines := 0
	if v, ok := input["context"]; ok {
		if n, ok := v.(float64); ok {
			contextLines = int(n)
		}
	}

	var sb strings.Builder
	totalMatches := 0
	truncated := false

	err = filepath.Walk(searchPath, func(filePath string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil || fi.IsDir() {
			return nil
		}
		if include != "" {
			matched, _ := filepath.Match(include, filepath.Base(filePath))
			if !matched {
				return nil
			}
		}
		if truncated {
			return filepath.SkipAll
		}

		matches, err := searchFile(re, filePath, contextLines)
		if err != nil {
			return nil // skip unreadable files silently
		}

		for _, m := range matches {
			if totalMatches >= maxMatchLines {
				truncated = true
				return filepath.SkipAll
			}
			sb.WriteString(m)
			sb.WriteByte('\n')
			totalMatches++
		}
		return nil
	})

	if err != nil {
		return tools.ToolResult{IsError: true, Output: fmt.Sprintf("walk error: %v", err)}, nil
	}

	if totalMatches == 0 {
		return tools.ToolResult{Output: "No matches found."}, nil
	}

	if truncated {
		fmt.Fprintf(&sb, "\n[Truncated: showing first %d matches]\n", maxMatchLines)
	}

	return tools.ToolResult{Output: sb.String()}, nil
}

// searchFile returns formatted match lines (with context) from a single file.
func searchFile(re *regexp.Regexp, path string, contextLines int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var results []string
	lastPrinted := -1

	for i, line := range lines {
		if !re.MatchString(line) {
			continue
		}

		start := i - contextLines
		if start < 0 {
			start = 0
		}
		end := i + contextLines
		if end >= len(lines) {
			end = len(lines) - 1
		}

		if lastPrinted >= 0 && start > lastPrinted+1 {
			results = append(results, "--")
		}

		for j := start; j <= end; j++ {
			if j <= lastPrinted {
				continue
			}
			lineNum := j + 1
			prefix := "  "
			if j == i {
				prefix = "> "
			}
			results = append(results, fmt.Sprintf("%s%s:%d: %s", prefix, path, lineNum, lines[j]))
			lastPrinted = j
		}
	}

	return results, nil
}
