package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/runtime"
)

// ---- /plan ----

type planCmd struct{}

func (c *planCmd) Name() string      { return "plan" }
func (c *planCmd) Aliases() []string { return nil }
func (c *planCmd) Description() string {
	return "Show current plan mode status or plan file content"
}

func (c *planCmd) Execute(_ context.Context, args []string, ctx *Context) (string, error) {
	runtimeState := ctx.GetRuntimeState()
	if runtimeState == nil {
		return "Runtime state not available", nil
	}

	var sb strings.Builder

	// Show plan mode status
	if runtimeState.IsPlanMode() {
		sb.WriteString("PLAN MODE ACTIVE\n")
		sb.WriteString("────────────────────────────────────────────\n")
		sb.WriteString("Mode: Planning\n")
		sb.WriteString("Tools: Read-only exploration tools enabled\n")

		// Show plan file path
		planPath := runtimeState.PlanFilePath()
		if planPath != "" {
			sb.WriteString(fmt.Sprintf("Plan file: %s\n", planPath))

			// Try to read plan file content
			if content, err := os.ReadFile(planPath); err == nil {
				sb.WriteString("\n────────────────────────────────────────────\n")
				sb.WriteString("Plan Content:\n")
				sb.WriteString("────────────────────────────────────────────\n")
				sb.WriteString(string(content))
			}
		}
	} else if runtimeState.Mode() == runtime.ModeImplement {
		sb.WriteString("IMPLEMENT MODE ACTIVE\n")
		sb.WriteString("────────────────────────────────────────────\n")
		sb.WriteString("Mode: Executing plan\n")
		sb.WriteString("Tools: All tools enabled for implementation\n")

		// Show plan file path
		planPath := runtimeState.PlanFilePath()
		if planPath != "" {
			sb.WriteString(fmt.Sprintf("Plan file: %s\n", planPath))

			// Try to read plan file content
			if content, err := os.ReadFile(planPath); err == nil {
				sb.WriteString("\n────────────────────────────────────────────\n")
				sb.WriteString("Plan Content:\n")
				sb.WriteString("────────────────────────────────────────────\n")
				sb.WriteString(string(content))
			}
		}
	} else {
		sb.WriteString("No active plan\n")
		sb.WriteString("────────────────────────────────────────────\n")
		sb.WriteString("Plan mode is not active.\n")
		sb.WriteString("Use 'plan' tool to create a new plan.\n")

		// Check if plan file exists
		cwd, err := os.Getwd()
		if err == nil {
			planPath := filepath.Join(cwd, ".claude", "plan.md")
			if _, err := os.Stat(planPath); err == nil {
				sb.WriteString(fmt.Sprintf("\nExisting plan file: %s\n", planPath))
			}
		}
	}

	// Show plan steps if available
	currentPlan := runtimeState.CurrentPlan()
	if currentPlan != nil && len(currentPlan.Steps) > 0 {
		sb.WriteString("\n────────────────────────────────────────────\n")
		sb.WriteString("Plan Steps:\n")
		sb.WriteString("────────────────────────────────────────────\n")
		for i, step := range currentPlan.Steps {
			status := "○"
			if step.Status == "completed" {
				status = "✓"
			} else if step.Status == "in_progress" {
				status = "●"
			}
			sb.WriteString(fmt.Sprintf("%s %d. %s\n", status, i+1, step.Title))
		}
	}

	return "<!-- raw -->\n" + sb.String(), nil
}