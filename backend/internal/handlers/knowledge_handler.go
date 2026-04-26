package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
)

// KnowledgeHandler handles knowledge base HTTP requests
type KnowledgeHandler struct {
	knowledgeService *services.KnowledgeService
}

// NewKnowledgeHandler creates a new knowledge handler
func NewKnowledgeHandler(knowledgeService *services.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{
		knowledgeService: knowledgeService,
	}
}

// CreateKnowledge handles POST /api/v1/knowledge
func (h *KnowledgeHandler) CreateKnowledge(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateKnowledgeBaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	kb, err := h.knowledgeService.CreateKnowledge(r.Context(), userID, &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create knowledge base")
		return
	}

	respondWithJSON(w, http.StatusCreated, kb)
}

// GetKnowledge handles GET /api/v1/knowledge/{id}
func (h *KnowledgeHandler) GetKnowledge(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "knowledge id is required")
		return
	}

	kb, err := h.knowledgeService.GetKnowledge(r.Context(), id)
	if err != nil {
		if err == services.ErrKnowledgeNotFound {
			respondWithError(w, http.StatusNotFound, "knowledge base not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to get knowledge base")
		return
	}

	respondWithJSON(w, http.StatusOK, kb)
}

// GetUserKnowledge handles GET /api/v1/knowledge (user's knowledge bases)
func (h *KnowledgeHandler) GetUserKnowledge(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, pageSize := getPageParams(r)

	response, err := h.knowledgeService.GetUserKnowledge(r.Context(), userID, page, pageSize)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get knowledge bases")
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// UpdateKnowledge handles PUT /api/v1/knowledge/{id}
func (h *KnowledgeHandler) UpdateKnowledge(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "knowledge id is required")
		return
	}

	var req models.UpdateKnowledgeBaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	kb, err := h.knowledgeService.UpdateKnowledge(r.Context(), id, &req)
	if err != nil {
		if err == services.ErrKnowledgeNotFound {
			respondWithError(w, http.StatusNotFound, "knowledge base not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to update knowledge base")
		return
	}

	respondWithJSON(w, http.StatusOK, kb)
}

// DeleteKnowledge handles DELETE /api/v1/knowledge/{id}
func (h *KnowledgeHandler) DeleteKnowledge(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "knowledge id is required")
		return
	}

	if err := h.knowledgeService.DeleteKnowledge(r.Context(), id); err != nil {
		if err == services.ErrKnowledgeNotFound {
			respondWithError(w, http.StatusNotFound, "knowledge base not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to delete knowledge base")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
