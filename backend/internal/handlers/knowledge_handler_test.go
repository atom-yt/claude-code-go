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

// MockKnowledgeService is a mock implementation of KnowledgeServiceInterface
type MockKnowledgeService struct {
	mock.Mock
}

func (m *MockKnowledgeService) CreateKnowledge(ctx context.Context, userID string, req *models.CreateKnowledgeBaseRequest) (*models.KnowledgeBase, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.KnowledgeBase), args.Error(1)
}

func (m *MockKnowledgeService) GetKnowledge(ctx context.Context, id string) (*models.KnowledgeBase, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.KnowledgeBase), args.Error(1)
}

func (m *MockKnowledgeService) GetUserKnowledge(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error) {
	args := m.Called(ctx, userID, page, pageSize)
	return args.Get(0).(*models.ListResponse), args.Error(1)
}

func (m *MockKnowledgeService) UpdateKnowledge(ctx context.Context, id string, req *models.UpdateKnowledgeBaseRequest) (*models.KnowledgeBase, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.KnowledgeBase), args.Error(1)
}

func (m *MockKnowledgeService) DeleteKnowledge(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupKnowledgeTest(userID string, method, path string, body interface{}) (*httptest.ResponseRecorder, *http.Request) {
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))

	if userID != "" {
		ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)
		req = req.WithContext(ctx)
	}

	// Extract ID from path
	if strings.Contains(path, "/knowledge/") {
		parts := strings.Split(path, "/")
		// Find to knowledge ID (comes after /knowledge/)
		for i, part := range parts {
			if part == "knowledge" && i+1 < len(parts) {
				id := parts[i+1]
				req = mux.SetURLVars(req, map[string]string{"id": id})
				break
			}
		}
	}

	rr := httptest.NewRecorder()
	return rr, req
}

func TestKnowledgeHandler_CreateKnowledge_Success(t *testing.T) {
	mockService := new(MockKnowledgeService)
	handler := NewKnowledgeHandler(mockService)

	userID := "user-123"
	req := models.CreateKnowledgeBaseRequest{
		Name:        "Test Knowledge",
		Description: "A test knowledge base",
		Type:        "documents",
		Source:      "local",
	}

	expectedKB := &models.KnowledgeBase{
		ID:          "kb-123",
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Source:      req.Source,
		CreatedAt:   time.Now(),
	}

	mockService.On("CreateKnowledge", mock.Anything, userID, mock.AnythingOfType("*models.CreateKnowledgeBaseRequest")).Return(expectedKB, nil)

	rr, httpReq := setupKnowledgeTest(userID, "POST", "/api/v1/knowledge", req)

	handler.CreateKnowledge(rr, httpReq)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp models.KnowledgeBase
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedKB.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestKnowledgeHandler_CreateKnowledge_Unauthorized(t *testing.T) {
	mockService := new(MockKnowledgeService)
	handler := NewKnowledgeHandler(mockService)

	req := models.CreateKnowledgeBaseRequest{
		Name:        "Test Knowledge",
		Description: "A test knowledge base",
		Type:        "documents",
	}

	rr, httpReq := setupKnowledgeTest("", "POST", "/api/v1/knowledge", req)

	handler.CreateKnowledge(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestKnowledgeHandler_GetKnowledge_Success(t *testing.T) {
	mockService := new(MockKnowledgeService)
	handler := NewKnowledgeHandler(mockService)

	kbID := "kb-123"
	expectedKB := &models.KnowledgeBase{
		ID:          kbID,
		UserID:      "user-123",
		Name:        "Test Knowledge",
		Description: "A test knowledge base",
		Type:        "documents",
		Source:      "local",
		CreatedAt:   time.Now(),
	}

	mockService.On("GetKnowledge", mock.Anything, kbID).Return(expectedKB, nil)

	rr, httpReq := setupKnowledgeTest("", "GET", "/api/v1/knowledge/"+kbID, nil)

	handler.GetKnowledge(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.KnowledgeBase
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedKB.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestKnowledgeHandler_GetKnowledge_NotFound(t *testing.T) {
	mockService := new(MockKnowledgeService)
	handler := NewKnowledgeHandler(mockService)

	kbID := "nonexistent-kb"
	mockService.On("GetKnowledge", mock.Anything, kbID).Return(nil, services.ErrKnowledgeNotFound)

	rr, httpReq := setupKnowledgeTest("", "GET", "/api/v1/knowledge/"+kbID, nil)

	handler.GetKnowledge(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "knowledge base not found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestKnowledgeHandler_DeleteKnowledge_Success(t *testing.T) {
	mockService := new(MockKnowledgeService)
	handler := NewKnowledgeHandler(mockService)

	kbID := "kb-123"
	mockService.On("DeleteKnowledge", mock.Anything, kbID).Return(nil)

	rr, httpReq := setupKnowledgeTest("", "DELETE", "/api/v1/knowledge/"+kbID, nil)

	handler.DeleteKnowledge(rr, httpReq)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	mockService.AssertExpectations(t)
}

func TestKnowledgeHandler_GetUserKnowledge_Success(t *testing.T) {
	mockService := new(MockKnowledgeService)
	handler := NewKnowledgeHandler(mockService)

	userID := "user-123"
	expectedKBs := []*models.KnowledgeBase{
		{ID: "kb-1", UserID: userID, Name: "KB 1"},
		{ID: "kb-2", UserID: userID, Name: "KB 2"},
	}

	expectedResponse := &models.ListResponse{
		Items:      expectedKBs,
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockService.On("GetUserKnowledge", mock.Anything, userID, 1, 10).Return(expectedResponse, nil)

	rr, httpReq := setupKnowledgeTest(userID, "GET", "/api/v1/knowledge?page=1&page_size=10", nil)

	handler.GetUserKnowledge(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)

	mockService.AssertExpectations(t)
}

func TestKnowledgeHandler_GetUserKnowledge_Unauthorized(t *testing.T) {
	mockService := new(MockKnowledgeService)
	handler := NewKnowledgeHandler(mockService)

	rr, httpReq := setupKnowledgeTest("", "GET", "/api/v1/knowledge", nil)

	handler.GetUserKnowledge(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestKnowledgeHandler_UpdateKnowledge_Success(t *testing.T) {
	mockService := new(MockKnowledgeService)
	handler := NewKnowledgeHandler(mockService)

	kbID := "kb-123"
	newName := "Updated KB"
	req := models.UpdateKnowledgeBaseRequest{
		Name: &newName,
	}

	expectedKB := &models.KnowledgeBase{
		ID:        kbID,
		Name:      newName,
		UpdatedAt: time.Now(),
	}

	mockService.On("UpdateKnowledge", mock.Anything, kbID, mock.AnythingOfType("*models.UpdateKnowledgeBaseRequest")).Return(expectedKB, nil)

	rr, httpReq := setupKnowledgeTest("", "PUT", "/api/v1/knowledge/"+kbID, req)

	handler.UpdateKnowledge(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.KnowledgeBase
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, newName, resp.Name)

	mockService.AssertExpectations(t)
}

func TestKnowledgeHandler_UpdateKnowledge_InvalidBody(t *testing.T) {
	mockService := new(MockKnowledgeService)
	handler := NewKnowledgeHandler(mockService)

	kbID := "kb-123"
	rr, httpReq := setupKnowledgeTest("", "PUT", "/api/v1/knowledge/"+kbID, "invalid json")

	handler.UpdateKnowledge(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request body", resp["error"])
}

func TestKnowledgeHandler_DeleteKnowledge_NotFound(t *testing.T) {
	mockService := new(MockKnowledgeService)
	handler := NewKnowledgeHandler(mockService)

	kbID := "nonexistent-kb"
	mockService.On("DeleteKnowledge", mock.Anything, kbID).Return(services.ErrKnowledgeNotFound)

	rr, httpReq := setupKnowledgeTest("", "DELETE", "/api/v1/knowledge/"+kbID, nil)

	handler.DeleteKnowledge(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "knowledge base not found", resp["error"])

	mockService.AssertExpectations(t)
}
