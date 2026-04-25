package feishu

import (
	"encoding/json"
	"os"
	"time"

	"github.com/atom-yt/claude-code-go/internal/interfaces"
)

// Config holds the configuration for the Feishu provider.
type Config struct {
	// Lark app credentials
	AppID             string `json:"app_id"`
	AppSecret         string `json:"app_secret"`
	EncryptKey        string `json:"encrypt_key,omitempty"`
	VerificationToken string `json:"verification_token,omitempty"`

	// Connection mode: "websocket", "webhook", or "dual"
	Mode ConnectionMode `json:"mode"`

	// WebSocket settings
	WebSocketURL string `json:"websocket_url,omitempty"`

	// Webhook settings
	WebhookURL    string `json:"webhook_url,omitempty"`
	WebhookPort   int    `json:"webhook_port,omitempty"`
	WebhookPath   string `json:"webhook_path,omitempty"`
	WebhookSecret string `json:"webhook_secret,omitempty"`

	// Session management
	MaxSessions     int           `json:"max_sessions"`
	SessionTimeout  time.Duration `json:"session_timeout"`
	PersistSessions bool          `json:"persist_sessions"`

	// Rate limiting
	RateLimitRequests int `json:"rate_limit_requests"` // Requests per second
	RateLimitBurst   int `json:"rate_limit_burst"`

	// Agent settings (can be overridden by ProviderConfig)
	Model           string `json:"model,omitempty"`
	SystemPrompt    string `json:"system_prompt,omitempty"`
	MaxHistorySize  int    `json:"max_history_size"`

	// Feature flags
	EnableMarkdown   bool `json:"enable_markdown"`
	EnableCards     bool `json:"enable_cards"`
	EnableImages    bool `json:"enable_images"`
	EnableReactions bool `json:"enable_reactions"`

	// Message formatting
	OutputFormat   OutputFormat `json:"output_format"`
	TruncateLength int          `json:"truncate_length"`
}

// ConnectionMode specifies how the provider connects to Feishu.
type ConnectionMode string

const (
	// ModeWebSocket uses WebSocket for real-time message push.
	ModeWebSocket ConnectionMode = "websocket"

	// ModeWebhook uses HTTP callbacks for event delivery.
	ModeWebhook ConnectionMode = "webhook"

	// ModeDual uses both WebSocket and Webhook for redundancy.
	ModeDual ConnectionMode = "dual"
)

// OutputFormat specifies how Agent output is formatted.
type OutputFormat string

const (
	// OutputFormatAuto automatically selects best format.
	OutputFormatAuto OutputFormat = "auto"

	// OutputFormatText uses plain text.
	OutputFormatText OutputFormat = "text"

	// OutputFormatMarkdown uses markdown formatting.
	OutputFormatMarkdown OutputFormat = "markdown"

	// OutputFormatCard uses interactive cards.
	OutputFormatCard OutputFormat = "card"
)

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Mode:             ModeDual,
		WebSocketURL:     "wss://open.feishu.cn/open-apis/bot/v4/ws",
		WebhookPort:      8080,
		WebhookPath:      "/webhook",
		MaxSessions:      100,
		SessionTimeout:   time.Hour,
		PersistSessions:  true,
		RateLimitRequests: 60,
		RateLimitBurst:   10,
		MaxHistorySize:   50,
		EnableMarkdown:   true,
		EnableCards:     true,
		EnableImages:    true,
		OutputFormat:    OutputFormatAuto,
		TruncateLength:   2000,
	}
}

// LoadConfig loads configuration from environment variables or settings map.
func LoadConfig(settings map[string]any) (*Config, error) {
	cfg := DefaultConfig()

	// Load from settings map
	if appID, ok := settings["app_id"].(string); ok && appID != "" {
		cfg.AppID = appID
	} else {
		cfg.AppID = os.Getenv("FEISHU_APP_ID")
	}

	if appSecret, ok := settings["app_secret"].(string); ok && appSecret != "" {
		cfg.AppSecret = appSecret
	} else {
		cfg.AppSecret = os.Getenv("FEISHU_APP_SECRET")
	}

	if encryptKey, ok := settings["encrypt_key"].(string); ok {
		cfg.EncryptKey = encryptKey
	} else {
		cfg.EncryptKey = os.Getenv("FEISHU_ENCRYPT_KEY")
	}

	if verifyToken, ok := settings["verification_token"].(string); ok {
		cfg.VerificationToken = verifyToken
	} else {
		cfg.VerificationToken = os.Getenv("FEISHU_VERIFICATION_TOKEN")
	}

	if modeStr, ok := settings["mode"].(string); ok {
		cfg.Mode = ConnectionMode(modeStr)
	}

	if wsURL, ok := settings["websocket_url"].(string); ok {
		cfg.WebSocketURL = wsURL
	}

	if webhookURL, ok := settings["webhook_url"].(string); ok {
		cfg.WebhookURL = webhookURL
	}

	if port, ok := settings["webhook_port"].(float64); ok {
		cfg.WebhookPort = int(port)
	}

	if path, ok := settings["webhook_path"].(string); ok {
		cfg.WebhookPath = path
	}

	if secret, ok := settings["webhook_secret"].(string); ok {
		cfg.WebhookSecret = secret
	}

	if maxSessions, ok := settings["max_sessions"].(float64); ok {
		cfg.MaxSessions = int(maxSessions)
	}

	if timeout, ok := settings["session_timeout"].(string); ok {
		if dur, err := time.ParseDuration(timeout); err == nil {
			cfg.SessionTimeout = dur
		}
	}

	if persist, ok := settings["persist_sessions"].(bool); ok {
		cfg.PersistSessions = persist
	}

	if rateLimit, ok := settings["rate_limit_requests"].(float64); ok {
		cfg.RateLimitRequests = int(rateLimit)
	}

	if burst, ok := settings["rate_limit_burst"].(float64); ok {
		cfg.RateLimitBurst = int(burst)
	}

	if model, ok := settings["model"].(string); ok {
		cfg.Model = model
	}

	if sysPrompt, ok := settings["system_prompt"].(string); ok {
		cfg.SystemPrompt = sysPrompt
	}

	if historySize, ok := settings["max_history_size"].(float64); ok {
		cfg.MaxHistorySize = int(historySize)
	}

	if md, ok := settings["enable_markdown"].(bool); ok {
		cfg.EnableMarkdown = md
	}

	if cards, ok := settings["enable_cards"].(bool); ok {
		cfg.EnableCards = cards
	}

	if images, ok := settings["enable_images"].(bool); ok {
		cfg.EnableImages = images
	}

	if reactions, ok := settings["enable_reactions"].(bool); ok {
		cfg.EnableReactions = reactions
	}

	if fmt, ok := settings["output_format"].(string); ok {
		cfg.OutputFormat = OutputFormat(fmt)
	}

	if truncate, ok := settings["truncate_length"].(float64); ok {
		cfg.TruncateLength = int(truncate)
	}

	return cfg, nil
}

// ToProviderConfig converts Feishu config to generic ProviderConfig.
func (c *Config) ToProviderConfig(registry any, client any) interfaces.ProviderConfig {
	return interfaces.ProviderConfig{
		Model:        c.Model,
		SystemPrompt: c.SystemPrompt,
		MaxHistory:   c.MaxHistorySize,
		Settings: map[string]any{
			"config": c,
		},
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.AppID == "" {
		return ErrMissingAppID
	}
	if c.AppSecret == "" {
		return ErrMissingAppSecret
	}

	switch c.Mode {
	case ModeWebSocket:
		if c.WebSocketURL == "" {
			return ErrMissingWebSocketURL
		}
	case ModeWebhook:
		if c.WebhookPort <= 0 {
			return ErrInvalidWebhookPort
		}
	case ModeDual:
		if c.WebSocketURL == "" {
			return ErrMissingWebSocketURL
		}
		if c.WebhookPort <= 0 {
			return ErrInvalidWebhookPort
		}
	default:
		return ErrInvalidMode
	}

	if c.MaxSessions <= 0 {
		c.MaxSessions = 100
	}

	if c.RateLimitRequests <= 0 {
		c.RateLimitRequests = 60
	}

	if c.RateLimitBurst <= 0 {
		c.RateLimitBurst = 10
	}

	return nil
}

// Configuration errors
var (
	ErrMissingAppID         = &ConfigError{field: "app_id", message: "app_id is required"}
	ErrMissingAppSecret     = &ConfigError{field: "app_secret", message: "app_secret is required"}
	ErrMissingWebSocketURL  = &ConfigError{field: "websocket_url", message: "websocket_url is required"}
	ErrInvalidWebhookPort  = &ConfigError{field: "webhook_port", message: "webhook_port must be > 0"}
	ErrInvalidMode          = &ConfigError{field: "mode", message: "invalid mode (use websocket, webhook, or dual)"}
)

// ConfigError represents a configuration validation error.
type ConfigError struct {
	field   string
	message string
}

func (e *ConfigError) Error() string {
	return e.message
}

func (e *ConfigError) Field() string {
	return e.field
}

// MarshalJSON implements json.Marshaler for ConfigError.
func (e *ConfigError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"field":   e.field,
		"message": e.message,
	})
}