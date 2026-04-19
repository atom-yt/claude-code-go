// Package task implements Task tools for managing background tasks and sub-agents.
package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

// TaskCreateTool creates a new task in the task list.
type TaskCreateTool struct{}

var _ tools.Tool = (*TaskCreateTool)(nil)

func (t *TaskCreateTool) Name() string            { return "TaskCreate" }
func (t *TaskCreateTool) IsReadOnly() bool        { return false }
func (t *TaskCreateTool) IsConcurrencySafe() bool { return false }

func (t *TaskCreateTool) Description() string {
	return "Use this tool proactively in these scenarios: " +
		"Complex multi-step tasks, Non-trivial and complex tasks, Plan mode, " +
		"User explicitly requests todo list, User provides multiple tasks. " +
		"This tool helps track progress, organize complex tasks, and demonstrate thoroughness."
}

func (t *TaskCreateTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"subject": map[string]any{
				"type":        "string",
				"description": "A brief, actionable title in imperative form (e.g., 'Fix authentication bug in login flow')",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Detailed description of what needs to be done, including context and acceptance criteria",
			},
			"activeForm": map[string]any{
				"type":        "string",
				"description": "Present continuous form shown in spinner when in_progress (e.g., 'Fixing authentication bug')",
			},
			"blocks": map[string]any{
				"type":        "array",
				"description": "Task IDs that cannot start until this one completes",
				"items":       map[string]any{"type": "string"},
			},
			"blockedBy": map[string]any{
				"type":        "array",
				"description": "Task IDs that must complete before this one can start",
				"items":       map[string]any{"type": "string"},
			},
		},
		"required": []string{"subject", "description"},
	}
}

func (t *TaskCreateTool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	subject, _ := input["subject"].(string)
	description, _ := input["description"].(string)
	activeForm, _ := input["activeForm"].(string)

	if subject == "" {
		return tools.ToolResult{Output: "Error: subject is required", IsError: true}, nil
	}
	if description == "" {
		return tools.ToolResult{Output: "Error: description is required", IsError: true}, nil
	}

	task := GetManager().Create(subject, description, activeForm)

	// Handle blocks/blockedBy
	if blocks, ok := input["blocks"].([]any); ok {
		for _, b := range blocks {
			if id, ok := b.(string); ok {
				GetManager().AddBlock(task.ID, id)
			}
		}
	}

	if blockedBy, ok := input["blockedBy"].([]any); ok {
		for _, b := range blockedBy {
			if id, ok := b.(string); ok {
				GetManager().AddBlockedBy(task.ID, id)
			}
		}
	}

	return tools.ToolResult{
		Output: fmt.Sprintf("Task #%s created: %s\n\n%s", task.ID, task.Subject, task.Description),
	}, nil
}

// TaskGetTool retrieves a task by ID.
type TaskGetTool struct{}

var _ tools.Tool = (*TaskGetTool)(nil)

func (t *TaskGetTool) Name() string            { return "TaskGet" }
func (t *TaskGetTool) IsReadOnly() bool        { return true }
func (t *TaskGetTool) IsConcurrencySafe() bool { return true }

func (t *TaskGetTool) Description() string {
	return "Retrieve the full details of a task by its ID, including dependencies (blocks/blockedBy)."
}

func (t *TaskGetTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"taskId": map[string]any{
				"type":        "string",
				"description": "The ID of the task to retrieve",
			},
		},
		"required": []string{"taskId"},
	}
}

func (t *TaskGetTool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	taskID, _ := input["taskId"].(string)
	if taskID == "" {
		return tools.ToolResult{Output: "Error: taskId is required", IsError: true}, nil
	}

	task, ok := GetManager().Get(taskID)
	if !ok {
		return tools.ToolResult{Output: fmt.Sprintf("Task not found: %s", taskID), IsError: true}, nil
	}

	return tools.ToolResult{Output: formatTaskDetail(task)}, nil
}

// TaskListTool lists all tasks.
type TaskListTool struct{}

var _ tools.Tool = (*TaskListTool)(nil)

func (t *TaskListTool) Name() string            { return "TaskList" }
func (t *TaskListTool) IsReadOnly() bool        { return true }
func (t *TaskListTool) IsConcurrencySafe() bool { return true }

func (t *TaskListTool) Description() string {
	return "List all tasks in the task list with their status, showing only: id, subject, status, owner, and blockedBy. Use TaskGet to retrieve full details including description."
}

func (t *TaskListTool) InputSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *TaskListTool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	tasks := GetManager().List()
	return tools.ToolResult{Output: formatTaskListSummary(tasks)}, nil
}

// TaskUpdateTool updates a task's status and details.
type TaskUpdateTool struct{}

var _ tools.Tool = (*TaskUpdateTool)(nil)

func (t *TaskUpdateTool) Name() string            { return "TaskUpdate" }
func (t *TaskUpdateTool) IsReadOnly() bool        { return false }
func (t *TaskUpdateTool) IsConcurrencySafe() bool { return false }

func (t *TaskUpdateTool) Description() string {
	return "Update a task in the task list. Mark tasks as resolved when done, update details when requirements change, establish dependencies with addBlocks/addBlockedBy, or change owner."
}

func (t *TaskUpdateTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"taskId": map[string]any{
				"type":        "string",
				"description": "The ID of the task to update",
			},
			"status": map[string]any{
				"type":        "string",
				"enum":        []string{"pending", "in_progress", "completed", "deleted"},
				"description": "New status for the task",
			},
			"subject": map[string]any{
				"type":        "string",
				"description": "New title for the task (imperative form)",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "New detailed description",
			},
			"activeForm": map[string]any{
				"type":        "string",
				"description": "New active form for spinner",
			},
			"owner": map[string]any{
				"type":        "string",
				"description": "New owner (agent name)",
			},
			"addBlocks": map[string]any{
				"type":        "array",
				"description": "Task IDs to add to blocks list",
				"items":       map[string]any{"type": "string"},
			},
			"addBlockedBy": map[string]any{
				"type":        "array",
				"description": "Task IDs to add to blockedBy list",
				"items":       map[string]any{"type": "string"},
			},
		},
		"required": []string{"taskId"},
	}
}

func (t *TaskUpdateTool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	taskID, _ := input["taskId"].(string)
	if taskID == "" {
		return tools.ToolResult{Output: "Error: taskId is required", IsError: true}, nil
	}

	task, err := GetManager().Update(taskID, func(task *Task) {
		if status, ok := input["status"].(string); ok {
			task.Status = TaskStatus(status)
		}
		if subject, ok := input["subject"].(string); ok {
			task.Subject = subject
		}
		if desc, ok := input["description"].(string); ok {
			task.Description = desc
		}
		if activeForm, ok := input["activeForm"].(string); ok {
			task.ActiveForm = activeForm
		}
		if owner, ok := input["owner"].(string); ok {
			task.Owner = owner
		}

		if addBlocks, ok := input["addBlocks"].([]any); ok {
			for _, b := range addBlocks {
				if id, ok := b.(string); ok {
					task.Blocks = append(task.Blocks, id)
					// Add reverse dependency (handled by AddBlock now)
					GetManager().AddBlock(taskID, id)
				}
			}
		}

		if addBlockedBy, ok := input["addBlockedBy"].([]any); ok {
			for _, b := range addBlockedBy {
				if id, ok := b.(string); ok {
					task.BlockedBy = append(task.BlockedBy, id)
					// Add reverse dependency (handled by AddBlockedBy now)
					GetManager().AddBlockedBy(taskID, id)
				}
			}
		}
	})

	if err != nil {
		return tools.ToolResult{Output: err.Error(), IsError: true}, nil
	}

	return tools.ToolResult{
		Output: fmt.Sprintf("Task #%s updated to %s: %s", task.ID, task.Status, task.Subject),
	}, nil
}

// TaskDeleteTool deletes a task from the list.
type TaskDeleteTool struct{}

var _ tools.Tool = (*TaskDeleteTool)(nil)

func (t *TaskDeleteTool) Name() string            { return "TaskDelete" }
func (t *TaskDeleteTool) IsReadOnly() bool        { return false }
func (t *TaskDeleteTool) IsConcurrencySafe() bool { return false }

func (t *TaskDeleteTool) Description() string {
	return "Delete a task from the task list. This is a permanent action."
}

func (t *TaskDeleteTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"taskId": map[string]any{
				"type":        "string",
				"description": "The ID of the task to delete",
			},
		},
		"required": []string{"taskId"},
	}
}

func (t *TaskDeleteTool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	taskID, _ := input["taskId"].(string)
	if taskID == "" {
		return tools.ToolResult{Output: "Error: taskId is required", IsError: true}, nil
	}

	if err := GetManager().Delete(taskID); err != nil {
		return tools.ToolResult{Output: err.Error(), IsError: true}, nil
	}

	return tools.ToolResult{Output: fmt.Sprintf("Task #%s deleted", taskID)}, nil
}

// TaskOutputTool retrieves the output of a completed task.
// This is a placeholder - in a full implementation, this would track
// the actual output of background tasks.
type TaskOutputTool struct{}

var _ tools.Tool = (*TaskOutputTool)(nil)

func (t *TaskOutputTool) Name() string            { return "TaskOutput" }
func (t *TaskOutputTool) IsReadOnly() bool        { return true }
func (t *TaskOutputTool) IsConcurrencySafe() bool { return true }

func (t *TaskOutputTool) Description() string {
	return "Retrieve the output of a completed task. Returns task results along with status information."
}

func (t *TaskOutputTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"taskId": map[string]any{
				"type":        "string",
				"description": "The ID of the task",
			},
			"block": map[string]any{
				"type":        "boolean",
				"description": "Whether to wait for task completion (default true)",
			},
			"timeout": map[string]any{
				"type":        "integer",
				"description": "Maximum time to wait in milliseconds (default 30000)",
			},
		},
		"required": []string{"taskId"},
	}
}

func (t *TaskOutputTool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	taskID, _ := input["taskId"].(string)
	if taskID == "" {
		return tools.ToolResult{Output: "Error: taskId is required", IsError: true}, nil
	}

	task, ok := GetManager().Get(taskID)
	if !ok {
		return tools.ToolResult{Output: fmt.Sprintf("Task not found: %s", taskID), IsError: true}, nil
	}

	// In a full implementation, this would return the actual task output
	return tools.ToolResult{
		Output: fmt.Sprintf("Task #%s: %s\nStatus: %s\n\nOutput: Task completed successfully.", task.ID, task.Subject, task.Status),
	}, nil
}

// Helper functions for formatting

func formatTaskDetail(task *Task) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Task #%s: %s\n", task.ID, task.Subject))
	sb.WriteString(fmt.Sprintf("Status: %s\n", task.Status))
	sb.WriteString(fmt.Sprintf("Owner: %s\n", task.Owner))
	sb.WriteString(fmt.Sprintf("Created: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Updated: %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05")))

	if len(task.Blocks) > 0 {
		sb.WriteString(fmt.Sprintf("Blocks: %s\n", strings.Join(task.Blocks, ", ")))
	}
	if len(task.BlockedBy) > 0 {
		sb.WriteString(fmt.Sprintf("Blocked by: %s\n", strings.Join(task.BlockedBy, ", ")))
	}

	sb.WriteString(fmt.Sprintf("\nDescription:\n%s\n", task.Description))

	if len(task.Metadata) > 0 {
		sb.WriteString("\nMetadata:\n")
		for k, v := range task.Metadata {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}

	return sb.String()
}

func formatTaskListSummary(tasks []*Task) string {
	if len(tasks) == 0 {
		return "No tasks in the list."
	}

	var sb strings.Builder
	sb.WriteString("Task List:\n")
	sb.WriteString("───────────────────────────────────────────────────────────────────────\n")

	// Order: in_progress first, then pending, then completed, then deleted
	order := []string{StatusInProgress, StatusPending, StatusCompleted, StatusDeleted}
	for _, status := range order {
		for _, task := range tasks {
			if string(task.Status) != status {
				continue
			}

			var icon string
			switch task.Status {
			case StatusPending:
				icon = "○"
			case StatusInProgress:
				icon = "◐"
			case StatusCompleted:
				icon = "●"
			case StatusDeleted:
				icon = "✕"
			}

			owner := ""
			if task.Owner != "" {
				owner = fmt.Sprintf(" (owner: %s)", task.Owner)
			}

			blockedBy := ""
			if len(task.BlockedBy) > 0 {
				blockedBy = fmt.Sprintf(" [blocked by: %s]", strings.Join(task.BlockedBy, ", "))
			}

			sb.WriteString(fmt.Sprintf("%s [#%s]%s %s%s\n", icon, task.ID, owner, task.Subject, blockedBy))
		}
	}

	sb.WriteString("───────────────────────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Total: %d tasks", len(tasks)))

	return sb.String()
}

// MarshalJSON implements json.Marshaler for Task
func (t *Task) MarshalJSON() ([]byte, error) {
	type Alias Task
	return json.Marshal((*Alias)(t))
}
