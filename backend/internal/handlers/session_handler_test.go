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

	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSessionService is a mock implementation of SessionServiceInterface
type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) CreateSession(ctx context.Context, userID string, req *models.CreateSessionRequest) (*models.Session, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionService) GetSession(ctx context.Context, id string) (*models.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionService) GetUserSessions(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error) {
	args := m.Called(ctx, userID, page, pageSize)
	return args.Get(0).(*models.ListResponse), args.Error(1)
}

func (m *MockSessionService) UpdateSession(ctx context.Context, id string, req *models.UpdateSessionRequest) (*models.Session, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionService) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionService) ArchiveSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionService) GetActiveSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Session), args.Error(1)
}

func setupSessionTest(userID string, method, path string, body interface{}) (*httptest.ResponseRecorder, *http.Request) {
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))

	if userID != "" {
		ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)
		req = req.WithContext(ctx)
	}

	// Extract ID from path
	if strings.Contains(path, "/sessions/") && !strings.HasSuffix(path, "/sessions") && !strings.Contains(path, "/sessions/active") {
		parts := strings.Split(path, "/")
		// Find the session ID (comes after /sessions/)
		for i, part := range parts {
			if part == "sessions" && i+1 < len(parts) {
				id := parts[i+1]
				// Skip special paths like "archive"
				if id != "active" && id != "archive" {
					req = mux.SetURLVars(req, map[string]string{"id": id})
				}
				break
			}
		}
	}

	rr := httptest.NewRecorder()
	return rr, req
}

func TestSessionHandler_CreateSession_Success(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	agentID := "agent-123"
	userID := "user-123"
	req := models.CreateSessionRequest{
		Title:   "Test Session",
		AgentID: agentID,
	}

	expectedSession := &models.Session{
		ID:        "session-123",
		UserID:    userID,
		AgentID:   agentID,
		Title:     req.Title,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService.On("CreateSession", mock.Anything, userID, mock.AnythingOfType("*models.CreateSessionRequest")).Return(expectedSession, nil)

	rr, httpReq := setupSessionTest(userID, "POST", "/api/v1/sessions", req)

	handler.CreateSession(rr, httpReq)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp models.Session
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedSession.ID, resp.ID)
	assert.Equal(t, expectedSession.Title, resp.Title)

	mockService.AssertExpectations(t)
}

func TestSessionHandler_CreateSession_Unauthorized(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	req := models.CreateSessionRequest{
		Title:   "Test Session",
		AgentID: "agent-123",
	}

	rr, httpReq := setupSessionTest("", "POST", "/api/v1/sessions", req)

	handler.CreateSession(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestSessionHandler_CreateSession_AgentNotFound(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	userID := "user-123"
	req := models.CreateSessionRequest{
		Title:   "Test Session",
		AgentID: "nonexistent-agent",
	}

	mockService.On("CreateSession", mock.Anything, userID, mock.AnythingOfType("*models.CreateSessionRequest")).Return(nil, services.ErrAgentNotFound)

	rr, httpReq := setupSessionTest(userID, "POST", "/api/v1/sessions", req)

	handler.CreateSession(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "agent not found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestSessionHandler_GetSession_Success(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	sessionID := "session-123"
	expectedSession := &models.Session{
		ID:        sessionID,
		UserID:    "user-123",
		Title:     "Test Session",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	mockService.On("GetSession", mock.Anything, sessionID).Return(expectedSession, nil)

	rr, httpReq := setupSessionTest("", "GET", "/api/v1/sessions/"+sessionID, nil)

	handler.GetSession(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Session
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedSession.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestSessionHandler_GetSession_NotFound(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	sessionID := "nonexistent-session"
	mockService.On("GetSession", mock.Anything, sessionID).Return(nil, services.ErrSessionNotFound)

	rr, httpReq := setupSessionTest("", "GET", "/api/v1/sessions/"+sessionID, nil)

	handler.GetSession(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "session not found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestSessionHandler_DeleteSession_Success(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	sessionID := "session-123"
	mockService.On("DeleteSession", mock.Anything, sessionID).Return(nil)

	rr, httpReq := setupSessionTest("", "DELETE", "/api/v1/sessions/"+sessionID, nil)

	handler.DeleteSession(rr, httpReq)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	mockService.AssertExpectations(t)
}

func TestSessionHandler_GetUserSessions_Success(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	userID := "user-123"
	expectedSessions := []*models.Session{
		{ID: "session-1", UserID: userID, Title: "Session 1"},
		{ID: "session-2", UserID: userID, Title: "Session 2"},
	}

	expectedResponse := &models.ListResponse{
		Items:      expectedSessions,
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockService.On("GetUserSessions", mock.Anything, userID, 1, 10).Return(expectedResponse, nil)

	rr, httpReq := setupSessionTest(userID, "GET", "/api/v1/sessions?page=1&page_size=10", nil)

	handler.GetUserSessions(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)

	mockService.AssertExpectations(t)
}

func TestSessionHandler_GetUserSessions_Unauthorized(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	rr, httpReq := setupSessionTest("", "GET", "/api/v1/sessions", nil)

	handler.GetUserSessions(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestSessionHandler_UpdateSession_Success(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	sessionID := "session-123"
	newTitle := "Updated Session"
	req := models.UpdateSessionRequest{
		Title: &newTitle,
	}

	expectedSession := &models.Session{
		ID:        sessionID,
		Title:     newTitle,
		Status:    "active",
		UpdatedAt: time.Now(),
	}

	mockService.On("UpdateSession", mock.Anything, sessionID, mock.AnythingOfType("*models.UpdateSessionRequest")).Return(expectedSession, nil)

	rr, httpReq := setupSessionTest("", "PUT", "/api/v1/sessions/"+sessionID, req)

	handler.UpdateSession(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Session
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, newTitle, resp.Title)

	mockService.AssertExpectations(t)
}

func TestSessionHandler_UpdateSession_InvalidBody(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	sessionID := "session-123"
	rr, httpReq := setupSessionTest("", "PUT", "/api/v1/sessions/"+sessionID, "invalid json")

	handler.UpdateSession(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request body", resp["error"])
}

func TestSessionHandler_ArchiveSession_Success(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	sessionID := "session-123"
	mockService.On("ArchiveSession", mock.Anything, sessionID).Return(nil)

	rr, httpReq := setupSessionTest("", "POST", "/api/v1/sessions/"+sessionID+"/archive", nil)

	handler.ArchiveSession(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "archived", resp["status"])

	mockService.AssertExpectations(t)
}

func TestSessionHandler_GetActiveSessions_Success(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	userID := "user-123"
	expectedSessions := []*models.Session{
		{ID: "session-1", UserID: userID, Title: "Active 1", Status: "active"},
		{ID: "session-2", UserID: userID, Title: "Active 2", Status: "active"},
	}

	mockService.On("GetActiveSessions", mock.Anything, userID).Return(expectedSessions, nil)

	rr, httpReq := setupSessionTest(userID, "GET", "/api/v1/sessions/active", nil)

	handler.GetActiveSessions(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []models.Session
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)

	mockService.AssertExpectations(t)
}

func TestSessionHandler_GetActiveSessions_Unauthorized(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	rr, httpReq := setupSessionTest("", "GET", "/api/v1/sessions/active", nil)

	handler.GetActiveSessions(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestSessionHandler_CreateSession_InvalidBody(t *testing.T) {
	mockService := new(MockSessionService)
	handler := NewSessionHandler(mockService)

	userID := "user-123"
	rr, httpReq := setupSessionTest(userID, "POST", "/api/v1/sessions", "invalid json")

	handler.CreateSession(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request body", resp["error"])
}
