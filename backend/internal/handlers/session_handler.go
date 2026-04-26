package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
)

// SessionHandler handles session-related HTTP requests
type SessionHandler struct {
	sessionService services.SessionServiceInterface
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionService services.SessionServiceInterface) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// CreateSession handles POST /api/v1/sessions
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	session, err := h.sessionService.CreateSession(r.Context(), userID, &req)
	if err != nil {
		if err == services.ErrAgentNotFound {
			respondWithError(w, http.StatusBadRequest, "agent not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	respondWithJSON(w, http.StatusCreated, session)
}

// GetSession handles GET /api/v1/sessions/{id}
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "session id is required")
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), id)
	if err != nil {
		if err == services.ErrSessionNotFound {
			respondWithError(w, http.StatusNotFound, "session not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to get session")
		return
	}

	respondWithJSON(w, http.StatusOK, session)
}

// GetUserSessions handles GET /api/v1/sessions
func (h *SessionHandler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, pageSize := getPageParams(r)

	response, err := h.sessionService.GetUserSessions(r.Context(), userID, page, pageSize)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get sessions")
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// UpdateSession handles PUT /api/v1/sessions/{id}
func (h *SessionHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "session id is required")
		return
	}

	var req models.UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	session, err := h.sessionService.UpdateSession(r.Context(), id, &req)
	if err != nil {
		if err == services.ErrSessionNotFound {
			respondWithError(w, http.StatusNotFound, "session not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to update session")
		return
	}

	respondWithJSON(w, http.StatusOK, session)
}

// DeleteSession handles DELETE /api/v1/sessions/{id}
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "session id is required")
		return
	}

	if err := h.sessionService.DeleteSession(r.Context(), id); err != nil {
		if err == services.ErrSessionNotFound {
			respondWithError(w, http.StatusNotFound, "session not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to delete session")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

// ArchiveSession handles POST /api/v1/sessions/{id}/archive
func (h *SessionHandler) ArchiveSession(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "session id is required")
		return
	}

	if err := h.sessionService.ArchiveSession(r.Context(), id); err != nil {
		if err == services.ErrSessionNotFound {
			respondWithError(w, http.StatusNotFound, "session not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to archive session")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

// GetActiveSessions handles GET /api/v1/sessions/active
func (h *SessionHandler) GetActiveSessions(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessions, err := h.sessionService.GetActiveSessions(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get active sessions")
		return
	}

	respondWithJSON(w, http.StatusOK, sessions)
}