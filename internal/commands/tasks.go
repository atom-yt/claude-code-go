// Package commands implements slash commands.
package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/taskstore"

	// Import subagent for Context.GetSubagentRuntime() return type
	_ "github.com/atom-yt/claude-code-go/internal/subagent"
)

// TaskCmd implements /tasks command to show task list.
type TaskCmd struct{}

func (c *TaskCmd) Name() string            { return "tasks" }
func (c *TaskCmd) Aliases() []string   { return []string{"tasks"} }
func (c *TaskCmd) Description() string { return "Show current tasks and their status" }

func (c *TaskCmd) Execute(_ context.Context, _ []string, ctx *Context) (string, error) {
	var sb strings.Builder

	store := ctx.GetTaskManager()
	if store == nil {
		return "Task store not initialized", nil
	}

	// Get all non-deleted tasks
	tasks := store.List("", "")
	if len(tasks) == 0 {
		sb.WriteString("No tasks in progress.\n")
		return sb.String(), nil
	}

	// Display tasks grouped by status
	sb.WriteString("Tasks:\n")
	sb.WriteString("────────────────────────────────────────────\n")

	// Count tasks by status
	pendingCount := 0
	inProgressCount := 0
	completedCount := 0

	for _, task := range tasks {
		switch task.Status {
		case taskstore.StatusPending:
			pendingCount++
		case taskstore.StatusInProgress:
			inProgressCount++
		case taskstore.StatusCompleted:
			completedCount++
		}
	}

	// Show in_progress tasks first
	if inProgressCount > 0 {
		sb.WriteString("● In Progress:\n")
		for _, task := range tasks {
			if task.Status == taskstore.StatusInProgress {
				activeForm := task.ActiveForm
				if activeForm == "" {
					activeForm = "Working on"
				}
				sb.WriteString(fmt.Sprintf("  %s [#%s]: %s\n", activeForm, task.ID, task.Subject))
			}
		}
	}

	// Show pending tasks
	if pendingCount > 0 {
		sb.WriteString("\n○ Pending:\n")
		for _, task := range tasks {
			if task.Status == taskstore.StatusPending {
				sb.WriteString(fmt.Sprintf("  [#%s]: %s\n", task.ID, task.Subject))
			}
		}
	}

	// Show completed tasks
	if completedCount > 0 {
		sb.WriteString("\n✓ Completed:\n")
		for _, task := range tasks {
			if task.Status == taskstore.StatusCompleted {
				sb.WriteString(fmt.Sprintf("  [#%s]: %s\n", task.ID, task.Subject))
			}
		}
	}

	sb.WriteString("────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Total: %d tasks (● %d in progress, ○ %d pending, ✓ %d completed)\n",
		pendingCount+inProgressCount+completedCount,
		inProgressCount, pendingCount, completedCount))

	// Show subagent info
	subagentRuntime := ctx.GetSubagentRuntime()
	if subagentRuntime != nil {
		activeCount := subagentRuntime.GetSubagentCount()
		if activeCount > 0 {
			sb.WriteString(fmt.Sprintf("\nBackground subagents: %d active\n", activeCount))
		} else {
			sb.WriteString("\nNo background subagents running\n")
		}
	}

	return "<!-- raw -->\n" + sb.String(), nil
}
