// Package agent provides a public API for the AI agent functionality.
//
// This package exports core agent capabilities that can be used by external
// services and modules (like the backend HTTP server).
//
// The main component is the ChatAgent, which provides streaming chat responses
// with tool execution support.
package agent

import (
	"context"
	"sync"

	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/tools"
)

// EventType represents the type of stream event
type EventType string

const (
	// EventTypeDelta is emitted when text content is generated
	EventTypeDelta EventType = "delta"
	// EventTypeToolCall is emitted when a tool is about to be executed
	EventTypeToolCall EventType = "tool_call"
	// EventTypeToolResult is emitted when a tool execution completes
	EventTypeToolResult EventType = "tool_result"
	// EventTypeError is emitted when an error occurs
	EventTypeError EventType = "error"
	// EventTypeDone is emitted when the response is complete
	EventTypeDone EventType = "done"
)

// StreamEvent represents a streaming event from the agent
type StreamEvent struct {
	Type        EventType
	Text        string
	ToolName    string
	ToolInput   map[string]any
	ToolOutput  string
	ToolIsError bool
	Error       error
	Usage       *api.Usage
}

// Config holds agent configuration
type Config struct {
	// API configuration
	APIKey   string
	BaseURL  string
	Model    string
	Provider string

	// Agent behavior
	SystemPrompt string

	// Tools
	Tools []string

	// Permissions (optional)
	Permissions *PermissionsConfig
}

// PermissionsConfig configures permission checking
type PermissionsConfig struct {
	AllowAll bool
}

// ChatAgent provides streaming chat functionality
type ChatAgent struct {
	mu       sync.RWMutex
	client   api.Streamer
	model    string
	provider string
	registry *tools.Registry
	system   string
}

// AgentFactory defines interface for creating ChatAgent instances
type AgentFactory interface {
	Create(ctx context.Context, cfg *Config) (*ChatAgent, error)
}

// ConfigFactory creates ChatAgent instances with given configuration
type ConfigFactory struct {
	defaultAPIKey    string
	defaultBaseURL   string
	defaultProvider string
	defaultModel   string
}

// NewConfigFactory creates a new agent factory
func NewConfigFactory(apiKey, baseURL, provider, model string) *ConfigFactory {
	return &ConfigFactory{
		defaultAPIKey:    apiKey,
		defaultBaseURL:   baseURL,
		defaultProvider: provider,
		defaultModel:   model,
	}
}

// Create creates a ChatAgent with the given configuration
func (f *ConfigFactory) Create(ctx context.Context, cfg *Config) (*ChatAgent, error) {
	// Use factory defaults for any nil values
	if cfg.APIKey == "" {
		cfg.APIKey = f.defaultAPIKey
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = f.defaultBaseURL
	}
	if cfg.Provider == "" {
		cfg.Provider = f.defaultProvider
	}
	if cfg.Model == "" {
		cfg.Model = f.defaultModel
	}

	return New(cfg)
}

// New creates a new ChatAgent
func New(cfg *Config) (*ChatAgent, error) {
	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-6"
	}
	if cfg.Provider == "" {
		cfg.Provider = "anthropic"
	}

	// Create API client
	client := createAPIClient(cfg.Provider, cfg.APIKey, cfg.BaseURL)

	// Create tool registry
	registry := tools.NewRegistry()

	return &ChatAgent{
		client:   client,
		model:    cfg.Model,
		provider: cfg.Provider,
		registry: registry,
		system:   cfg.SystemPrompt,
	}, nil
}

// SetSystemPrompt sets the system prompt for the agent
func (a *ChatAgent) SetSystemPrompt(prompt string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.system = prompt
}

// Chat sends a message and returns a channel of streaming events
func (a *ChatAgent) Chat(ctx context.Context, message string) <-chan StreamEvent {
	eventCh := make(chan StreamEvent, 64)
	go func() {
		defer close(eventCh)
		a.run(ctx, message, eventCh)
	}()
	return eventCh
}

func (a *ChatAgent) run(ctx context.Context, message string, ch chan<- StreamEvent) {
	a.mu.RLock()
	client := a.client
	model := a.model
	system := a.system
	a.mu.RUnlock()

	// Build API request
	req := api.MessagesRequest{
		Model:    model,
		Messages: []api.Message{api.TextMessage(api.RoleUser, message)},
	}
	if system != "" {
		req.SetSystemWithCaching(system)
	}

	// Call API
	apiCh := client.StreamMessages(ctx, req)

	for ev := range apiCh {
		switch ev.Type {
		case api.EventTextDelta:
			ch <- StreamEvent{Type: EventTypeDelta, Text: ev.Text}
			if ev.Usage != nil {
				ch <- StreamEvent{Type: EventTypeDelta, Usage: ev.Usage}
			}

		case api.EventToolUse:
			if ev.ToolUse != nil {
				ch <- StreamEvent{
					Type:      EventTypeToolCall,
					ToolName:  ev.ToolUse.Name,
					ToolInput: ev.ToolUse.Input,
				}
			}

		case api.EventError:
			ch <- StreamEvent{Type: EventTypeError, Error: ev.Error}
			return

		case api.EventMessageStop:
			ch <- StreamEvent{Type: EventTypeDone}
		}
	}
}

// createAPIClient creates an API client based on provider
func createAPIClient(provider, apiKey, baseURL string) api.Streamer {
	switch provider {
	case "anthropic", "":
		if baseURL != "" {
			return api.NewWithBaseURL(apiKey, baseURL)
		}
		return api.New(apiKey)

	case "openai", "kimi", "deepseek", "qwen", "codex", "ark", "ark-openai":
		// Use OpenAI client
		return api.NewOpenAI(apiKey, baseURL)

	case "ark-anthropic":
		if baseURL != "" {
			return api.NewWithBaseURL(apiKey, baseURL)
		}
		return api.New(apiKey)

	default:
		return api.New(apiKey)
	}
}