// Package agent implements the main AI agent loop.
package agent

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/atom-yt/claude-code-go/internal/api"
)

// EventType labels each event sent through the stream channel.
type EventType string

const (
	EventTextDelta  EventType = "text_delta"  // partial assistant text
	EventToolCall   EventType = "tool_call"   // tool about to be executed
	EventToolResult EventType = "tool_result" // tool execution finished
	EventDone       EventType = "done"        // stream complete
	EventError      EventType = "error"       // unrecoverable error
)

// StreamEvent is emitted by Agent.Query for each piece of the response.
type StreamEvent struct {
	Type        EventType
	Text        string         // EventTextDelta
	ToolName    string         // EventToolCall / EventToolResult
	ToolInput   map[string]any // EventToolCall
	ToolOutput  string         // EventToolResult
	ToolIsError bool           // EventToolResult
	Error       error          // EventError
	Usage       *api.Usage     // Token usage from API response
	NextCmd     tea.Cmd        // injected by TUI to chain channel reads
}
