package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessageService is a mock implementation of MessageServiceInterface
type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) CreateMessage(ctx context.Context, sessionID string, req *models.CreateMessageRequest) (*models.Message, error) {
	args := m.Called(ctx, sessionID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageService) GetMessage(ctx context.Context, id string) (*models.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageService) GetSessionMessages(ctx context.Context, sessionID string, page, pageSize int) (*models.ListResponse, error) {
	args := m.Called(ctx, sessionID, page, pageSize)
	return args.Get(0).(*models.ListResponse), args.Error(1)
}

func (m *MockMessageService) GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]*models.Message, error) {
	args := m.Called(ctx, sessionID, limit)
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockMessageService) DeleteMessage(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupMessageTest(method, path string, body interface{}) (*httptest.ResponseRecorder, *http.Request) {
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))

	// Extract ID from path for mux routing
	// Format: /api/v1/sessions/{id}/messages
	parts := strings.Split(path, "/")
	// Find the ID between "sessions" and "messages"
	for i, part := range parts {
		if part == "sessions" && i+2 < len(parts) && parts[i+2] == "messages" {
			id := parts[i+1]
			req = mux.SetURLVars(req, map[string]string{"id": id})
			break
		} else if part == "messages" && i+1 < len(parts) && parts[i+1] == "recent" {
			// For /sessions/{id}/messages/recent, still get the session id
			if i > 0 && parts[i-2] == "sessions" {
				id := parts[i-1]
				req = mux.SetURLVars(req, map[string]string{"id": id})
				break
			}
		} else if part == "messages" && i+1 < len(parts) {
			// For /messages/{id}
			id := parts[i+1]
			req = mux.SetURLVars(req, map[string]string{"id": id})
			break
		}
	}

	rr := httptest.NewRecorder()
	return rr, req
}

func TestMessageHandler_CreateMessage_Success(t *testing.T) {
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	sessionID := "session-123"
	req := models.CreateMessageRequest{
		Role:    "user",
		Content: map[string]interface{}{"text": "Hello"},
	}

	expectedMessage := &models.Message{
		ID:        "msg-123",
		SessionID: sessionID,
		Role:      req.Role,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	mockService.On("CreateMessage", mock.Anything, sessionID, mock.AnythingOfType("*models.CreateMessageRequest")).Return(expectedMessage, nil)

	rr, httpReq := setupMessageTest("POST", "/api/v1/sessions/"+sessionID+"/messages", req)

	handler.CreateMessage(rr, httpReq)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp models.Message
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestMessageHandler_CreateMessage_SessionNotFound(t *testing.T) {
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	sessionID := "nonexistent-session"
	req := models.CreateMessageRequest{
		Role:    "user",
		Content: map[string]interface{}{"text": "Hello"},
	}

	mockService.On("CreateMessage", mock.Anything, sessionID, mock.AnythingOfType("*models.CreateMessageRequest")).Return(nil, services.ErrSessionNotFound)

	rr, httpReq := setupMessageTest("POST", "/api/v1/sessions/"+sessionID+"/messages", req)

	handler.CreateMessage(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "session not found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestMessageHandler_GetMessage_Success(t *testing.T) {
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	messageID := "msg-123"
	expectedMessage := &models.Message{
		ID:        messageID,
		SessionID: "session-123",
		Role:      "user",
		Content:   map[string]interface{}{"text": "Hello"},
		CreatedAt: time.Now(),
	}

	mockService.On("GetMessage", mock.Anything, messageID).Return(expectedMessage, nil)

	rr, httpReq := setupMessageTest("GET", "/api/v1/messages/"+messageID, nil)

	handler.GetMessage(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Message
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestMessageHandler_GetMessage_NotFound(t *testing.T) {
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	messageID := "nonexistent-msg"
	mockService.On("GetMessage", mock.Anything, messageID).Return(nil, services.ErrMessageNotFound)

	rr, httpReq := setupMessageTest("GET", "/api/v1/messages/"+messageID, nil)

	handler.GetMessage(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "message not found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestMessageHandler_DeleteMessage_Success(t *testing.T) {
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	messageID := "msg-123"
	mockService.On("DeleteMessage", mock.Anything, messageID).Return(nil)

	rr, httpReq := setupMessageTest("DELETE", "/api/v1/messages/"+messageID, nil)

	handler.DeleteMessage(rr, httpReq)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	mockService.AssertExpectations(t)
}

func TestMessageHandler_GetRecentMessages_Success(t *testing.T) {
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	sessionID := "session-123"
	limit := 5
	expectedMessages := []*models.Message{
		{ID: "msg-1", SessionID: sessionID, Role: "user"},
		{ID: "msg-2", SessionID: sessionID, Role: "assistant"},
	}

	mockService.On("GetRecentMessages", mock.Anything, sessionID, limit).Return(expectedMessages, nil)

	rr, httpReq := setupMessageTest("GET", "/api/v1/sessions/"+sessionID+"/messages/recent?limit=5", nil)

	handler.GetRecentMessages(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []models.Message
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)

	mockService.AssertExpectations(t)
}

func TestMessageHandler_CreateMessage_InvalidBody(t *testing.T) {
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	sessionID := "session-123"
	rr, httpReq := setupMessageTest("POST", "/api/v1/sessions/"+sessionID+"/messages", "invalid json")

	handler.CreateMessage(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request body", resp["error"])
}

func TestMessageHandler_DeleteMessage_NotFound(t *testing.T) {
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)

	messageID := "nonexistent-msg"
	mockService.On("DeleteMessage", mock.Anything, messageID).Return(services.ErrMessageNotFound)

	rr, httpReq := setupMessageTest("DELETE", "/api/v1/messages/"+messageID, nil)

	handler.DeleteMessage(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "message not found", resp["error"])

	mockService.AssertExpectations(t)
}
