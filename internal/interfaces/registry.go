package interfaces

import (
	"context"
	"fmt"
	"sync"

	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/tools"
)

// Provider defines the contract for an input/output interface provider.
// Each provider (Feishu, WeChat, Telegram, etc.) implements this interface.
type Provider interface {
	// Name returns the unique identifier for this provider.
	Name() string

	// Description returns a human-readable description of this provider.
	Description() string

	// Start initializes and starts the provider.
	// It should handle all setup (WebSocket connections, HTTP servers, etc.)
	// and block until the provider is stopped via the context.
	Start(ctx context.Context, config ProviderConfig) error

	// Stop gracefully shuts down the provider.
	Stop(ctx context.Context) error

	// SendMessage sends a message to the external platform.
	SendMessage(ctx context.Context, req SendMessageRequest) error

	// HealthCheck returns the current health status of the provider.
	HealthCheck(ctx context.Context) error
}

// ProviderConfig contains configuration common to all interface providers.
type ProviderConfig struct {
	// Agent configuration
	Model         string
	SystemPrompt  string
	MaxHistory    int

	// Tool and MCP configuration
	ToolRegistry  *tools.Registry
	APIClient     api.Streamer

	// Provider identification
	Provider      string

	// Provider-specific settings (passed through as map)
	Settings      map[string]any
}

// SendMessageRequest represents a request to send a message.
type SendMessageRequest struct {
	// ConversationID identifies the target conversation (chat ID, user ID, etc.)
	ConversationID string

	// MessageID is the ID of the message being replied to (optional)
	MessageID string

	// Content is the message content (text, markdown, or structured data)
	Content any

	// Format specifies the message format (text, markdown, card, etc.)
	Format MessageFormat

	// Options for platform-specific features
	Options map[string]any
}

// MessageFormat represents the format of a message.
type MessageFormat string

const (
	// FormatText is plain text.
	FormatText MessageFormat = "text"

	// FormatMarkdown is markdown-formatted text.
	FormatMarkdown MessageFormat = "markdown"

	// FormatCard is an interactive card (platform-specific).
	FormatCard MessageFormat = "card"

	// FormatImage is an image attachment.
	FormatImage MessageFormat = "image"

	// FormatFile is a file attachment.
	FormatFile MessageFormat = "file"
)

// Registry manages registered interface providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register registers a provider with the registry.
// This should typically be called from the provider package's init() function.
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p == nil {
		panic("cannot register nil provider")
	}

	name := p.Name()
	if name == "" {
		panic("provider name cannot be empty")
	}

	if _, exists := r.providers[name]; exists {
		panic(fmt.Sprintf("provider %q already registered", name))
	}

	r.providers[name] = p
}

// Get retrieves a provider by name.
func (r *Registry) Get(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// List returns all registered provider names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// globalRegistry is the default registry used by the application.
var globalRegistry = NewRegistry()

// Register registers a provider with the global registry.
func Register(p Provider) {
	globalRegistry.Register(p)
}

// Get retrieves a provider from the global registry.
func Get(name string) (Provider, bool) {
	return globalRegistry.Get(name)
}

// List returns all registered providers from the global registry.
func List() []string {
	return globalRegistry.List()
}