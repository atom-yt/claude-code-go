package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
)

// MessageHandler handles message-related HTTP requests
type MessageHandler struct {
	messageService *services.MessageService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// CreateMessage handles POST /api/v1/sessions/{id}/messages
func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := getURLParam(r, "id")
	if sessionID == "" {
		respondWithError(w, http.StatusBadRequest, "session id is required")
		return
	}

	var req models.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	message, err := h.messageService.CreateMessage(r.Context(), sessionID, &req)
	if err != nil {
		if err == services.ErrSessionNotFound {
			respondWithError(w, http.StatusNotFound, "session not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to create message")
		return
	}

	respondWithJSON(w, http.StatusCreated, message)
}

// GetMessage handles GET /api/v1/messages/{id}
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "message id is required")
		return
	}

	message, err := h.messageService.GetMessage(r.Context(), id)
	if err != nil {
		if err == services.ErrMessageNotFound {
			respondWithError(w, http.StatusNotFound, "message not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to get message")
		return
	}

	respondWithJSON(w, http.StatusOK, message)
}

// GetSessionMessages handles GET /api/v1/sessions/{id}/messages
func (h *MessageHandler) GetSessionMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := getURLParam(r, "id")
	if sessionID == "" {
		respondWithError(w, http.StatusBadRequest, "session id is required")
		return
	}

	page, pageSize := getPageParams(r)

	response, err := h.messageService.GetSessionMessages(r.Context(), sessionID, page, pageSize)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get messages")
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetRecentMessages handles GET /api/v1/sessions/{id}/messages/recent
func (h *MessageHandler) GetRecentMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := getURLParam(r, "id")
	if sessionID == "" {
		respondWithError(w, http.StatusBadRequest, "session id is required")
		return
	}

	limit := getQueryInt(r, "limit", 10)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	messages, err := h.messageService.GetRecentMessages(r.Context(), sessionID, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get recent messages")
		return
	}

	respondWithJSON(w, http.StatusOK, messages)
}

// DeleteMessage handles DELETE /api/v1/messages/{id}
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "message id is required")
		return
	}

	if err := h.messageService.DeleteMessage(r.Context(), id); err != nil {
		if err == services.ErrMessageNotFound {
			respondWithError(w, http.StatusNotFound, "message not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to delete message")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}