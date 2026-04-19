// Package planmode implements plan mode tools (EnterPlanMode, ExitPlanMode).
// These tools allow Claude to enter and exit a planning mode for complex implementation tasks.
package planmode

import (
	"context"
	"fmt"
	"time"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

// EnterPlanModeTool implements the EnterPlanMode tool.
// This tool should be used when planning the implementation steps of a task that requires writing code.
type EnterPlanModeTool struct{}

var _ tools.Tool = (*EnterPlanModeTool)(nil)

func (t *EnterPlanModeTool) Name() string           { return "EnterPlanMode" }
func (t *EnterPlanModeTool) IsReadOnly() bool        { return true }
func (t *EnterPlanModeTool) IsConcurrencySafe() bool { return true }

func (t *EnterPlanModeTool) Description() string {
	return "Use this tool proactively when you're about to start a non-trivial implementation task. " +
		"Getting user sign-off on your approach before writing code prevents wasted effort and ensures alignment. " +
		"This tool transitions you into plan mode where you can explore the codebase and design an implementation approach for user approval. " +
		"In plan mode, you'll thoroughly explore the codebase using Glob, Grep, and Read tools, understand existing patterns and architecture, " +
		"design an implementation approach, and present your plan to the user for approval."
}

func (t *EnterPlanModeTool) InputSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *EnterPlanModeTool) Call(ctx context.Context, input map[string]any) (tools.ToolResult, error) {
	// In a full implementation, this would transition the agent to plan mode.
	// For now, return a message indicating plan mode has been entered.
	return tools.ToolResult{
		Output: fmt.Sprintf("[%s] Plan mode entered. Use ExitPlanMode when your plan is complete and ready for user approval.\n\n",
			time.Now().Format("15:04:05")),
	}, nil
}

// ExitPlanModeTool implements the ExitPlanMode tool.
// This tool should be used when the plan is complete and ready for user approval.
type ExitPlanModeTool struct{}

var _ tools.Tool = (*ExitPlanModeTool)(nil)

func (t *ExitPlanModeTool) Name() string           { return "ExitPlanMode" }
func (t *ExitPlanModeTool) IsReadOnly() bool        { return true }
func (t *ExitPlanModeTool) IsConcurrencySafe() bool { return true }

func (t *ExitPlanModeTool) Description() string {
	return "Use this tool when you are in plan mode and have finished writing your plan to the plan file and are ready for the user to review and approve it. " +
		"This tool does NOT take the plan content as a parameter - it should have been written to the plan file specified in the plan mode system message. " +
		"This tool simply signals that you're done planning and ready for user approval."
}

func (t *ExitPlanModeTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"allowedPrompts": map[string]any{
				"type":        "array",
				"description": "Optional list of permission prompts for implementation. These describe categories of actions rather than specific commands.",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"prompt": map[string]any{
							"type":        "string",
							"description": "Semantic description of the action, e.g. \"run tests\", \"install dependencies\"",
						},
						"tool": map[string]any{
							"type":        "string",
							"description": "The tool this prompt applies to (currently only Bash is supported)",
							"enum":        []string{"Bash"},
						},
					},
				},
			},
		},
	}
}

func (t *ExitPlanModeTool) Call(ctx context.Context, input map[string]any) (tools.ToolResult, error) {
	// In a full implementation, this would read the plan file and present it to the user for approval.
	// For now, return a message indicating the plan is ready for review.
	return tools.ToolResult{
		Output: fmt.Sprintf("[%s] Plan mode exited. The plan has been presented to the user for approval.\n",
			time.Now().Format("15:04:05")),
	}, nil
}
