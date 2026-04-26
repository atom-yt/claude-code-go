package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
	"github.com/atom-yt/claude-code-go/pkg/agent"
	"github.com/stretchr/testify/assert"
)

func TestChatHandler_HandleChat_Unauthorized(t *testing.T) {
	factory := agent.NewConfigFactory("test-key", "", "anthropic", "claude-3-5-sonnet-20241022")
	handler := NewChatHandler(factory)

	chatReq := ChatRequest{
		SessionID: "session-123",
		Message:   "Hello, world!",
		Stream:   true,
	}

	reqBody, _ := json.Marshal(chatReq)
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader(reqBody))

	rr := httptest.NewRecorder()

	handler.HandleChat(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestChatHandler_HandleChat_StreamingRequired(t *testing.T) {
	userID := "user-123"
	factory := agent.NewConfigFactory("test-key", "", "anthropic", "claude-3-5-sonnet-20241022")
	handler := NewChatHandler(factory)

	chatReq := ChatRequest{
		SessionID: "session-123",
		Message:   "Hello, world!",
		Stream:   false,
	}

	reqBody, _ := json.Marshal(chatReq)
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))

	rr := httptest.NewRecorder()

	handler.HandleChat(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "only streaming mode is supported", resp["error"])
}

func TestChatHandler_HandleWebSocket_Unauthorized(t *testing.T) {
	factory := agent.NewConfigFactory("test-key", "", "anthropic", "claude-3-5-sonnet-20241022")
	handler := NewChatHandler(factory)

	req := httptest.NewRequest("GET", "/api/v1/chat/ws?session_id=session-123", nil)
	rr := httptest.NewRecorder()

	handler.HandleWebSocket(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestChatHandler_HandleWebSocket_NoSessionID(t *testing.T) {
	userID := "user-123"
	factory := agent.NewConfigFactory("test-key", "", "anthropic", "claude-3-5-sonnet-20241022")
	handler := NewChatHandler(factory)

	req := httptest.NewRequest("GET", "/api/v1/chat/ws", nil)
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))

	rr := httptest.NewRecorder()

	handler.HandleWebSocket(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "session_id required", resp["error"])
}

func TestChatHandler_HandleChat_InvalidBody(t *testing.T) {
	userID := "user-123"
	factory := agent.NewConfigFactory("test-key", "", "anthropic", "claude-3-5-sonnet-20241022")
	handler := NewChatHandler(factory)

	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader([]byte("invalid json")))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))

	rr := httptest.NewRecorder()

	handler.HandleChat(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request body", resp["error"])
}

func TestChatHandler_HandleChat_SSEHeaders(t *testing.T) {
	userID := "user-123"

	// Create a mock agent factory that returns a mock agent
	factory := agent.NewConfigFactory("test-key", "", "anthropic", "claude-3-5-sonnet-20241022")
	handler := NewChatHandler(factory)

	chatReq := ChatRequest{
		SessionID: "session-123",
		Message:   "Hello, world!",
		Stream:   true,
	}

	reqBody, _ := json.Marshal(chatReq)
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))

	rr := httptest.NewRecorder()

	// This will fail at agent creation, but we can still check headers before that
	handler.HandleChat(rr, req)

	// Check that SSE headers were set
	assert.Equal(t, "text/event-stream", rr.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rr.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", rr.Header().Get("Connection"))
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}
