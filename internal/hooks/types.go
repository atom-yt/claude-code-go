// Package hooks implements the lifecycle hook system.
package hooks

// Event identifies a lifecycle point where hooks can fire.
type Event string

const (
	EventSessionStart     Event = "session_start"
	EventPreToolCall      Event = "pre_tool_call"
	EventPostToolCall     Event = "post_tool_call"
	EventUserPromptSubmit Event = "user_prompt_submit"
	EventStop             Event = "stop"
)

// CommandType identifies how a hook is executed.
type CommandType string

const (
	TypeCommand CommandType = "command" // shell command
	TypeHTTP    CommandType = "http"    // HTTP POST
)

// HookCommand is one executable hook step.
type HookCommand struct {
	Type    CommandType       `json:"type"`
	Command string            `json:"command"` // for type=command
	URL     string            `json:"url"`     // for type=http
	Headers map[string]string `json:"headers"` // for type=http
	Timeout int               `json:"timeout"` // seconds; 0 → default (30s)
}

// Matcher is one hook group: a pattern and the commands to run when matched.
type Matcher struct {
	// ToolPattern is matched against the tool name (empty = match all).
	// Supports simple glob-style wildcards via filepath.Match.
	ToolPattern string        `json:"matcher"`
	Hooks       []HookCommand `json:"hooks"`
}

// Input is the payload passed to a hook command (serialised as JSON env vars).
type Input struct {
	Event      Event          `json:"event"`
	ToolName   string         `json:"tool_name,omitempty"`
	ToolInput  map[string]any `json:"tool_input,omitempty"`
	SessionID  string         `json:"session_id,omitempty"`
	UserPrompt string         `json:"user_prompt,omitempty"` // User's input text for user_prompt_submit event
}

// Result is the outcome of running a hook command.
type Result struct {
	// Decision: "" (continue) | "deny" (block tool call)
	Decision string
	// Reason is the denial reason returned by the hook.
	Reason string
	// Output is the raw stdout/body returned by the hook.
	Output string
}
