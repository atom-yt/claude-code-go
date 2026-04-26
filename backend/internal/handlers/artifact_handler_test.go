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

// MockArtifactService is a mock implementation of ArtifactServiceInterface
type MockArtifactService struct {
	mock.Mock
}

func (m *MockArtifactService) CreateArtifact(ctx context.Context, userID string, req *models.CreateArtifactRequest) (*models.Artifact, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Artifact), args.Error(1)
}

func (m *MockArtifactService) GetArtifact(ctx context.Context, id string) (*models.Artifact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Artifact), args.Error(1)
}

func (m *MockArtifactService) GetUserArtifacts(ctx context.Context, userID string, search string, page, pageSize int) (*models.ListResponse, error) {
	args := m.Called(ctx, userID, search, page, pageSize)
	return args.Get(0).(*models.ListResponse), args.Error(1)
}

func (m *MockArtifactService) DeleteArtifact(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockArtifactService) GetArtifactStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func setupArtifactTest(userID string, method, path string, body interface{}) (*httptest.ResponseRecorder, *http.Request) {
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))

	if userID != "" {
		ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)
		req = req.WithContext(ctx)
	}

	// Extract ID from path
	if strings.Contains(path, "/artifacts/") {
		parts := strings.Split(path, "/")
		// Find to artifact ID (comes after /artifacts/)
		for i, part := range parts {
			if part == "artifacts" && i+1 < len(parts) {
				id := parts[i+1]
				req = mux.SetURLVars(req, map[string]string{"id": id})
				break
			}
		}
	}

	rr := httptest.NewRecorder()
	return rr, req
}

func TestArtifactHandler_CreateArtifact_Success(t *testing.T) {
	mockService := new(MockArtifactService)
	handler := NewArtifactHandler(mockService)

	userID := "user-123"
	req := models.CreateArtifactRequest{
		Title:   "Test Artifact",
		Content: "This is test content",
		Tags:    []string{"test", "artifact"},
	}

	expectedArtifact := &models.Artifact{
		ID:        "artifact-123",
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		Tags:      req.Tags,
		CreatedAt: time.Now(),
	}

	mockService.On("CreateArtifact", mock.Anything, userID, mock.AnythingOfType("*models.CreateArtifactRequest")).Return(expectedArtifact, nil)

	rr, httpReq := setupArtifactTest(userID, "POST", "/api/v1/artifacts", req)

	handler.CreateArtifact(rr, httpReq)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp models.Artifact
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedArtifact.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestArtifactHandler_CreateArtifact_Unauthorized(t *testing.T) {
	mockService := new(MockArtifactService)
	handler := NewArtifactHandler(mockService)

	req := models.CreateArtifactRequest{
		Title:   "Test Artifact",
		Content: "This is test content",
	}

	rr, httpReq := setupArtifactTest("", "POST", "/api/v1/artifacts", req)

	handler.CreateArtifact(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestArtifactHandler_GetArtifact_Success(t *testing.T) {
	mockService := new(MockArtifactService)
	handler := NewArtifactHandler(mockService)

	artifactID := "artifact-123"
	expectedArtifact := &models.Artifact{
		ID:        artifactID,
		UserID:    "user-123",
		Title:     "Test Artifact",
		Content:   "This is test content",
		CreatedAt: time.Now(),
	}

	mockService.On("GetArtifact", mock.Anything, artifactID).Return(expectedArtifact, nil)

	rr, httpReq := setupArtifactTest("", "GET", "/api/v1/artifacts/"+artifactID, nil)

	handler.GetArtifact(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Artifact
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedArtifact.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestArtifactHandler_GetArtifact_NotFound(t *testing.T) {
	mockService := new(MockArtifactService)
	handler := NewArtifactHandler(mockService)

	artifactID := "nonexistent-artifact"
	mockService.On("GetArtifact", mock.Anything, artifactID).Return(nil, services.ErrArtifactNotFound)

	rr, httpReq := setupArtifactTest("", "GET", "/api/v1/artifacts/"+artifactID, nil)

	handler.GetArtifact(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "artifact not found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestArtifactHandler_DeleteArtifact_Success(t *testing.T) {
	mockService := new(MockArtifactService)
	handler := NewArtifactHandler(mockService)

	artifactID := "artifact-123"
	mockService.On("DeleteArtifact", mock.Anything, artifactID).Return(nil)

	rr, httpReq := setupArtifactTest("", "DELETE", "/api/v1/artifacts/"+artifactID, nil)

	handler.DeleteArtifact(rr, httpReq)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	mockService.AssertExpectations(t)
}

func TestArtifactHandler_GetUserArtifacts_Success(t *testing.T) {
	mockService := new(MockArtifactService)
	handler := NewArtifactHandler(mockService)

	userID := "user-123"
	expectedArtifacts := []*models.Artifact{
		{ID: "artifact-1", UserID: userID, Title: "Artifact 1"},
		{ID: "artifact-2", UserID: userID, Title: "Artifact 2"},
	}

	expectedResponse := &models.ListResponse{
		Items:      expectedArtifacts,
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockService.On("GetUserArtifacts", mock.Anything, userID, "", 1, 10).Return(expectedResponse, nil)

	rr, httpReq := setupArtifactTest(userID, "GET", "/api/v1/artifacts?page=1&page_size=10", nil)

	handler.GetUserArtifacts(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)

	mockService.AssertExpectations(t)
}

func TestArtifactHandler_GetUserArtifacts_WithSearch(t *testing.T) {
	mockService := new(MockArtifactService)
	handler := NewArtifactHandler(mockService)

	userID := "user-123"
	search := "test"
	expectedArtifacts := []*models.Artifact{
		{ID: "artifact-1", UserID: userID, Title: "Test Artifact"},
	}

	expectedResponse := &models.ListResponse{
		Items:      expectedArtifacts,
		Total:      1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockService.On("GetUserArtifacts", mock.Anything, userID, search, 1, 10).Return(expectedResponse, nil)

	rr, httpReq := setupArtifactTest(userID, "GET", "/api/v1/artifacts?search="+search+"&page=1&page_size=10", nil)

	handler.GetUserArtifacts(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.Total)

	mockService.AssertExpectations(t)
}

func TestArtifactHandler_GetUserArtifacts_Unauthorized(t *testing.T) {
	mockService := new(MockArtifactService)
	handler := NewArtifactHandler(mockService)

	rr, httpReq := setupArtifactTest("", "GET", "/api/v1/artifacts", nil)

	handler.GetUserArtifacts(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestArtifactHandler_GetArtifactStats_Success(t *testing.T) {
	mockService := new(MockArtifactService)
	handler := NewArtifactHandler(mockService)

	userID := "user-123"
	expectedStats := map[string]interface{}{
		"total":       10,
		"by_type":     map[string]int{"code": 5, "document": 3, "other": 2},
		"recent_week": 3,
	}

	mockService.On("GetArtifactStats", mock.Anything, userID).Return(expectedStats, nil)

	rr, httpReq := setupArtifactTest(userID, "GET", "/api/v1/artifacts/stats", nil)

	handler.GetArtifactStats(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), resp["total"])

	mockService.AssertExpectations(t)
}

func TestArtifactHandler_GetArtifactStats_Unauthorized(t *testing.T) {
	mockService := new(MockArtifactService)
	handler := NewArtifactHandler(mockService)

	rr, httpReq := setupArtifactTest("", "GET", "/api/v1/artifacts/stats", nil)

	handler.GetArtifactStats(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestArtifactHandler_DeleteArtifact_NotFound(t *testing.T) {
	mockService := new(MockArtifactService)
	handler := NewArtifactHandler(mockService)

	artifactID := "nonexistent-artifact"
	mockService.On("DeleteArtifact", mock.Anything, artifactID).Return(services.ErrArtifactNotFound)

	rr, httpReq := setupArtifactTest("", "DELETE", "/api/v1/artifacts/"+artifactID, nil)

	handler.DeleteArtifact(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "artifact not found", resp["error"])

	mockService.AssertExpectations(t)
}
