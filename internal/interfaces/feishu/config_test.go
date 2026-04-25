package feishu

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, ModeDual, cfg.Mode)
	assert.Equal(t, "wss://open.feishu.cn/open-apis/bot/v4/ws", cfg.WebSocketURL)
	assert.Equal(t, 8080, cfg.WebhookPort)
	assert.Equal(t, "/webhook", cfg.WebhookPath)
	assert.Equal(t, 100, cfg.MaxSessions)
	assert.Equal(t, time.Hour, cfg.SessionTimeout)
	assert.Equal(t, 60, cfg.RateLimitRequests)
	assert.Equal(t, 10, cfg.RateLimitBurst)
	assert.Equal(t, 50, cfg.MaxHistorySize)
	assert.True(t, cfg.EnableMarkdown)
	assert.True(t, cfg.EnableCards)
	assert.True(t, cfg.EnableImages)
	assert.Equal(t, OutputFormatAuto, cfg.OutputFormat)
	assert.Equal(t, 2000, cfg.TruncateLength)
}

func TestConfigLoadFromMap(t *testing.T) {
	settings := map[string]any{
		"app_id":         "cli_test",
		"app_secret":     "secret123",
		"mode":           "webhook",
		"webhook_port":   float64(9000),
		"webhook_path":   "/api/webhook",
		"max_sessions":   float64(50),
		"enable_markdown": false,
	}

	cfg, err := LoadConfig(settings)
	require.NoError(t, err)

	assert.Equal(t, "cli_test", cfg.AppID)
	assert.Equal(t, "secret123", cfg.AppSecret)
	assert.Equal(t, ModeWebhook, cfg.Mode)
	assert.Equal(t, 9000, cfg.WebhookPort)
	assert.Equal(t, "/api/webhook", cfg.WebhookPath)
	assert.Equal(t, 50, cfg.MaxSessions)
	assert.False(t, cfg.EnableMarkdown)
}

func TestConfigValidation(t *testing.T) {
	t.Run("Missing AppID", func(t *testing.T) {
		cfg := &Config{
			AppSecret: "secret",
			Mode:      ModeWebhook,
		}
		err := cfg.Validate()
		assert.Equal(t, ErrMissingAppID, err)
	})

	t.Run("Missing AppSecret", func(t *testing.T) {
		cfg := &Config{
			AppID: "cli_test",
			Mode:  ModeWebhook,
		}
		err := cfg.Validate()
		assert.Equal(t, ErrMissingAppSecret, err)
	})

	t.Run("Invalid Webhook Port", func(t *testing.T) {
		cfg := &Config{
			AppID:       "cli_test",
			AppSecret:   "secret",
			Mode:        ModeWebhook,
			WebhookPort: 0,
		}
		err := cfg.Validate()
		assert.Equal(t, ErrInvalidWebhookPort, err)
	})

	t.Run("Invalid Mode", func(t *testing.T) {
		cfg := &Config{
			AppID:     "cli_test",
			AppSecret: "secret",
			Mode:      ConnectionMode("invalid"),
		}
		err := cfg.Validate()
		assert.Equal(t, ErrInvalidMode, err)
	})

	t.Run("Valid Config", func(t *testing.T) {
		cfg := &Config{
			AppID:       "cli_test",
			AppSecret:   "secret",
			Mode:        ModeDual,
			WebSocketURL: "wss://test.com/ws",
			WebhookPort: 8080,
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}

func TestConnectionMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     ConnectionMode
		expected string
	}{
		{"WebSocket", ModeWebSocket, "websocket"},
		{"Webhook", ModeWebhook, "webhook"},
		{"Dual", ModeDual, "dual"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.mode))
		})
	}
}

func TestOutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   OutputFormat
		expected string
	}{
		{"Auto", OutputFormatAuto, "auto"},
		{"Text", OutputFormatText, "text"},
		{"Markdown", OutputFormatMarkdown, "markdown"},
		{"Card", OutputFormatCard, "card"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.format))
		})
	}
}

func TestEventTypes(t *testing.T) {
	t.Run("URL Verification Event", func(t *testing.T) {
		eventJSON := `{
			"token": "test-token",
			"type": "url_verification",
			"challenge": "challenge-string"
		}`

		var event Event
		err := json.Unmarshal([]byte(eventJSON), &event)
		require.NoError(t, err)

		assert.Equal(t, "test-token", event.Token)
		assert.Equal(t, "url_verification", event.EventType)
		assert.Equal(t, "challenge-string", event.Challenge)
	})

}

func TestTextMessageContent(t *testing.T) {
	content := TextMessageContent{
		Text: "Hello, World!",
	}

	contentJSON, err := json.Marshal(content)
	require.NoError(t, err)

	var parsed TextMessageContent
	err = json.Unmarshal(contentJSON, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "Hello, World!", parsed.Text)
}

func TestIncomingMessage(t *testing.T) {
	msg := IncomingMessage{
		ChatID:    "oc_xxx",
		MessageID: "om_xxx",
		UserID:    "ou_xxx",
		UserName:  "Test User",
		Content:   "Hello",
		Timestamp: time.Now(),
	}

	assert.Equal(t, "oc_xxx", msg.ChatID)
	assert.Equal(t, "ou_xxx", msg.UserID)
	assert.Equal(t, "Hello", msg.Content)
}

func TestCardElement(t *testing.T) {
	card := &Card{
		Header: &CardHeader{
			Title: &CardText{
				Tag:     "plain_text",
				Content: "Card Title",
			},
			Template: "blue",
		},
		Elements: []*CardElement{
			{
				Tag: "div",
				Text: &CardText{
					Tag:     "lark_md",
					Content: "Card content",
				},
			},
		},
	}

	cardJSON, err := json.Marshal(card)
	require.NoError(t, err)

	var parsed Card
	err = json.Unmarshal(cardJSON, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "Card Title", parsed.Header.Title.Content)
	assert.Equal(t, "blue", parsed.Header.Template)
	assert.Len(t, parsed.Elements, 1)
	assert.Equal(t, "div", parsed.Elements[0].Tag)
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(60, 10)

	assert.True(t, rl.Allow())

	stats := rl.Stats()
	assert.Equal(t, 60, stats.RPS)
	assert.Equal(t, 10, stats.Burst)
}

func TestFormatter(t *testing.T) {
	cfg := DefaultConfig()
	formatter := NewFormatter(cfg)

	t.Run("FormatText", func(t *testing.T) {
		text := "Hello, World!"
		result := formatter.FormatText(text)
		assert.Equal(t, text, result)
	})

	t.Run("FormatText with Truncation", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.TruncateLength = 10
		formatter := NewFormatter(cfg)

		text := "This is a very long message that should be truncated"
		result := formatter.FormatText(text)
		assert.Len(t, result, 13) // 10 + "..."
		assert.True(t, strings.HasSuffix(result, "..."))
	})

	t.Run("ShouldUseCard", func(t *testing.T) {
		assert.True(t, formatter.ShouldUseCard("Error: something went wrong"))
		assert.True(t, formatter.ShouldUseCard("```code```"))
		assert.True(t, formatter.ShouldUseCard("| table | header |"))
		assert.False(t, formatter.ShouldUseCard("simple text"))
	})
}

func TestSignatureVerifier(t *testing.T) {
	verifier := NewSignatureVerifier("test-token", "test-secret")

	t.Run("VerifyToken", func(t *testing.T) {
		assert.True(t, verifier.VerifyToken("test-token"))
		assert.False(t, verifier.VerifyToken("wrong-token"))
	})
}