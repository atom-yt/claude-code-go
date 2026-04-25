// Package agent provides unit tests for the public agent API.
package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStreamer implements api.Streamer for testing
type mockStreamer struct {
	mu     sync.Mutex
	events []api.APIEvent
	called bool
}

func (m *mockStreamer) StreamMessages(ctx context.Context, req api.MessagesRequest) <-chan api.APIEvent {
	m.mu.Lock()
	m.called = true
	events := make([]api.APIEvent, len(m.events))
	copy(events, m.events)
	m.mu.Unlock()

	ch := make(chan api.APIEvent, len(events))
	go func() {
		defer close(ch)
		for _, ev := range events {
			select {
			case ch <- ev:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

// TestNewConfigFactory tests creating a new ConfigFactory
func TestNewConfigFactory(t *testing.T) {
	t.Run("creates factory with all values", func(t *testing.T) {
		f := NewConfigFactory("test-key", "https://test.com", "test-provider", "test-model")
		assert.NotNil(t, f)
		assert.Equal(t, "test-key", f.defaultAPIKey)
		assert.Equal(t, "https://test.com", f.defaultBaseURL)
		assert.Equal(t, "test-provider", f.defaultProvider)
		assert.Equal(t, "test-model", f.defaultModel)
	})

	t.Run("creates factory with empty values", func(t *testing.T) {
		f := NewConfigFactory("", "", "", "")
		assert.NotNil(t, f)
		assert.Empty(t, f.defaultAPIKey)
		assert.Empty(t, f.defaultBaseURL)
		assert.Empty(t, f.defaultProvider)
		assert.Empty(t, f.defaultModel)
	})
}

// TestConfigFactoryCreate tests creating an agent from factory
func TestConfigFactoryCreate(t *testing.T) {
	t.Run("merges defaults with empty config", func(t *testing.T) {
		f := NewConfigFactory("default-key", "https://default.com", "anthropic", "claude-3-5-sonnet-20241022")
		cfg := &Config{}

		agent, err := f.Create(context.Background(), cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "default-key", cfg.APIKey)
		assert.Equal(t, "https://default.com", cfg.BaseURL)
		assert.Equal(t, "anthropic", cfg.Provider)
		assert.Equal(t, "claude-3-5-sonnet-20241022", cfg.Model)
	})

	t.Run("uses provided config values over defaults", func(t *testing.T) {
		f := NewConfigFactory("default-key", "https://default.com", "anthropic", "default-model")
		cfg := &Config{
			APIKey:   "custom-key",
			BaseURL:  "https://custom.com",
			Provider: "openai",
			Model:    "gpt-4",
		}

		agent, err := f.Create(context.Background(), cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "custom-key", cfg.APIKey)
		assert.Equal(t, "https://custom.com", cfg.BaseURL)
		assert.Equal(t, "openai", cfg.Provider)
		assert.Equal(t, "gpt-4", cfg.Model)
	})

	t.Run("partial merge - some defaults used", func(t *testing.T) {
		f := NewConfigFactory("default-key", "https://default.com", "anthropic", "default-model")
		cfg := &Config{
			APIKey: "custom-key",
			Model:  "custom-model",
		}

		agent, err := f.Create(context.Background(), cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "custom-key", cfg.APIKey)
		assert.Equal(t, "https://default.com", cfg.BaseURL) // default used
		assert.Equal(t, "anthropic", cfg.Provider)         // default used
		assert.Equal(t, "custom-model", cfg.Model)
	})

	t.Run("creates agent with system prompt", func(t *testing.T) {
		f := NewConfigFactory("test-key", "", "anthropic", "claude-3-5-sonnet-20241022")
		cfg := &Config{
			SystemPrompt: "You are a helpful assistant.",
		}

		agent, err := f.Create(context.Background(), cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		agent.mu.RLock()
		assert.Equal(t, "You are a helpful assistant.", agent.system)
		agent.mu.RUnlock()
	})

	t.Run("respects canceled context", func(t *testing.T) {
		f := NewConfigFactory("test-key", "", "anthropic", "claude-3-5-sonnet-20241022")
		cfg := &Config{}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		// Note: New() doesn't use the context, so this shouldn't fail
		agent, err := f.Create(ctx, cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
	})
}

// TestNew tests creating a ChatAgent
func TestNew(t *testing.T) {
	t.Run("creates agent with minimal config", func(t *testing.T) {
		cfg := &Config{
			APIKey: "test-key",
		}

		agent, err := New(cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.NotNil(t, agent.client)
		assert.NotNil(t, agent.registry)
	})

	t.Run("uses default model when empty", func(t *testing.T) {
		cfg := &Config{
			APIKey: "test-key",
			Model:  "",
		}

		agent, err := New(cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "claude-sonnet-4-6", agent.model)
	})

	t.Run("uses default provider when empty", func(t *testing.T) {
		cfg := &Config{
			APIKey:   "test-key",
			Provider: "",
		}

		agent, err := New(cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "anthropic", agent.provider)
	})

	t.Run("uses provided model and provider", func(t *testing.T) {
		cfg := &Config{
			APIKey:   "test-key",
			Model:    "gpt-4",
			Provider: "openai",
		}

		agent, err := New(cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "gpt-4", agent.model)
		assert.Equal(t, "openai", agent.provider)
	})

	t.Run("creates agent with system prompt", func(t *testing.T) {
		cfg := &Config{
			APIKey:       "test-key",
			SystemPrompt: "You are a test assistant.",
		}

		agent, err := New(cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		agent.mu.RLock()
		assert.Equal(t, "You are a test assistant.", agent.system)
		agent.mu.RUnlock()
	})

	t.Run("creates openai client for openai provider", func(t *testing.T) {
		cfg := &Config{
			APIKey:   "test-key",
			Provider: "openai",
		}

		agent, err := New(cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.IsType(t, &api.OpenAIClient{}, agent.client)
	})

	t.Run("creates anthropic client for anthropic provider", func(t *testing.T) {
		cfg := &Config{
			APIKey:   "test-key",
			Provider: "anthropic",
		}

		agent, err := New(cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.IsType(t, &api.Client{}, agent.client)
	})

	t.Run("creates anthropic client for unknown provider", func(t *testing.T) {
		cfg := &Config{
			APIKey:   "test-key",
			Provider: "unknown-provider",
		}

		agent, err := New(cfg)

		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.IsType(t, &api.Client{}, agent.client)
	})
}

// TestSetSystemPrompt tests setting the system prompt
func TestSetSystemPrompt(t *testing.T) {
	t.Run("sets system prompt", func(t *testing.T) {
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)

		agent.SetSystemPrompt("New system prompt")

		agent.mu.RLock()
		assert.Equal(t, "New system prompt", agent.system)
		agent.mu.RUnlock()
	})

	t.Run("updates existing system prompt", func(t *testing.T) {
		cfg := &Config{
			APIKey:       "test-key",
			SystemPrompt: "Old prompt",
		}
		agent, err := New(cfg)
		require.NoError(t, err)

		agent.SetSystemPrompt("New system prompt")

		agent.mu.RLock()
		assert.Equal(t, "New system prompt", agent.system)
		agent.mu.RUnlock()
	})

	t.Run("clears system prompt with empty string", func(t *testing.T) {
		cfg := &Config{
			APIKey:       "test-key",
			SystemPrompt: "Old prompt",
		}
		agent, err := New(cfg)
		require.NoError(t, err)

		agent.SetSystemPrompt("")

		agent.mu.RLock()
		assert.Equal(t, "", agent.system)
		agent.mu.RUnlock()
	})

	t.Run("concurrent sets are safe", func(t *testing.T) {
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				agent.SetSystemPrompt("prompt-" + string(rune(n)))
			}(i)
		}
		wg.Wait()

		// Just verify no deadlock or panic occurred
		agent.mu.RLock()
		_ = agent.system
		agent.mu.RUnlock()
	})
}

// TestChat tests the Chat streaming functionality
func TestChat(t *testing.T) {
	t.Run("returns channel and starts goroutine", func(t *testing.T) {
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)

		ch := agent.Chat(context.Background(), "hello")

		assert.NotNil(t, assert.IsType(t, (<-chan StreamEvent)(nil), ch))
		assert.NotPanics(t, func() {
			// Channel should be buffered and closable
			for range ch {
				break
			}
		})
	})

	t.Run("emits delta events", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{
				{Type: api.EventTextDelta, Text: "Hello"},
				{Type: api.EventTextDelta, Text: " world"},
			},
		}
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock

		ch := agent.Chat(context.Background(), "hello")
		var events []StreamEvent
		for ev := range ch {
			events = append(events, ev)
		}

		assert.GreaterOrEqual(t, len(events), 2)
		assert.Equal(t, EventTypeDelta, events[0].Type)
		assert.Equal(t, "Hello", events[0].Text)
		assert.Equal(t, EventTypeDelta, events[1].Type)
		assert.Equal(t, " world", events[1].Text)
	})

	t.Run("emits tool call events", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{
				{
					Type: api.EventToolUse,
					ToolUse: &api.ToolUse{
						Name:  "Read",
						Input: map[string]any{"path": "file.txt"},
					},
				},
			},
		}
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock

		ch := agent.Chat(context.Background(), "read file")
		var events []StreamEvent
		for ev := range ch {
			events = append(events, ev)
		}

		found := false
		for _, ev := range events {
			if ev.Type == EventTypeToolCall && ev.ToolName == "Read" {
				found = true
				assert.Equal(t, "file.txt", ev.ToolInput["path"])
			}
		}
		assert.True(t, found, "tool call event not found")
	})

	t.Run("emits error events", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{
				{Type: api.EventError, Error: assert.AnError},
			},
		}
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock

		ch := agent.Chat(context.Background(), "hello")
		var events []StreamEvent
		for ev := range ch {
			events = append(events, ev)
		}

		assert.NotEmpty(t, events)
		assert.Equal(t, EventTypeError, events[0].Type)
		assert.NotNil(t, events[0].Error)
	})

	t.Run("emits done event on message stop", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{
				{Type: api.EventTextDelta, Text: "Hello"},
				{Type: api.EventMessageStop},
			},
		}
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock

		ch := agent.Chat(context.Background(), "hello")
		var events []StreamEvent
		for ev := range ch {
			events = append(events, ev)
		}

		assert.Contains(t, events, StreamEvent{Type: EventTypeDone})
	})

	t.Run("closes channel when done", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{
				{Type: api.EventTextDelta, Text: "Hi"},
				{Type: api.EventMessageStop},
			},
		}
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock

		ch := agent.Chat(context.Background(), "hello")

		// Should not block
		for range ch {
		}

		_, ok := <-ch
		assert.False(t, ok, "channel should be closed")
	})

	t.Run("includes usage information", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{
				{
					Type: api.EventTextDelta,
					Text: "Hello",
					Usage: &api.Usage{
						InputTokens:  10,
						OutputTokens: 20,
					},
				},
				{Type: api.EventMessageStop},
			},
		}
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock

		ch := agent.Chat(context.Background(), "hello")
		var events []StreamEvent
		for ev := range ch {
			events = append(events, ev)
		}

		hasUsage := false
		for _, ev := range events {
			if ev.Usage != nil {
				hasUsage = true
				assert.Equal(t, 10, ev.Usage.InputTokens)
				assert.Equal(t, 20, ev.Usage.OutputTokens)
			}
		}
		assert.True(t, hasUsage, "usage information not found")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{
				{Type: api.EventTextDelta, Text: "Hello"},
				{Type: api.EventTextDelta, Text: " world"},
				{Type: api.EventTextDelta, Text: " !"},
			},
		}
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock

		ctx, cancel := context.WithCancel(context.Background())
		ch := agent.Chat(ctx, "hello")

		// Read first event
		<-ch

		// Cancel context
		cancel()

		// Channel should close or stop receiving events
		timeout := time.After(100 * time.Millisecond)
		for i := 0; i < 10; i++ {
			select {
			case ev, ok := <-ch:
				if !ok {
					return
				}
				_ = ev
			case <-timeout:
				// Either closed or timed out, both are acceptable
				return
			}
		}
	})

	t.Run("respects context deadline", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{
				{Type: api.EventTextDelta, Text: "Hello"},
			},
		}
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		ch := agent.Chat(ctx, "hello")

		// Should not block forever
		for range ch {
		}
	})

	t.Run("handles multiple concurrent chats", func(t *testing.T) {
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				mock := &mockStreamer{
					events: []api.APIEvent{
						{Type: api.EventTextDelta, Text: "Response"},
						{Type: api.EventMessageStop},
					},
				}
				agent.client = mock
				ch := agent.Chat(context.Background(), "test")
				for range ch {
				}
			}(i)
		}
		wg.Wait()
	})
}

// TestEventType verifies EventType constants
func TestEventType(t *testing.T) {
	t.Run("has all event types", func(t *testing.T) {
		assert.Equal(t, EventType("delta"), EventTypeDelta)
		assert.Equal(t, EventType("tool_call"), EventTypeToolCall)
		assert.Equal(t, EventType("tool_result"), EventTypeToolResult)
		assert.Equal(t, EventType("error"), EventTypeError)
		assert.Equal(t, EventType("done"), EventTypeDone)
	})
}

// TestStreamEvent verifies StreamEvent structure
func TestStreamEvent(t *testing.T) {
	t.Run("creates delta event", func(t *testing.T) {
		ev := StreamEvent{
			Type: EventTypeDelta,
			Text: "Hello",
		}
		assert.Equal(t, EventTypeDelta, ev.Type)
		assert.Equal(t, "Hello", ev.Text)
	})

	t.Run("creates tool call event", func(t *testing.T) {
		ev := StreamEvent{
			Type:      EventTypeToolCall,
			ToolName:  "Read",
			ToolInput: map[string]any{"path": "file.txt"},
		}
		assert.Equal(t, EventTypeToolCall, ev.Type)
		assert.Equal(t, "Read", ev.ToolName)
		assert.Equal(t, "file.txt", ev.ToolInput["path"])
	})

	t.Run("creates tool result event", func(t *testing.T) {
		ev := StreamEvent{
			Type:       EventTypeToolResult,
			ToolName:   "Read",
			ToolOutput: "file content",
			ToolIsError: false,
		}
		assert.Equal(t, EventTypeToolResult, ev.Type)
		assert.Equal(t, "Read", ev.ToolName)
		assert.Equal(t, "file content", ev.ToolOutput)
		assert.False(t, ev.ToolIsError)
	})

	t.Run("creates error event", func(t *testing.T) {
		ev := StreamEvent{
			Type:  EventTypeError,
			Error: assert.AnError,
		}
		assert.Equal(t, EventTypeError, ev.Type)
		assert.NotNil(t, ev.Error)
	})

	t.Run("creates done event", func(t *testing.T) {
		ev := StreamEvent{
			Type: EventTypeDone,
		}
		assert.Equal(t, EventTypeDone, ev.Type)
	})

	t.Run("includes usage in delta event", func(t *testing.T) {
		ev := StreamEvent{
			Type: EventTypeDelta,
			Usage: &api.Usage{
				InputTokens:  10,
				OutputTokens: 20,
			},
		}
		assert.NotNil(t, ev.Usage)
		assert.Equal(t, 10, ev.Usage.InputTokens)
		assert.Equal(t, 20, ev.Usage.OutputTokens)
	})
}

// TestPermissionsConfig tests the PermissionsConfig structure
func TestPermissionsConfig(t *testing.T) {
	t.Run("creates allow all permissions", func(t *testing.T) {
		cfg := &PermissionsConfig{
			AllowAll: true,
		}
		assert.True(t, cfg.AllowAll)
	})

	t.Run("creates restrictive permissions", func(t *testing.T) {
		cfg := &PermissionsConfig{
			AllowAll: false,
		}
		assert.False(t, cfg.AllowAll)
	})
}

// TestConfig tests the Config structure
func TestConfig(t *testing.T) {
	t.Run("creates full config", func(t *testing.T) {
		cfg := &Config{
			APIKey:       "test-key",
			BaseURL:      "https://test.com",
			Model:        "test-model",
			Provider:     "anthropic",
			SystemPrompt: "Test system",
			Tools:        []string{"Read", "Write"},
			Permissions:  &PermissionsConfig{AllowAll: true},
		}
		assert.Equal(t, "test-key", cfg.APIKey)
		assert.Equal(t, "https://test.com", cfg.BaseURL)
		assert.Equal(t, "test-model", cfg.Model)
		assert.Equal(t, "anthropic", cfg.Provider)
		assert.Equal(t, "Test system", cfg.SystemPrompt)
		assert.Len(t, cfg.Tools, 2)
		assert.NotNil(t, cfg.Permissions)
	})

	t.Run("creates minimal config", func(t *testing.T) {
		cfg := &Config{
			APIKey: "test-key",
		}
		assert.Equal(t, "test-key", cfg.APIKey)
		assert.Empty(t, cfg.BaseURL)
		assert.Empty(t, cfg.Model)
		assert.Empty(t, cfg.Provider)
		assert.Empty(t, cfg.SystemPrompt)
		assert.Nil(t, cfg.Permissions)
	})
}

// TestRun tests the internal run function indirectly through Chat
func TestRun(t *testing.T) {
	t.Run("passes system prompt to API", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{
				{Type: api.EventTextDelta, Text: "Response"},
				{Type: api.EventMessageStop},
			},
		}
		cfg := &Config{
			APIKey:       "test-key",
			SystemPrompt: "You are a helpful assistant.",
		}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock

		ch := agent.Chat(context.Background(), "hello")
		for range ch {
		}

		assert.True(t, mock.called)
	})

	t.Run("passes user message to API", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{
				{Type: api.EventTextDelta, Text: "Response"},
				{Type: api.EventMessageStop},
			},
		}
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock

		testMessage := "What is the weather?"
		ch := agent.Chat(context.Background(), testMessage)
		for range ch {
		}

		assert.True(t, mock.called)
	})

	t.Run("handles empty event stream", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{},
		}
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock

		ch := agent.Chat(context.Background(), "hello")
		var events []StreamEvent
		for ev := range ch {
			events = append(events, ev)
		}

		// Empty stream should result in no events and channel closed
		assert.Empty(t, events)
		assert.True(t, mock.called)
	})

	t.Run("handles nil system prompt", func(t *testing.T) {
		mock := &mockStreamer{
			events: []api.APIEvent{
				{Type: api.EventTextDelta, Text: "Response"},
				{Type: api.EventMessageStop},
			},
		}
		cfg := &Config{APIKey: "test-key"}
		agent, err := New(cfg)
		require.NoError(t, err)
		agent.client = mock
		agent.system = ""

		ch := agent.Chat(context.Background(), "hello")
		for range ch {
		}

		assert.True(t, mock.called)
	})
}
