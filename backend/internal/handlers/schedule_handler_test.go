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

// MockScheduleService is a mock implementation of ScheduleServiceInterface
type MockScheduleService struct {
	mock.Mock
}

func (m *MockScheduleService) CreateSchedule(ctx context.Context, userID string, req *models.CreateScheduledTaskRequest) (*models.ScheduledTask, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ScheduledTask), args.Error(1)
}

func (m *MockScheduleService) GetSchedule(ctx context.Context, id string) (*models.ScheduledTask, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ScheduledTask), args.Error(1)
}

func (m *MockScheduleService) GetUserSchedules(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error) {
	args := m.Called(ctx, userID, page, pageSize)
	return args.Get(0).(*models.ListResponse), args.Error(1)
}

func (m *MockScheduleService) UpdateSchedule(ctx context.Context, id string, req *models.UpdateScheduledTaskRequest) (*models.ScheduledTask, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ScheduledTask), args.Error(1)
}

func (m *MockScheduleService) DeleteSchedule(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockScheduleService) ToggleSchedule(ctx context.Context, id string, enabled bool) error {
	args := m.Called(ctx, id, enabled)
	return args.Error(0)
}

func setupScheduleTest(userID string, method, path string, body interface{}) (*httptest.ResponseRecorder, *http.Request) {
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))

	if userID != "" {
		ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)
		req = req.WithContext(ctx)
	}

	// Extract ID from path
	if strings.Contains(path, "/schedules/") {
		parts := strings.Split(path, "/")
		// Find the schedule ID (comes after /schedules/)
		for i, part := range parts {
			if part == "schedules" && i+1 < len(parts) {
				// For /schedules/{id} or /schedules/{id}/toggle
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

func TestScheduleHandler_CreateSchedule_Success(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	userID := "user-123"
	req := models.CreateScheduledTaskRequest{
		Title:        "Test Schedule",
		Prompt:       "Run this task",
		ScheduleType: "daily",
		ScheduleTime: "09:00",
	}

	expectedTask := &models.ScheduledTask{
		ID:           "schedule-123",
		UserID:       userID,
		Title:        req.Title,
		Prompt:       req.Prompt,
		ScheduleType: req.ScheduleType,
		ScheduleTime: req.ScheduleTime,
		Enabled:      true,
		CreatedAt:    time.Now(),
	}

	mockService.On("CreateSchedule", mock.Anything, userID, mock.AnythingOfType("*models.CreateScheduledTaskRequest")).Return(expectedTask, nil)

	rr, httpReq := setupScheduleTest(userID, "POST", "/api/v1/schedules", req)

	handler.CreateSchedule(rr, httpReq)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp models.ScheduledTask
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedTask.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestScheduleHandler_CreateSchedule_Unauthorized(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	req := models.CreateScheduledTaskRequest{
		Title:        "Test Schedule",
		Prompt:       "Run this task",
		ScheduleType: "daily",
		ScheduleTime: "09:00",
	}

	rr, httpReq := setupScheduleTest("", "POST", "/api/v1/schedules", req)

	handler.CreateSchedule(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestScheduleHandler_GetSchedule_Success(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	scheduleID := "schedule-123"
	expectedTask := &models.ScheduledTask{
		ID:           scheduleID,
		UserID:       "user-123",
		Title:        "Test Schedule",
		Prompt:       "Run this task",
		ScheduleType: "daily",
		ScheduleTime: "09:00",
		Enabled:      true,
		CreatedAt:    time.Now(),
	}

	mockService.On("GetSchedule", mock.Anything, scheduleID).Return(expectedTask, nil)

	rr, httpReq := setupScheduleTest("", "GET", "/api/v1/schedules/"+scheduleID, nil)

	handler.GetSchedule(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ScheduledTask
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedTask.ID, resp.ID)

	mockService.AssertExpectations(t)
}

func TestScheduleHandler_GetSchedule_NotFound(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	scheduleID := "nonexistent-schedule"
	mockService.On("GetSchedule", mock.Anything, scheduleID).Return(nil, services.ErrScheduleNotFound)

	rr, httpReq := setupScheduleTest("", "GET", "/api/v1/schedules/"+scheduleID, nil)

	handler.GetSchedule(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "scheduled task not found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestScheduleHandler_DeleteSchedule_Success(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	scheduleID := "schedule-123"
	mockService.On("DeleteSchedule", mock.Anything, scheduleID).Return(nil)

	rr, httpReq := setupScheduleTest("", "DELETE", "/api/v1/schedules/"+scheduleID, nil)

	handler.DeleteSchedule(rr, httpReq)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	mockService.AssertExpectations(t)
}

func TestScheduleHandler_GetUserSchedules_Success(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	userID := "user-123"
	expectedSchedules := []*models.ScheduledTask{
		{ID: "schedule-1", UserID: userID, Title: "Schedule 1"},
		{ID: "schedule-2", UserID: userID, Title: "Schedule 2"},
	}

	expectedResponse := &models.ListResponse{
		Items:      expectedSchedules,
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockService.On("GetUserSchedules", mock.Anything, userID, 1, 10).Return(expectedResponse, nil)

	rr, httpReq := setupScheduleTest(userID, "GET", "/api/v1/schedules?page=1&page_size=10", nil)

	handler.GetUserSchedules(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ListResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)

	mockService.AssertExpectations(t)
}

func TestScheduleHandler_GetUserSchedules_Unauthorized(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	rr, httpReq := setupScheduleTest("", "GET", "/api/v1/schedules", nil)

	handler.GetUserSchedules(rr, httpReq)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestScheduleHandler_UpdateSchedule_Success(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	scheduleID := "schedule-123"
	newTitle := "Updated Schedule"
	req := models.UpdateScheduledTaskRequest{
		Title: &newTitle,
	}

	expectedTask := &models.ScheduledTask{
		ID:        scheduleID,
		Title:     newTitle,
		UpdatedAt: time.Now(),
	}

	mockService.On("UpdateSchedule", mock.Anything, scheduleID, mock.AnythingOfType("*models.UpdateScheduledTaskRequest")).Return(expectedTask, nil)

	rr, httpReq := setupScheduleTest("", "PUT", "/api/v1/schedules/"+scheduleID, req)

	handler.UpdateSchedule(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.ScheduledTask
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, newTitle, resp.Title)

	mockService.AssertExpectations(t)
}

func TestScheduleHandler_UpdateSchedule_InvalidBody(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	scheduleID := "schedule-123"
	rr, httpReq := setupScheduleTest("", "PUT", "/api/v1/schedules/"+scheduleID, "invalid json")

	handler.UpdateSchedule(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request body", resp["error"])
}

func TestScheduleHandler_ToggleSchedule_Enable(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	scheduleID := "schedule-123"
	body := struct {
		Enabled bool `json:"enabled"`
	}{Enabled: true}

	mockService.On("ToggleSchedule", mock.Anything, scheduleID, true).Return(nil)

	rr, httpReq := setupScheduleTest("", "PUT", "/api/v1/schedules/"+scheduleID+"/toggle", body)

	handler.ToggleSchedule(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["enabled"])

	mockService.AssertExpectations(t)
}

func TestScheduleHandler_ToggleSchedule_Disable(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	scheduleID := "schedule-123"
	body := struct {
		Enabled bool `json:"enabled"`
	}{Enabled: false}

	mockService.On("ToggleSchedule", mock.Anything, scheduleID, false).Return(nil)

	rr, httpReq := setupScheduleTest("", "PUT", "/api/v1/schedules/"+scheduleID+"/toggle", body)

	handler.ToggleSchedule(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["enabled"])

	mockService.AssertExpectations(t)
}

func TestScheduleHandler_ToggleSchedule_NotFound(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	scheduleID := "nonexistent-schedule"
	body := struct {
		Enabled bool `json:"enabled"`
	}{Enabled: true}

	mockService.On("ToggleSchedule", mock.Anything, scheduleID, true).Return(services.ErrScheduleNotFound)

	rr, httpReq := setupScheduleTest("", "PUT", "/api/v1/schedules/"+scheduleID+"/toggle", body)

	handler.ToggleSchedule(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "scheduled task not found", resp["error"])

	mockService.AssertExpectations(t)
}

func TestScheduleHandler_DeleteSchedule_NotFound(t *testing.T) {
	mockService := new(MockScheduleService)
	handler := NewScheduleHandler(mockService)

	scheduleID := "nonexistent-schedule"
	mockService.On("DeleteSchedule", mock.Anything, scheduleID).Return(services.ErrScheduleNotFound)

	rr, httpReq := setupScheduleTest("", "DELETE", "/api/v1/schedules/"+scheduleID, nil)

	handler.DeleteSchedule(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "scheduled task not found", resp["error"])

	mockService.AssertExpectations(t)
}
