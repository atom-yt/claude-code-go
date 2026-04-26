package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
)

// ScheduleHandler handles scheduled task HTTP requests
type ScheduleHandler struct {
	scheduleService services.ScheduleServiceInterface
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(scheduleService services.ScheduleServiceInterface) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
	}
}

// CreateSchedule handles POST /api/v1/schedules
func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateScheduledTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	task, err := h.scheduleService.CreateSchedule(r.Context(), userID, &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create scheduled task")
		return
	}

	respondWithJSON(w, http.StatusCreated, task)
}

// GetSchedule handles GET /api/v1/schedules/{id}
func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "schedule id is required")
		return
	}

	task, err := h.scheduleService.GetSchedule(r.Context(), id)
	if err != nil {
		if err == services.ErrScheduleNotFound {
			respondWithError(w, http.StatusNotFound, "scheduled task not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to get scheduled task")
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

// GetUserSchedules handles GET /api/v1/schedules (user's scheduled tasks)
func (h *ScheduleHandler) GetUserSchedules(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, pageSize := getPageParams(r)

	response, err := h.scheduleService.GetUserSchedules(r.Context(), userID, page, pageSize)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get scheduled tasks")
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// UpdateSchedule handles PUT /api/v1/schedules/{id}
func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "schedule id is required")
		return
	}

	var req models.UpdateScheduledTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	task, err := h.scheduleService.UpdateSchedule(r.Context(), id, &req)
	if err != nil {
		if err == services.ErrScheduleNotFound {
			respondWithError(w, http.StatusNotFound, "scheduled task not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to update scheduled task")
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

// DeleteSchedule handles DELETE /api/v1/schedules/{id}
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "schedule id is required")
		return
	}

	if err := h.scheduleService.DeleteSchedule(r.Context(), id); err != nil {
		if err == services.ErrScheduleNotFound {
			respondWithError(w, http.StatusNotFound, "scheduled task not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to delete scheduled task")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

// ToggleSchedule handles PUT /api/v1/schedules/{id}/toggle
func (h *ScheduleHandler) ToggleSchedule(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "schedule id is required")
		return
	}

	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	if err := h.scheduleService.ToggleSchedule(r.Context(), id, body.Enabled); err != nil {
		if err == services.ErrScheduleNotFound {
			respondWithError(w, http.StatusNotFound, "scheduled task not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to toggle scheduled task")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"enabled": body.Enabled})
}
