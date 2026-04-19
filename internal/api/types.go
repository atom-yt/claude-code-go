// Package api provides the Anthropic Messages API client.
package api

// Role is the message sender role.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// ContentBlock represents a single piece of message content.
type ContentBlock struct {
	Type      string         `json:"type"`                // "text" | "tool_use" | "tool_result"
	Text      string         `json:"text,omitempty"`
	// tool_use fields
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Input     map[string]any `json:"input,omitempty"`
	// tool_result fields
	ToolUseID string         `json:"tool_use_id,omitempty"`
	Content   string         `json:"content,omitempty"`
	IsError   bool           `json:"is_error,omitempty"`
}

// Message is a single turn in a conversation.
type Message struct {
	Role    Role           `json:"role"`
	Content []ContentBlock `json:"content"`
}

// TextMessage is a convenience constructor for a plain text message.
func TextMessage(role Role, text string) Message {
	return Message{
		Role:    role,
		Content: []ContentBlock{{Type: "text", Text: text}},
	}
}

// ToolResultMessage builds a user message containing one or more tool results.
func ToolResultMessage(results []ToolResult) Message {
	blocks := make([]ContentBlock, len(results))
	for i, r := range results {
		blocks[i] = ContentBlock{
			Type:      "tool_result",
			ToolUseID: r.ToolUseID,
			Content:   r.Output,
			IsError:   r.IsError,
		}
	}
	return Message{Role: RoleUser, Content: blocks}
}

// ToolUse is a tool invocation requested by the model.
type ToolUse struct {
	ID    string
	Name  string
	Input map[string]any
}

// ToolResult is the output of a tool execution returned to the model.
type ToolResult struct {
	ToolUseID string
	Output    string
	IsError   bool
}

// MessagesRequest is the payload sent to POST /v1/messages.
type MessagesRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	Messages  []Message    `json:"messages"`
	Tools     []ToolSpec   `json:"tools,omitempty"`
	System    string       `json:"system,omitempty"`
	Stream    bool         `json:"stream"`
}

// ToolSpec is the API format for a tool definition.
type ToolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// APIEventType labels each SSE event emitted by the client.
type APIEventType string

const (
	EventTextDelta   APIEventType = "text_delta"
	EventToolUse     APIEventType = "tool_use"
	EventMessageStop APIEventType = "message_stop"
	EventError       APIEventType = "error"
)

// APIEvent is a parsed SSE event delivered through the streaming channel.
type APIEvent struct {
	Type       APIEventType
	Text       string   // non-empty for EventTextDelta
	ToolUse    *ToolUse // non-nil for EventToolUse
	StopReason string   // non-empty for EventMessageStop
	Error      error    // non-nil for EventError
	Usage      *Usage   // non-nil when usage data is available
}

// Usage tracks token consumption reported at end of stream.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
