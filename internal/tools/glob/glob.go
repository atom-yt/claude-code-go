// Package glob implements the Glob file search tool.
package glob

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

const maxResults = 1000

// Tool matches files by glob pattern.
type Tool struct{}

var _ tools.Tool = (*Tool)(nil)

func (t *Tool) Name() string            { return "Glob" }
func (t *Tool) IsReadOnly() bool         { return true }
func (t *Tool) IsConcurrencySafe() bool  { return true }

func (t *Tool) Description() string {
	return "Find files matching a glob pattern (supports ** for recursive matching). " +
		"Returns matching paths sorted by modification time (most recent first)."
}

func (t *Tool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Glob pattern, e.g. \"**/*.go\" or \"src/**/*.ts\".",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Directory to search in. Defaults to current working directory.",
			},
		},
		"required": []string{"pattern"},
	}
}

type fileInfo struct {
	path    string
	modTime time.Time
}

func (t *Tool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	pattern, _ := input["pattern"].(string)
	if pattern == "" {
		return tools.ToolResult{IsError: true, Output: "pattern is required"}, nil
	}

	root, _ := input["path"].(string)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return tools.ToolResult{IsError: true, Output: fmt.Sprintf("cannot get working directory: %v", err)}, nil
		}
	}

	fsys := os.DirFS(root)
	matches, err := doublestar.Glob(fsys, pattern)
	if err != nil {
		return tools.ToolResult{IsError: true, Output: fmt.Sprintf("glob error: %v", err)}, nil
	}

	if len(matches) == 0 {
		return tools.ToolResult{Output: "No files matched the pattern."}, nil
	}

	// Stat each match for sorting by modification time.
	infos := make([]fileInfo, 0, len(matches))
	for _, m := range matches {
		fullPath := root + "/" + m
		fi, err := os.Stat(fullPath)
		if err != nil || fi.IsDir() {
			continue
		}
		infos = append(infos, fileInfo{path: fullPath, modTime: fi.ModTime()})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].modTime.After(infos[j].modTime)
	})

	truncated := false
	if len(infos) > maxResults {
		infos = infos[:maxResults]
		truncated = true
	}

	var sb strings.Builder
	for _, fi := range infos {
		sb.WriteString(fi.path)
		sb.WriteByte('\n')
	}
	if truncated {
		fmt.Fprintf(&sb, "\n[Showing first %d results. Refine your pattern to narrow results.]\n", maxResults)
	}

	return tools.ToolResult{Output: sb.String()}, nil
}
