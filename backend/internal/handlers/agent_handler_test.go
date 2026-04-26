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

// MockAgentService is a mock implementation of AgentServiceInterface
type MockAgentService struct {
	mock.Mock
}

func (m *MockAgentService) CreateAgent(ctx context.Context, userID string, req *models.CreateAgentRequest) (*models.Agent, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func (m *MockAgentService) GetAgent(ctx context.Context, id string) (*models.Agent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func (m *MockAgentService) GetUserAgents(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error) {
	args := m.Called(ctx, userID, page, pageSize)
	return args.Get(0).(*models.ListResponse), args.Error(1)
}

func (m *MockAgentService) ListAgents(ctx context.Context, page, pageSize int) (*models.ListResponse, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).(*models.ListResponse), args.Error(1)
}

func (m *MockAgentService) UpdateAgent(ctx context.Context, id string, req *models.UpdateAgentRequest) (*models.Agent, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func (m *MockAgentService) DeleteAgent(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentService) GetDefaultAgent(ctx context.Context, userID string) (*models.Agent, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func setupAgentTest(userID string, method, path string, body interface{}) (*httptest.ResponseRecorder, *http.Request) {
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))

	if userID != "" {
		ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)
		req = req.WithContext(ctx)
	}

	// Extract ID from path
	if strings.Contains(path, "/agents/") && !strings.Contains(path, "/agents/list") && !strings.Contains(path, "/agents/default") {
		parts := strings.Split(path, "/")
		// Find to agent ID (comes after /agents/)
		for i, part := range parts {
			if part == "agents" && i+1 < len(parts) {
				id := parts[i+1]
				// Skip special paths like "list" and "default"
				if id != "list" && id != "default" {
					req = mux.SetURLVars(req, map[string]string{"id": id})
				}
				break
			}
		}
	}

	rr := httptest.NewRecorder()
	return rr, req
}

func TestAgentHandler_CreateAgent_Success(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	userID := "user-123"
	req := models.CreateAgentRequest{
		Name:         "Test Agent",
		Description:  "A test agent",
		SystemPrompt: "You are a helpful assistant",
		Model:        "gpt-4",
		Provider:     "openai",
	}

	expectedAgent := &models.Agent{
		ID:           "agent-123",
		UserID:       userID,
		Name:         req.Name,
		Description:  req.Description,
		SystemPrompt: req.SystemPrompt,
		Model:        req.Model,
		Provider:     req.Provider,
		CreatedAt:    time.Now(),
	}

	mockService.On("CreateAgent", mock.Anything, userID, mock.AnythingOfType("*models.CreateAgentRequest")).Return(expectedAgent, nil)

	rr, httpReq := setupAgentTest(userID, "POST", "/api/v1/agents", req)

	handler.CreateAgent(rr, httpReq)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp models.Agent
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedAgent.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestAgentHandler_CreateAgent_Unauthorized(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	req := models.CreateAgentRequest{
		Name:         "Test Agent",
		Description:  "A test agent",
		SystemPrompt: "You are a helpful assistant",
		Model:        "gpt-4",
		Provider:     "openai",
	}

	rr, httpReq := setupAgentTest("", "POST", "/api/v1/agents", req)

	handler.CreateAgent(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestAgentHandler_GetAgent_Success(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	agentID := "agent-123"
	expectedAgent := &models.Agent{
		ID:        agentID,
		UserID:    "user-123",
		Name:      "Test Agent",
		Model:     "gpt-4",
		CreatedAt: time.Now(),
	}

	mockService.On("GetAgent", mock.Anything, agentID).Return(expectedAgent, nil)

	rr, httpReq := setupAgentTest("", "GET", "/api/v1/agents/"+agentID, nil)

	handler.GetAgent(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Agent
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedAgent.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestAgentHandler_GetAgent_NotFound(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	agentID := "nonexistent-agent"
	mockService.On("GetAgent", mock.Anything, agentID).Return(nil, services.ErrAgentNotFound)

	rr, httpReq := setupAgentTest("", "GET", "/api/v1/agents/"+agentID, nil)

	handler.GetAgent(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "agent not found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestAgentHandler_DeleteAgent_Success(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	agentID := "agent-123"
	mockService.On("DeleteAgent", mock.Anything, agentID).Return(nil)

	rr, httpReq := setupAgentTest("", "DELETE", "/api/v1/agents/"+agentID, nil)

	handler.DeleteAgent(rr, httpReq)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	mockService.AssertExpectations(t)
}

func TestAgentHandler_GetUserAgents_Success(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	userID := "user-123"
	expectedAgents := []*models.Agent{
		{ID: "agent-1", UserID: userID, Name: "Agent 1"},
		{ID: "agent-2", UserID: userID, Name: "Agent 2"},
	}

	expectedResponse := &models.ListResponse{
		Items:      expectedAgents,
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockService.On("GetUserAgents", mock.Anything, userID, 1, 10).Return(expectedResponse, nil)

	rr, httpReq := setupAgentTest(userID, "GET", "/api/v1/agents?page=1&page_size=10", nil)

	handler.GetUserAgents(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)

	mockService.AssertExpectations(t)
}

func TestAgentHandler_GetUserAgents_Unauthorized(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	rr, httpReq := setupAgentTest("", "GET", "/api/v1/agents", nil)

	handler.GetUserAgents(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestAgentHandler_GetUserAgents_Pagination(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	userID := "user-123"
	expectedResponse := &models.ListResponse{
		Items:      []*models.Agent{},
		Total:      25,
		Page:       2,
		PageSize:   10,
		TotalPages: 3,
	}

	mockService.On("GetUserAgents", mock.Anything, userID, 2, 10).Return(expectedResponse, nil)

	rr, httpReq := setupAgentTest(userID, "GET", "/api/v1/agents?page=2&page_size=10", nil)

	handler.GetUserAgents(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, int64(25), resp.Total)

	mockService.AssertExpectations(t)
}

func TestAgentHandler_ListAgents_Success(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	expectedAgents := []*models.Agent{
		{ID: "agent-1", Name: "Agent 1", UserID: "user-1"},
		{ID: "agent-2", Name: "Agent 2", UserID: "user-2"},
	}

	expectedResponse := &models.ListResponse{
		Items:      expectedAgents,
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockService.On("ListAgents", mock.Anything, 1, 10).Return(expectedResponse, nil)

	rr, httpReq := setupAgentTest("", "GET", "/api/v1/agents/list?page=1&page_size=10", nil)

	handler.ListAgents(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)

	mockService.AssertExpectations(t)
}

func TestAgentHandler_UpdateAgent_Success(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	agentID := "agent-123"
	newName := "Updated Agent"
	req := models.UpdateAgentRequest{
		Name: &newName,
	}

	expectedAgent := &models.Agent{
		ID:        agentID,
		Name:      newName,
		UpdatedAt: time.Now(),
	}

	mockService.On("UpdateAgent", mock.Anything, agentID, mock.AnythingOfType("*models.UpdateAgentRequest")).Return(expectedAgent, nil)

	rr, httpReq := setupAgentTest("", "PUT", "/api/v1/agents/"+agentID, req)

	handler.UpdateAgent(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Agent
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, newName, resp.Name)

	mockService.AssertExpectations(t)
}

func TestAgentHandler_UpdateAgent_InvalidBody(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	agentID := "agent-123"
	rr, httpReq := setupAgentTest("", "PUT", "/api/v1/agents/"+agentID, "invalid json")

	handler.UpdateAgent(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request body", resp["error"])
}

func TestAgentHandler_GetDefaultAgent_Success(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	userID := "user-123"
	expectedAgent := &models.Agent{
		ID:      "agent-default",
		Name:    "Default Agent",
		UserID:  userID,
	}

	mockService.On("GetDefaultAgent", mock.Anything, userID).Return(expectedAgent, nil)

	rr, httpReq := setupAgentTest(userID, "GET", "/api/v1/agents/default", nil)

	handler.GetDefaultAgent(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Agent
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedAgent.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestAgentHandler_GetDefaultAgent_NotFound(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	userID := "user-123"
	mockService.On("GetDefaultAgent", mock.Anything, userID).Return(nil, services.ErrAgentNotFound)

	rr, httpReq := setupAgentTest(userID, "GET", "/api/v1/agents/default", nil)

	handler.GetDefaultAgent(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "no default agent found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestAgentHandler_GetDefaultAgent_Unauthorized(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	rr, httpReq := setupAgentTest("", "GET", "/api/v1/agents/default", nil)

	handler.GetDefaultAgent(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestAgentHandler_DeleteAgent_NotFound(t *testing.T) {
	mockService := new(MockAgentService)
	handler := NewAgentHandler(mockService)

	agentID := "nonexistent-agent"
	mockService.On("DeleteAgent", mock.Anything, agentID).Return(services.ErrAgentNotFound)

	rr, httpReq := setupAgentTest("", "DELETE", "/api/v1/agents/"+agentID, nil)

	handler.DeleteAgent(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "agent not found", resp["error"])

	mockService.AssertExpectations(t)
}
