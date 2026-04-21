// Package commands implements slash commands.
package commands

import (
	"context"

	"github.com/atom-yt/claude-code-go/internal/subagent"
	"github.com/atom-yt/claude-code-go/internal/taskstore"
)

// Context carries runtime objects available to commands.
type Context struct {
	// GetMessages returns the current conversation history (UI layer).
	GetMessages func() []Message
	// ClearMessages removes all messages.
	ClearMessages func()
	// AppendMessage adds a message to the UI.
	AppendMessage func(role, content string)
	// GetModel returns the current model name.
	GetModel func() string
	// SetModel switches the model.
	SetModel func(model string)
	// GetProvider returns the current provider name.
	GetProvider func() string
	// SetProvider switches the provider and rebuilds the API client.
	SetProvider func(provider string)
	// GetCost returns cumulative token counts.
	GetCost func() (inputTokens, outputTokens int)

	// GetTaskManager returns the task store for task operations.
	GetTaskManager func() *taskstore.Store
	// GetSubagentRuntime returns the subagent runtime for background tasks.
	GetSubagentRuntime func() *subagent.Runtime
	// GetTaskCount returns the count of active tasks (for status bar).
	GetTaskCount func() int

	// CompactHistory triggers history compaction.
	CompactHistory func(ctx context.Context) error
	// ConsolidateMemory triggers memory consolidation.
	ConsolidateMemory func(ctx context.Context) (string, error)
	// ConsolidateStatus returns the consolidation status.
	ConsolidateStatus func(ctx context.Context) (string, error)
	// GetConfig returns the current configuration.
	GetConfig func() map[string]any
	// RestoreSession restores a session by ID and returns the session info.
	RestoreSession func(id string) (string, error)
}

// Message is a simplified message for command use.
type Message struct {
	Role    string
	Content string
}

// Command is a slash command the user can type.
type Command interface {
	Name() string
	Aliases() []string
	Description() string
	Execute(ctx context.Context, args []string, cmdCtx *Context) (string, error)
}
