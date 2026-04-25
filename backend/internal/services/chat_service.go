package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/atom-yt/claude-code-go/pkg/agent"
)

// ChatRequest represents a chat request from user
type ChatRequest struct {
	SessionID   string        `json:"session_id"`
	Message     string        `json:"message"`
	AgentConfig *AgentConfig  `json:"agent_config,omitempty"`
	Stream      bool          `json:"stream"`
}

// AgentConfig holds agent configuration
type AgentConfig struct {
	Model        string `json:"model"`
	Provider     string `json:"provider"`
	APIKey       string `json:"api_key"`
	BaseURL      string `json:"base_url"`
	SystemPrompt string `json:"system_prompt"`
}

// ChatEvent represents a chat event (streaming or final result)
type ChatEvent struct {
	Type      string                 `json:"type"` // "delta", "tool_call", "tool_result", "error", "done"
	SessionID string                 `json:"session_id"`
	Data      map[string]interface{} `json:"data"`
}

const (
	EventTypeDelta       = "delta"
	EventTypeToolCall    = "tool_call"
	EventTypeToolResult  = "tool_result"
	EventTypeError       = "error"
	EventTypeDone        = "done"
)

var (
	// ErrInvalidAgentConfig is returned when agent configuration is invalid
	ErrInvalidAgentConfig = fmt.Errorf("invalid agent configuration")
	// ErrAgentCreationFailed is returned when agent creation fails
	ErrAgentCreationFailed = fmt.Errorf("agent creation failed")
)

// ChatService manages chat sessions and agent instances
type ChatService struct {
	// Default configuration
	defaultProvider string
	defaultModel   string

	// Agent instances cache (sessionID -> *agent.ChatAgent)
	mu     sync.RWMutex
	agents map[string]*agent.ChatAgent
}

// NewChatService creates a new chat service
func NewChatService(apiKey, baseURL, defaultProvider, defaultModel string) *ChatService {
	return &ChatService{
		defaultProvider: defaultProvider,
		defaultModel:   defaultModel,
		agents:         make(map[string]*agent.ChatAgent),
	}
}

// Chat sends a message to agent and returns a channel of events
func (s *ChatService) Chat(ctx context.Context, req *ChatRequest) (<-chan ChatEvent, error) {
	// Get or create agent for this session
	agt, err := s.getOrCreateAgent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAgentCreationFailed, err)
	}

	// Query := agent
	eventCh := agt.Chat(ctx, req.Message)

	// Convert agent events to chat events
	chatCh := make(chan ChatEvent, 64)
	go s.forwardEvents(ctx, eventCh, chatCh, req.SessionID)

	return chatCh, nil
}

// ChatSync sends a message and returns complete response (non-streaming)
func (s *ChatService) ChatSync(ctx context.Context, req *ChatRequest) (string, error) {
	if req.Stream {
		return "", fmt.Errorf("streaming is not supported in sync mode")
	}

	chatCh, err := s.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	var fullText string
	for event := range chatCh {
		switch event.Type {
		case EventTypeDelta:
			if text, ok := event.Data["text"].(string); ok {
				fullText += text
			}
		case EventTypeError:
			if errMsg, ok := event.Data["error"].(string); ok {
				return "", fmt.Errorf("agent error: %s", errMsg)
			}
		case EventTypeDone:
			return fullText, nil
		}
	}

	return fullText, nil
}

// getOrCreateAgent gets an existing agent or creates a new one for session
func (s *ChatService) getOrCreateAgent(ctx context.Context, req *ChatRequest) (*agent.ChatAgent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if agent already exists for this session
	if agt, exists := s.agents[req.SessionID]; exists {
		return agt, nil
	}

	// Determine configuration
	cfg := s.defaultAgentConfig()
	if req.AgentConfig != nil {
		if req.AgentConfig.Model != "" {
			cfg.Model = req.AgentConfig.Model
		}
		if req.AgentConfig.Provider != "" {
			cfg.Provider = req.AgentConfig.Provider
		}
		if req.AgentConfig.APIKey != "" {
			cfg.APIKey = req.AgentConfig.APIKey
		}
		if req.AgentConfig.BaseURL != "" {
			cfg.BaseURL = req.AgentConfig.BaseURL
		}
		if req.AgentConfig.SystemPrompt != "" {
			cfg.SystemPrompt = req.AgentConfig.SystemPrompt
		}
	}

	// Create agent
	agt, err := agent.New(cfg)
	if err != nil {
		return nil, err
	}

	// Cache := agent
	s.agents[req.SessionID] = agt

	return agt, nil
}

// defaultAgentConfig returns default agent configuration
func (s *ChatService) defaultAgentConfig() *agent.Config {
	return &agent.Config{
		Model:   s.defaultModel,
		Provider: s.defaultProvider,
		// API key and base URL should be set from env or request
		APIKey:  "",
		BaseURL: "",
		Permissions: &agent.PermissionsConfig{
			AllowAll: true,
		},
	}
}

// forwardEvents converts agent StreamEvents to ChatEvents
func (s *ChatService) forwardEvents(ctx context.Context, agentCh <-chan agent.StreamEvent, chatCh chan<- ChatEvent, sessionID string) {
	defer close(chatCh)

	for {
		select {
		case <-ctx.Done():
			chatCh <- ChatEvent{
				Type:      EventTypeError,
				SessionID: sessionID,
				Data: map[string]interface{}{
					"error": "request cancelled",
				},
			}
			return

		case ev, ok := <-agentCh:
			if !ok {
				return
			}

			switch ev.Type {
			case agent.EventTypeDelta:
				chatCh <- ChatEvent{
					Type:      EventTypeDelta,
					SessionID: sessionID,
					Data: map[string]interface{}{
						"text": ev.Text,
					},
				}
				if ev.Usage != nil {
					chatCh <- ChatEvent{
						Type: EventTypeDelta,
						Data: map[string]interface{}{
							"usage": ev.Usage,
						},
					}
				}

			case agent.EventTypeToolCall:
				if ev.ToolName != "" {
					chatCh <- ChatEvent{
						Type:      EventTypeToolCall,
						SessionID: sessionID,
						Data: map[string]interface{}{
							"tool":  ev.ToolName,
							"input": ev.ToolInput,
						},
					}
				}

			case agent.EventTypeToolResult:
				if ev.ToolName != "" {
					chatCh <- ChatEvent{
						Type:       EventTypeToolResult,
						SessionID: sessionID,
						Data: map[string]interface{}{
							"tool":     ev.ToolName,
							"output":   ev.ToolOutput,
							"is_error": ev.ToolIsError,
						},
					}
				}

			case agent.EventTypeError:
				chatCh <- ChatEvent{
					Type:      EventTypeError,
					SessionID: sessionID,
					Data: map[string]interface{}{
						"error": ev.Error.Error(),
					},
				}
				return

			case agent.EventTypeDone:
				chatCh <- ChatEvent{
					Type:      EventTypeDone,
					SessionID: sessionID,
					Data:      map[string]interface{}{},
				}
				return
			}
		}
	}
}

// RemoveSession removes an agent from the cache
func (s *ChatService) RemoveSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.agents, sessionID)
}

// GetSessionCount returns the number of active sessions
func (s *ChatService) GetSessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.agents)
}