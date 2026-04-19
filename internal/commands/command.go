// Package commands implements slash commands.
package commands

import "context"

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
	// CompactHistory triggers history compaction.
	CompactHistory func(ctx context.Context) error
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
