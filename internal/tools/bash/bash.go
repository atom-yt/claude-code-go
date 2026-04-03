// Package bash implements the Bash shell execution tool.
package bash

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

const (
	defaultTimeout = 120 * time.Second
	maxTimeout     = 600 * time.Second
)

// Tool executes shell commands via bash.
type Tool struct{}

var _ tools.Tool = (*Tool)(nil)

func (t *Tool) Name() string            { return "Bash" }
func (t *Tool) IsReadOnly() bool         { return false }
func (t *Tool) IsConcurrencySafe() bool  { return false }

func (t *Tool) Description() string {
	return "Execute a shell command using bash. Returns combined stdout and stderr. " +
		"Commands time out after 120 seconds by default (max 600s)."
}

func (t *Tool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The shell command to execute.",
			},
			"timeout": map[string]any{
				"type":        "integer",
				"description": "Timeout in seconds (max 600). Defaults to 120.",
			},
		},
		"required": []string{"command"},
	}
}

func (t *Tool) Call(ctx context.Context, input map[string]any) (tools.ToolResult, error) {
	command, _ := input["command"].(string)
	if command == "" {
		return tools.ToolResult{IsError: true, Output: "command is required"}, nil
	}

	timeout := defaultTimeout
	if v, ok := input["timeout"]; ok {
		switch n := v.(type) {
		case float64:
			d := time.Duration(n) * time.Second
			if d > maxTimeout {
				d = maxTimeout
			}
			if d > 0 {
				timeout = d
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()

	output := buf.String()
	if ctx.Err() == context.DeadlineExceeded {
		return tools.ToolResult{
			IsError: true,
			Output:  output + fmt.Sprintf("\n[Command timed out after %.0fs]", timeout.Seconds()),
		}, nil
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			output += fmt.Sprintf("\nExit code: %d", exitErr.ExitCode())
			return tools.ToolResult{IsError: true, Output: output}, nil
		}
		return tools.ToolResult{IsError: true, Output: fmt.Sprintf("exec error: %v\n%s", err, output)}, nil
	}

	return tools.ToolResult{Output: output}, nil
}
