package feishu

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/atom-yt/claude-code-go/internal/agent"
	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/interfaces"
)

// EventHandler handles incoming Feishu events.
type EventHandler struct {
	config       *Config
	client       *Client
	sessions     *SessionManager
	queue        *MessageQueue
	formatter    *Formatter
	media        *MediaHandler
	rateLimiter  *RateLimiter

	mu           sync.RWMutex
	processing   map[string]bool // Track messages being processed
}

// NewEventHandler creates a new event handler.
func NewEventHandler(
	cfg *Config,
	client *Client,
	sessions *SessionManager,
	queue *MessageQueue,
	formatter *Formatter,
	media *MediaHandler,
) *EventHandler {
	return &EventHandler{
		config:     cfg,
		client:     client,
		sessions:   sessions,
		queue:      queue,
		formatter:  formatter,
		media:      media,
		rateLimiter: NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitBurst),
		processing: make(map[string]bool),
	}
}

// HandleMessage processes an incoming message.
func (eh *EventHandler) HandleMessage(ctx context.Context, msg *IncomingMessage) {
	// Check rate limit
	if err := eh.rateLimiter.WaitUser(ctx, msg.UserID); err != nil {
		eh.logError("rate limit exceeded for user %s: %v", msg.UserID, err)
		return
	}

	// Check if already processing
	key := eh.messageKey(msg)
	eh.mu.Lock()
	if eh.processing[key] {
		eh.mu.Unlock()
		return
	}
	eh.processing[key] = true
	eh.mu.Unlock()
	defer func() {
		eh.mu.Lock()
		delete(eh.processing, key)
		eh.mu.Unlock()
	}()

	// Get or create session
	session, err := eh.sessions.GetOrCreate(
		msg.ChatID,
		msg.UserID,
		msg.UserName,
		msg.TenantKey,
	)
	if err != nil {
		eh.logError("failed to get session: %v", err)
		eh.sendErrorResponse(ctx, msg.ChatID, "Session error")
		return
	}

	// Process media if present
	if len(msg.Images) > 0 || len(msg.Files) > 0 {
		go eh.processMedia(ctx, session, msg)
	}

	// Build input message
	input := eh.buildInputMessage(msg)

	// Stream response to Feishu
	if err := eh.streamResponse(ctx, session, msg.ChatID, input); err != nil {
		eh.logError("failed to stream response: %v", err)
		eh.sendErrorResponse(ctx, msg.ChatID, "Response error")
	}
}

// buildInputMessage builds the input message for the Agent.
func (eh *EventHandler) buildInputMessage(msg *IncomingMessage) string {
	var sb strings.Builder

	sb.WriteString(msg.Content)

	// Add image references
	if len(msg.Images) > 0 {
		sb.WriteString("\n\n[Attached images: ")
		for i, img := range msg.Images {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(img.ImageKey)
		}
		sb.WriteString("]")
	}

	// Add file references
	if len(msg.Files) > 0 {
		sb.WriteString("\n\n[Attached files: ")
		for i, file := range msg.Files {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(file.Name)
		}
		sb.WriteString("]")
	}

	return sb.String()
}

// streamResponse streams Agent response to Feishu.
func (eh *EventHandler) streamResponse(
	ctx context.Context,
	session *Session,
	chatID string,
	input string,
) error {
	// Add user message to history
	eh.sessions.AddMessage(chatID, api.TextMessage(api.RoleUser, input))

	// Track streaming state
	var fullResponse strings.Builder
	var streamingMu sync.Mutex

	// Start agent query
	eventCh := session.Agent.Query(ctx, input)

	// Process events from the channel
	for event := range eventCh {
		streamingMu.Lock()

		switch event.Type {
		case agent.EventTextDelta:
			fullResponse.WriteString(event.Text)

			// Send incremental updates (throttled)
			if fullResponse.Len() >= 500 || fullResponse.Len()%200 == 0 {
				eh.sendUpdate(ctx, chatID, fullResponse.String())
			}

		case agent.EventToolCall:
			// Tool use event
			eh.sendToolNotification(ctx, chatID, event.ToolName)

		case agent.EventToolResult:
			// Tool result event - could be shown for feedback

		case agent.EventError:
			eh.logError("stream error: %v", event.Error)

		case agent.EventDone:
			// Send final response
			eh.sendFinalResponse(ctx, chatID, fullResponse.String())

			// Add to history
			eh.sessions.AddMessage(chatID, api.TextMessage(api.RoleAssistant, fullResponse.String()))
		}

		streamingMu.Unlock()
	}

	return nil
}

// sendUpdate sends an update message.
func (eh *EventHandler) sendUpdate(ctx context.Context, chatID, content string) {
	// Format and send
	if eh.formatter.ShouldUseCard(content) {
		card := eh.formatter.FormatCard("Response", content)
		_ = eh.client.SendCard(ctx, chatID, card)
	} else {
		_ = eh.client.SendMarkdown(ctx, chatID, eh.formatter.FormatMarkdown(content))
	}
}

// sendFinalResponse sends the final response.
func (eh *EventHandler) sendFinalResponse(ctx context.Context, chatID, content string) {
	if content == "" {
		return
	}

	// Determine format based on content
	format := interfaces.FormatText
	if eh.formatter.ShouldUseCard(content) {
		format = interfaces.FormatCard
	} else if eh.config.EnableMarkdown {
		format = interfaces.FormatMarkdown
	}

	switch format {
	case interfaces.FormatCard:
		card := eh.formatter.FormatCard("Response", content)
		_ = eh.client.SendCard(ctx, chatID, card)

	case interfaces.FormatMarkdown:
		_ = eh.client.SendMarkdown(ctx, chatID, eh.formatter.FormatMarkdown(content))

	default:
		_ = eh.client.SendText(ctx, chatID, eh.formatter.FormatText(content))
	}
}

// sendToolNotification sends a tool usage notification.
func (eh *EventHandler) sendToolNotification(ctx context.Context, chatID, toolName string) {
	notification := fmt.Sprintf("**Using tool:** %s", toolName)
	_ = eh.client.SendMarkdown(ctx, chatID, notification)
}

// sendErrorResponse sends an error response.
func (eh *EventHandler) sendErrorResponse(ctx context.Context, chatID, message string) {
	card := eh.formatter.FormatCard("Error", message)
	_ = eh.client.SendCard(ctx, chatID, card)
}

// processMedia processes attached media files.
func (eh *EventHandler) processMedia(ctx context.Context, session *Session, msg *IncomingMessage) {
	// Process images
	for _, img := range msg.Images {
		_, err := eh.media.ProcessImage(ctx, &img)
		if err != nil {
			eh.logError("failed to process image %s: %v", img.ImageKey, err)
		}
	}

	// Process files
	for _, file := range msg.Files {
		_, err := eh.media.ProcessFile(ctx, &file)
		if err != nil {
			eh.logError("failed to process file %s: %v", file.FileKey, err)
		}
	}
}

// HandleCardAction handles actions from interactive cards.
func (eh *EventHandler) HandleCardAction(ctx context.Context, chatID, userID string, action *CardAction) {
	// Get session
	_, exists := eh.sessions.Get(chatID)
	if !exists {
		eh.logError("session not found for card action")
		return
	}

	// Process action based on action ID
	switch action.Action {
	case "clear_history":
		if err := eh.sessions.ClearHistory(chatID); err == nil {
			_ = eh.client.SendText(ctx, chatID, "History cleared.")
		}

	case "new_session":
		eh.sessions.Remove(chatID)
		_ = eh.client.SendText(ctx, chatID, "Started new session.")

	case "help":
		helpText := `
**Available Commands:**
• Type any message to chat with Claude
• Cards provide quick actions
• History is automatically managed
		`
		_ = eh.client.SendMarkdown(ctx, chatID, helpText)

	default:
		// Custom action
		input := fmt.Sprintf("[Action: %s]", action.Action)
		eh.HandleMessage(ctx, &IncomingMessage{
			ChatID:    chatID,
			UserID:    userID,
			Content:   input,
			Timestamp: time.Now(),
		})
	}
}

// messageKey generates a unique key for a message.
func (eh *EventHandler) messageKey(msg *IncomingMessage) string {
	return msg.ChatID + ":" + msg.MessageID
}

// logError logs an error message.
func (eh *EventHandler) logError(format string, args ...any) {
	fmt.Printf("[EventHandler] ERROR: "+format+"\n", args...)
}

// Stats returns handler statistics.
func (eh *EventHandler) Stats() map[string]any {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	return map[string]any{
		"processing_count": len(eh.processing),
		"active_sessions":  eh.sessions.ActiveCount(),
		"queue_size":       eh.queue.Size(),
	}
}