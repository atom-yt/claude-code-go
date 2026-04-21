// Package planmode implements plan mode tools (EnterPlanMode, ExitPlanMode).
// These tools now integrate with the runtime state manager for persistent plan tracking.
package planmode

import (
	"context"
	"fmt"
	"time"

	"github.com/atom-yt/claude-code-go/internal/runtime"
	"github.com/atom-yt/claude-code-go/internal/tools"
)

// EnterPlanModeTool implements a EnterPlanMode tool.
type EnterPlanModeTool struct {
	State *runtime.State
}

var _ tools.Tool = (*EnterPlanModeTool)(nil)

func (t *EnterPlanModeTool) Name() string            { return "EnterPlanMode" }
func (t *EnterPlanModeTool) IsReadOnly() bool        { return true }
func (t *EnterPlanModeTool) IsConcurrencySafe() bool { return true }

func (t *EnterPlanModeTool) Description() string {
	return "Use this tool to enter a planning mode for complex implementation tasks. " +
		"In plan mode, you'll explore the codebase using Glob, Grep, and Read tools, understand existing patterns and architecture, " +
		"design an implementation approach, and present your plan to the user for approval. " +
		"Use ExitPlanMode when your plan is complete and ready for user approval."
}

func (t *EnterPlanModeTool) InputSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *EnterPlanModeTool) Call(ctx context.Context, input map[string]any) (tools.ToolResult, error) {
	title, _ := input["title"].(string)
	if title == "" {
		return tools.ToolResult{
			Output: "Error: title parameter is required",
			IsError: true,
		}, nil
	}

	planPath, err := t.State.EnterPlanMode(ctx, title)
	if err != nil {
		return tools.ToolResult{
			Output: fmt.Sprintf("Error entering plan mode: %v", err),
			IsError: true,
		}, nil
	}

	return tools.ToolResult{
		Output: fmt.Sprintf("[%s] Plan mode entered.\n\nPlan file: %s\n\nUse ExitPlanMode when your plan is complete and ready for user approval.\n\nThe plan file path can be accessed via: cat %s\n",
			time.Now().Format("15:04:05"), planPath, planPath),
	}, nil
}

// ExitPlanModeTool implements an ExitPlanMode tool.
type ExitPlanModeTool struct {
	State *runtime.State
}

var _ tools.Tool = (*ExitPlanModeTool)(nil)

func (t *ExitPlanModeTool) Name() string            { return "ExitPlanMode" }
func (t *ExitPlanModeTool) IsReadOnly() bool        { return true }
func (t *ExitPlanModeTool) IsConcurrencySafe() bool { return true }

func (t *ExitPlanModeTool) Description() string {
	return "Use this tool when you are in plan mode and have finished writing your plan to the plan file and are ready for user approval. " +
		"This tool will mark the plan as approved so you can proceed with implementation."
}

func (t *ExitPlanModeTool) InputSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *ExitPlanModeTool) Call(ctx context.Context, input map[string]any) (tools.ToolResult, error) {
	// Mark current plan as approved
	if err := t.State.ApprovePlan(); err != nil {
		return tools.ToolResult{
			Output: fmt.Sprintf("Error approving plan: %v", err),
			IsError: true,
		}, nil
	}

	// Exit plan mode
	t.State.SetMode(runtime.ModeImplement)

	return tools.ToolResult{
		Output: fmt.Sprintf("[%s] Plan approved. You can now proceed with implementation.\n\nUse EnterPlanMode to plan more tasks.\n\nPlan file: %s\n",
			time.Now().Format("15:04:05"), t.State.PlanFilePath()),
	}, nil
}