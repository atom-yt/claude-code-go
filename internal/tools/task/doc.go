// Package task provides tools for managing background tasks and sub-agents.
//
// The Task tools enable agents to:
//
// 1. Track progress on complex multi-step tasks
// 2. Organize work into a structured task list
// 3. Manage task dependencies (blocks/blockedBy)
// 4. Demonstrate thoroughness to the user
//
// Available tools:
//
// - TaskCreate: Create a new task
// - TaskGet: Retrieve task details by ID
// - TaskList: List all tasks with summary info
// - TaskUpdate: Update task status and details
// - TaskDelete: Delete a task
// - TaskOutput: Retrieve task output
//
// Task statuses:
//
// - pending: Task not yet started
// - in_progress: Task currently being worked on
// - completed: Task finished
// - deleted: Task has been removed
//
// Tasks can have dependencies:
//
// - blocks: Task IDs that cannot start until this one completes
// - blockedBy: Task IDs that must complete before this one can start
package task
