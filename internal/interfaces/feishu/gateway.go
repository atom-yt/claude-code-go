package feishu

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/interfaces"
	"github.com/atom-yt/claude-code-go/internal/tools"
)

// Gateway implements the interfaces.Provider interface for Feishu.
// It orchestrates all Feishu components: WebSocket, Webhook, SessionManager, etc.
type Gateway struct {
	config   *Config

	// Components
	client      *Client
	sessions    *SessionManager
	queue       *MessageQueue
	formatter   *Formatter
	media       *MediaHandler
	handler     *EventHandler

	// Connections
	websocket   *WebSocketClient
	webhook     *WebhookServer

	// State
	mu          sync.RWMutex
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewGateway creates a new Feishu gateway.
func NewGateway() *Gateway {
	return &Gateway{}
}

// Name returns the provider name.
func (g *Gateway) Name() string {
	return "feishu"
}

// Description returns a human-readable description.
func (g *Gateway) Description() string {
	return "Feishu/Lark integration provider with WebSocket and Webhook support"
}

// Start initializes and starts the gateway.
func (g *Gateway) Start(ctx context.Context, cfg interfaces.ProviderConfig) error {
	g.mu.Lock()
	if g.running {
		g.mu.Unlock()
		return fmt.Errorf("gateway already running")
	}

	g.ctx, g.cancel = context.WithCancel(ctx)
	g.mu.Unlock()

	// Extract Feishu config
	feishuCfg, err := g.extractConfig(cfg.Settings)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Validate config
	if err := feishuCfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	g.config = feishuCfg

	// Initialize components
	g.client = NewClient(feishuCfg)
	g.formatter = NewFormatter(feishuCfg)
	g.media = NewMediaHandler(g.client, feishuCfg)

	// Initialize API client and tool registry
	var apiClient api.Streamer
	var toolReg *tools.Registry

	if cfg.APIClient != nil {
		apiClient = cfg.APIClient
	}

	if cfg.ToolRegistry != nil {
		toolReg = cfg.ToolRegistry
	}

	// Initialize session manager
	model := cfg.Model
	provider := cfg.Provider
	if feishuCfg.Model != "" {
		model = feishuCfg.Model
	}

	g.sessions = NewSessionManager(g.ctx, feishuCfg, toolReg, apiClient, model, provider)

	// Initialize message queue
	g.queue = NewMessageQueue(g.ctx, feishuCfg.RateLimitRequests, feishuCfg.RateLimitBurst)

	// Initialize event handler
	g.handler = NewEventHandler(
		feishuCfg,
		g.client,
		g.sessions,
		g.queue,
		g.formatter,
		g.media,
	)

	// Start based on mode
	switch feishuCfg.Mode {
	case ModeWebSocket:
		if err := g.startWebSocket(g.ctx); err != nil {
			return err
		}

	case ModeWebhook:
		if err := g.startWebhook(g.ctx); err != nil {
			return err
		}

	case ModeDual:
		if err := g.startWebSocket(g.ctx); err != nil {
			return err
		}
		if err := g.startWebhook(g.ctx); err != nil {
			g.Stop(g.ctx)
			return err
		}

	default:
		return fmt.Errorf("invalid mode: %s", feishuCfg.Mode)
	}

	// Start cleanup routine
	g.wg.Add(1)
	go g.cleanupLoop()

	g.mu.Lock()
	g.running = true
	g.mu.Unlock()

	return nil
}

// Stop gracefully shuts down the gateway.
func (g *Gateway) Stop(ctx context.Context) error {
	g.mu.Lock()
	if !g.running {
		g.mu.Unlock()
		return nil
	}
	g.mu.Unlock()

	// Signal shutdown
	g.cancel()

	// Stop WebSocket
	if g.websocket != nil {
		_ = g.websocket.Disconnect()
	}

	// Stop webhook
	if g.webhook != nil {
		_ = g.webhook.Shutdown(ctx)
	}

	// Close queue
	if g.queue != nil {
		g.queue.Close()
	}

	// Close sessions
	if g.sessions != nil {
		g.sessions.Close()
	}

	// Wait for goroutines
	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Clean shutdown
		return nil
	case <-ctx.Done():
		// Force shutdown timeout
		return ctx.Err()
	}
}

// SendMessage sends a message to Feishu.
func (g *Gateway) SendMessage(ctx context.Context, req interfaces.SendMessageRequest) error {
	if g.client == nil {
		return fmt.Errorf("client not initialized")
	}

	switch req.Format {
	case interfaces.FormatText:
		content, ok := req.Content.(string)
		if !ok {
			return fmt.Errorf("invalid content type for text format")
		}
		return g.client.SendText(ctx, req.ConversationID, content)

	case interfaces.FormatMarkdown:
		content, ok := req.Content.(string)
		if !ok {
			return fmt.Errorf("invalid content type for markdown format")
		}
		return g.client.SendMarkdown(ctx, req.ConversationID, content)

	case interfaces.FormatCard:
		card, ok := req.Content.(*Card)
		if !ok {
			return fmt.Errorf("invalid content type for card format")
		}
		return g.client.SendCard(ctx, req.ConversationID, card)

	default:
		return fmt.Errorf("unsupported format: %s", req.Format)
	}
}

// HealthCheck checks the health of the gateway.
func (g *Gateway) HealthCheck(ctx context.Context) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.running {
		return fmt.Errorf("gateway not running")
	}

	// Check client
	if g.client == nil {
		return fmt.Errorf("client not initialized")
	}

	// Test API access
	_, err := g.client.GetTenantAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("API access failed: %w", err)
	}

	return nil
}

// startWebSocket starts the WebSocket connection.
func (g *Gateway) startWebSocket(ctx context.Context) error {
	g.websocket = NewWebSocketClient(g.config, g.handler)

	if err := g.websocket.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect WebSocket: %w", err)
	}

	// Monitor connection
	g.wg.Add(1)
	go g.monitorWebSocket(ctx)

	return nil
}

// startWebhook starts the webhook server.
func (g *Gateway) startWebhook(ctx context.Context) error {
	g.webhook = NewWebhookServer(g.config, g.handler)

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := g.webhook.Start(ctx); err != nil {
			// Webhook startup error
			g.cancel()
		}
	}()

	return nil
}

// monitorWebSocket monitors the WebSocket connection and reconnects if needed.
func (g *Gateway) monitorWebSocket(ctx context.Context) {
	defer g.wg.Done()

	reconnectDelay := 5 * time.Second
	maxDelay := 5 * time.Minute

	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(5 * time.Second):
			// Check connection status
			if g.websocket != nil && g.websocket.IsConnected() {
				continue
			}

			// Connection lost, try to reconnect
			g.logInfo("WebSocket disconnected, reconnecting in %v...", reconnectDelay)

			select {
			case <-time.After(reconnectDelay):
			case <-ctx.Done():
				return
			}

			if g.websocket == nil {
				g.websocket = NewWebSocketClient(g.config, g.handler)
			}

			if err := g.websocket.Connect(ctx); err != nil {
				g.logError("reconnect failed: %v", err)
				reconnectDelay = min(reconnectDelay*2, maxDelay)
			} else {
				g.logInfo("WebSocket reconnected")
				reconnectDelay = 5 * time.Second
			}
		}
	}
}

// cleanupLoop periodically cleans up inactive sessions.
func (g *Gateway) cleanupLoop() {
	defer g.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			g.sessions.CleanupInactive()
		}
	}
}

// extractConfig extracts Feishu config from provider settings.
func (g *Gateway) extractConfig(settings map[string]any) (*Config, error) {
	cfg, err := LoadConfig(settings)
	if err != nil {
		return nil, err
	}

	// Apply defaults from ProviderConfig if set
	// These take precedence over Feishu config settings
	return cfg, nil
}

// Stats returns gateway statistics.
func (g *Gateway) Stats() map[string]any {
	g.mu.RLock()
	defer g.mu.RUnlock()

	stats := map[string]any{
		"running":    g.running,
		"mode":       string(g.config.Mode),
		"sessions":   g.sessions.ActiveCount(),
		"queue_size": g.queue.Size(),
	}

	if g.handler != nil {
		handlerStats := g.handler.Stats()
		for k, v := range handlerStats {
			stats[k] = v
		}
	}

	return stats
}

// logInfo logs an info message.
func (g *Gateway) logInfo(format string, args ...any) {
	fmt.Printf("[FeishuGateway] INFO: "+format+"\n", args...)
}

// logError logs an error message.
func (g *Gateway) logError(format string, args ...any) {
	fmt.Printf("[FeishuGateway] ERROR: "+format+"\n", args...)
}

// min returns the minimum of two durations.
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// Register the Feishu provider on package initialization.
func init() {
	interfaces.Register(&Gateway{})
}