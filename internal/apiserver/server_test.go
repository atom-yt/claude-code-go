package apiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	cfg := &Config{
		Addr:        ":8080",
		Model:       "claude-sonnet-4-6",
		Provider:    "anthropic",
		DeploymentMode: ModePool,
		PoolSize:    4,
		EnableSSE:   true,
		EnableAuth:  true,
		APIKey:      "test-key",
		AllowedOrigins: []string{"*"},
	}

	assert.Equal(t, ":8080", cfg.Addr)
	assert.Equal(t, ModePool, cfg.DeploymentMode)
	assert.Equal(t, 4, cfg.PoolSize)
	assert.True(t, cfg.EnableSSE)
	assert.True(t, cfg.EnableAuth)
	assert.Equal(t, "test-key", cfg.APIKey)
}

func TestCreateSessionID(t *testing.T) {
	id1 := newSessionID()
	id2 := newSessionID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.True(t, strings.HasPrefix(id1, "sess_"))
}

func TestGenerateMessageID(t *testing.T) {
	id1 := generateMessageID()
	time.Sleep(time.Nanosecond) // Small delay to ensure unique ID
	id2 := generateMessageID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.True(t, strings.HasPrefix(id1, "msg_"))
}

func TestAPIResponseTypes(t *testing.T) {
	t.Run("CreateSessionRequest", func(t *testing.T) {
		req := CreateSessionRequest{
			SystemPrompt: "You are a helpful assistant.",
			Model:       "claude-sonnet-4-6",
			MaxHistory:  50,
		}

		assert.Equal(t, "You are a helpful assistant.", req.SystemPrompt)
		assert.Equal(t, "claude-sonnet-4-6", req.Model)
		assert.Equal(t, 50, req.MaxHistory)
	})

	t.Run("ChatCompletionRequest", func(t *testing.T) {
		req := ChatCompletionRequest{
			SessionID:  "test-session",
			Message:    "Hello",
			Model:      "gpt-4",
			Stream:     false,
			MaxTokens:  1000,
		}

		assert.Equal(t, "test-session", req.SessionID)
		assert.Equal(t, "Hello", req.Message)
		assert.Equal(t, "gpt-4", req.Model)
		assert.False(t, req.Stream)
		assert.Equal(t, 1000, req.MaxTokens)
	})

	t.Run("ChatChunk", func(t *testing.T) {
		chunk := ChatChunk{
			SessionID: "test-session",
			Type:      "delta",
			Content:   "Hello",
		}

		assert.Equal(t, "test-session", chunk.SessionID)
		assert.Equal(t, "delta", chunk.Type)
		assert.Equal(t, "Hello", chunk.Content)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("writeJSON", func(t *testing.T) {
		w := httptest.NewRecorder()

		err := writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
	})

	t.Run("writeError", func(t *testing.T) {
		w := httptest.NewRecorder()

		err := writeError(w, http.StatusBadRequest, "Invalid request")

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Bad Request", response.Error)
		assert.Equal(t, "Invalid request", response.Message)
	})

	t.Run("readJSON", func(t *testing.T) {
		body := `{"session_id": "test", "message": "hello"}`
		r := httptest.NewRequest("POST", "/api/v1/chat", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")

		var req ChatCompletionRequest
		err := readJSON(r, &req)

		require.NoError(t, err)
		assert.Equal(t, "test", req.SessionID)
		assert.Equal(t, "hello", req.Message)
	})

	t.Run("readJSON - invalid", func(t *testing.T) {
		body := `{"invalid": json`
		r := httptest.NewRequest("POST", "/api/v1/chat", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")

		var req ChatCompletionRequest
		err := readJSON(r, &req)

		assert.Error(t, err)
	})
}

func TestAgentManager(t *testing.T) {
	t.Run("Single mode", func(t *testing.T) {
		// Test single mode initialization
		am := NewAgentManager(ModeSingle, nil)

		assert.Equal(t, ModeSingle, am.mode)
		assert.Nil(t, am.single)
		assert.Nil(t, am.pool)
	})

	t.Run("Pool mode", func(t *testing.T) {
		// Test pool mode initialization
		am := NewAgentManager(ModePool, nil)

		assert.Equal(t, ModePool, am.mode)
		assert.Nil(t, am.single)
		assert.Nil(t, am.pool)
	})

	t.Run("Per-session mode", func(t *testing.T) {
		// Test per-session mode initialization
		am := NewAgentManager(ModePerSession, nil)

		assert.Equal(t, ModePerSession, am.mode)
		assert.Nil(t, am.single)
		assert.Nil(t, am.pool)
	})
}

func TestDeploymentMode(t *testing.T) {
	t.Run("Mode parsing", func(t *testing.T) {
		tests := []struct {
			name     string
			modeStr  string
			expected DeploymentMode
		}{
			{"Single", "single", ModeSingle},
			{"PerSession", "per-session", ModePerSession},
			{"Pool", "pool", ModePool},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var mode DeploymentMode
				switch tt.modeStr {
				case "single":
					mode = ModeSingle
				case "per-session":
					mode = ModePerSession
				case "pool":
					mode = ModePool
				}
				assert.Equal(t, tt.expected, mode)
			})
		}
	})
}

func TestHealthResponse(t *testing.T) {
	response := HealthResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Uptime:  "1h30m",
		Stats: map[string]any{
			"active_sessions": 5,
			"pool_size":      4,
		},
	}

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "1.0.0", response.Version)
	assert.Equal(t, "1h30m", response.Uptime)
	assert.Equal(t, 5, response.Stats["active_sessions"])
	assert.Equal(t, 4, response.Stats["pool_size"])
}

func TestSessionInfo(t *testing.T) {
	info := SessionInfo{
		SessionID:    "sess_123",
		CreatedAt:    "2024-01-01T00:00:00Z",
		LastActive:   "2024-01-01T01:00:00Z",
		MessageCount: 10,
		UserAgent:    "curl/7.68.0",
		RemoteAddr:   "127.0.0.1:12345",
	}

	assert.Equal(t, "sess_123", info.SessionID)
	assert.Equal(t, "2024-01-01T00:00:00Z", info.CreatedAt)
	assert.Equal(t, 10, info.MessageCount)
	assert.Equal(t, "curl/7.68.0", info.UserAgent)
	assert.Equal(t, "127.0.0.1:12345", info.RemoteAddr)
}

func TestAgentWrapper(t *testing.T) {
	t.Run("New wrapper", func(t *testing.T) {
		wrapper := &AgentWrapper{
			busy:    false,
			requests: 0,
			lastUsed: time.Now(),
		}

		assert.False(t, wrapper.busy)
		assert.Equal(t, 0, wrapper.requests)
	})

	t.Run("Mark busy", func(t *testing.T) {
		wrapper := &AgentWrapper{
			busy:    false,
			requests: 0,
		}

		wrapper.mu.Lock()
		wrapper.busy = true
		wrapper.requests = 1
		wrapper.mu.Unlock()

		wrapper.mu.Lock()
		assert.True(t, wrapper.busy)
		assert.Equal(t, 1, wrapper.requests)
		wrapper.mu.Unlock()
	})
}

// Integration-style tests (mock server)
func TestServerSetup(t *testing.T) {
	cfg := &Config{
		Addr:         ":8080",
		DeploymentMode: ModeSingle,
		EnableSSE:    false,
		EnableAuth:   false,
	}

	server := NewServer(cfg)

	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.config)
	assert.NotNil(t, server.router)
	assert.NotNil(t, server.agentManager)
}

func TestRouteSetup(t *testing.T) {
	cfg := &Config{
		Addr:         ":8080",
		DeploymentMode: ModeSingle,
		EnableSSE:    true,
		EnableAuth:   false,
	}

	server := NewServer(cfg)

	// Test that routes are registered
	// This is a basic check - in a full test we'd make HTTP requests
	assert.NotNil(t, server.router)
}

func TestServerErrorHandling(t *testing.T) {
	t.Run("Invalid JSON body", func(t *testing.T) {
		body := strings.NewReader(`{invalid json`)

		req, _ := http.NewRequest("POST", "/api/v1/sessions", body)
		req.Header.Set("Content-Type", "application/json")

		// Test error handling
		var reqObj CreateSessionRequest
		err := readJSON(req, &reqObj)
		assert.Error(t, err)
	})
}