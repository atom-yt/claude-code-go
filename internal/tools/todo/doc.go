// Package todo provides the TodoWrite tool for task list management.
//
// TodoWrite helps track progress on complex multi-step tasks by creating
// and managing a structured task list. Tasks can have the following statuses:
//
// - pending: Task not yet started
// - in_progress: Task currently being worked on
// - completed: Task finished
//
// Each task can have:
// - A unique ID (auto-generated if not provided)
// - A subject (brief title in imperative form)
// - A description (detailed explanation)
// - An active form (shown in spinner when in_progress)
// - An owner (optional, for sub-agent assignment)
//
// The tool maintains a global task list that persists across multiple calls,
// allowing agents to update task statuses as they progress through work.
package todo
