// Package planmode provides tools for entering and exiting plan mode.
//
// Plan mode is used when designing implementation steps for complex tasks.
// It allows the AI agent to:
//
// 1. Thoroughly explore the codebase using Glob, Grep, and Read tools
// 2. Understand existing patterns and architecture
// 3. Design an implementation approach
// 4. Present the plan to the user for approval
//
// The two tools provided are:
//
// - EnterPlanMode: Transitions the agent into plan mode
// - ExitPlanMode: Exits plan mode and presents the plan for approval
package planmode
