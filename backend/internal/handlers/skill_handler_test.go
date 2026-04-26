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

// MockSkillService is a mock implementation of SkillServiceInterface
type MockSkillService struct {
	mock.Mock
}

func (m *MockSkillService) CreateSkill(ctx context.Context, userID string, req *models.CreateSkillRequest) (*models.Skill, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Skill), args.Error(1)
}

func (m *MockSkillService) GetSkill(ctx context.Context, id string) (*models.Skill, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Skill), args.Error(1)
}

func (m *MockSkillService) GetUserSkills(ctx context.Context, userID string, category string, page, pageSize int) (*models.ListResponse, error) {
	args := m.Called(ctx, userID, category, page, pageSize)
	return args.Get(0).(*models.ListResponse), args.Error(1)
}

func (m *MockSkillService) UpdateSkill(ctx context.Context, id string, req *models.UpdateSkillRequest) (*models.Skill, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Skill), args.Error(1)
}

func (m *MockSkillService) DeleteSkill(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSkillService) ToggleSkill(ctx context.Context, id string, enabled bool) error {
	args := m.Called(ctx, id, enabled)
	return args.Error(0)
}

func setupSkillTest(userID string, method, path string, body interface{}) (*httptest.ResponseRecorder, *http.Request) {
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))

	if userID != "" {
		ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)
		req = req.WithContext(ctx)
	}

	// Extract ID from path
	if strings.Contains(path, "/skills/") {
		parts := strings.Split(path, "/")
		// Find the skill ID (comes after /skills/)
		for i, part := range parts {
			if part == "skills" && i+1 < len(parts) {
				// For /skills/{id} or /skills/{id}/toggle
				// Take the next part, but stop before "toggle" if present
				id := parts[i+1]
				if id == "toggle" && i+2 < len(parts) {
					id = parts[i+2]
				}
				req = mux.SetURLVars(req, map[string]string{"id": id})
				break
			}
		}
	}

	rr := httptest.NewRecorder()
	return rr, req
}

func TestSkillHandler_CreateSkill_Success(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	userID := "user-123"
	req := models.CreateSkillRequest{
		Name:        "Test Skill",
		Description: "A test skill",
		Category:    "personal",
	}

	userIDPtr := &userID
	expectedSkill := &models.Skill{
		ID:          "skill-123",
		UserID:      userIDPtr,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Enabled:     true,
		CreatedAt:   time.Now(),
	}

	mockService.On("CreateSkill", mock.Anything, userID, mock.AnythingOfType("*models.CreateSkillRequest")).Return(expectedSkill, nil)

	rr, httpReq := setupSkillTest(userID, "POST", "/api/v1/skills", req)

	handler.CreateSkill(rr, httpReq)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp models.Skill
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedSkill.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestSkillHandler_CreateSkill_Unauthorized(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	req := models.CreateSkillRequest{
		Name:        "Test Skill",
		Description: "A test skill",
		Category:    "personal",
	}

	rr, httpReq := setupSkillTest("", "POST", "/api/v1/skills", req)

	handler.CreateSkill(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestSkillHandler_GetSkill_Success(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	skillID := "skill-123"
	userID := "user-123"
	userIDPtr := &userID
	expectedSkill := &models.Skill{
		ID:        skillID,
		UserID:    userIDPtr,
		Name:      "Test Skill",
		Category:  "personal",
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	mockService.On("GetSkill", mock.Anything, skillID).Return(expectedSkill, nil)

	rr, httpReq := setupSkillTest("", "GET", "/api/v1/skills/"+skillID, nil)

	handler.GetSkill(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Skill
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedSkill.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestSkillHandler_GetSkill_NotFound(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	skillID := "nonexistent-skill"
	mockService.On("GetSkill", mock.Anything, skillID).Return(nil, services.ErrSkillNotFound)

	rr, httpReq := setupSkillTest("", "GET", "/api/v1/skills/"+skillID, nil)

	handler.GetSkill(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "skill not found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestSkillHandler_DeleteSkill_Success(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	skillID := "skill-123"
	mockService.On("DeleteSkill", mock.Anything, skillID).Return(nil)

	rr, httpReq := setupSkillTest("", "DELETE", "/api/v1/skills/"+skillID, nil)

	handler.DeleteSkill(rr, httpReq)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	mockService.AssertExpectations(t)
}

func TestSkillHandler_GetUserSkills_Success(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	userID := "user-123"
	expectedSkills := []*models.Skill{
		{ID: "skill-1", Name: "Skill 1", Category: "personal"},
		{ID: "skill-2", Name: "Skill 2", Category: "team"},
	}

	expectedResponse := &models.ListResponse{
		Items:      expectedSkills,
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockService.On("GetUserSkills", mock.Anything, userID, "", 1, 10).Return(expectedResponse, nil)

	rr, httpReq := setupSkillTest(userID, "GET", "/api/v1/skills?page=1&page_size=10", nil)

	handler.GetUserSkills(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)

	mockService.AssertExpectations(t)
}

func TestSkillHandler_GetUserSkills_WithCategory(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	userID := "user-123"
	category := "personal"
	expectedSkills := []*models.Skill{
		{ID: "skill-1", Name: "Skill 1", Category: "personal"},
	}

	expectedResponse := &models.ListResponse{
		Items:      expectedSkills,
		Total:      1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockService.On("GetUserSkills", mock.Anything, userID, category, 1, 10).Return(expectedResponse, nil)

	rr, httpReq := setupSkillTest(userID, "GET", "/api/v1/skills?category="+category+"&page=1&page_size=10", nil)

	handler.GetUserSkills(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.Total)

	mockService.AssertExpectations(t)
}

func TestSkillHandler_GetUserSkills_Unauthorized(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	rr, httpReq := setupSkillTest("", "GET", "/api/v1/skills", nil)

	handler.GetUserSkills(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestSkillHandler_UpdateSkill_Success(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	skillID := "skill-123"
	newName := "Updated Skill"
	req := models.UpdateSkillRequest{
		Name: &newName,
	}

	expectedSkill := &models.Skill{
		ID:        skillID,
		Name:      newName,
		UpdatedAt: time.Now(),
	}

	mockService.On("UpdateSkill", mock.Anything, skillID, mock.AnythingOfType("*models.UpdateSkillRequest")).Return(expectedSkill, nil)

	rr, httpReq := setupSkillTest("", "PUT", "/api/v1/skills/"+skillID, req)

	handler.UpdateSkill(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Skill
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, newName, resp.Name)

	mockService.AssertExpectations(t)
}

func TestSkillHandler_UpdateSkill_InvalidBody(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	skillID := "skill-123"
	rr, httpReq := setupSkillTest("", "PUT", "/api/v1/skills/"+skillID, "invalid json")

	handler.UpdateSkill(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request body", resp["error"])
}

func TestSkillHandler_ToggleSkill_Enable(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	skillID := "skill-123"
	body := struct {
		Enabled bool `json:"enabled"`
	}{Enabled: true}

	mockService.On("ToggleSkill", mock.Anything, skillID, true).Return(nil)

	rr, httpReq := setupSkillTest("", "PUT", "/api/v1/skills/"+skillID+"/toggle", body)

	handler.ToggleSkill(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]bool
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["enabled"])

	mockService.AssertExpectations(t)
}

func TestSkillHandler_ToggleSkill_Disable(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	skillID := "skill-123"
	body := struct {
		Enabled bool `json:"enabled"`
	}{Enabled: false}

	mockService.On("ToggleSkill", mock.Anything, skillID, false).Return(nil)

	rr, httpReq := setupSkillTest("", "PUT", "/api/v1/skills/"+skillID+"/toggle", body)

	handler.ToggleSkill(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]bool
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["enabled"])

	mockService.AssertExpectations(t)
}

func TestSkillHandler_ToggleSkill_NotFound(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	skillID := "nonexistent-skill"
	body := struct {
		Enabled bool `json:"enabled"`
	}{Enabled: true}

	mockService.On("ToggleSkill", mock.Anything, skillID, true).Return(services.ErrSkillNotFound)

	rr, httpReq := setupSkillTest("", "PUT", "/api/v1/skills/"+skillID+"/toggle", body)

	handler.ToggleSkill(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "skill not found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestSkillHandler_DeleteSkill_NotFound(t *testing.T) {
	mockService := new(MockSkillService)
	handler := NewSkillHandler(mockService)

	skillID := "nonexistent-skill"
	mockService.On("DeleteSkill", mock.Anything, skillID).Return(services.ErrSkillNotFound)

	rr, httpReq := setupSkillTest("", "DELETE", "/api/v1/skills/"+skillID, nil)

	handler.DeleteSkill(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "skill not found", resp["error"])

	mockService.AssertExpectations(t)
}
