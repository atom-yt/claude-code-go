package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
)

// ArtifactHandler handles artifact-related HTTP requests
type ArtifactHandler struct {
	artifactService *services.ArtifactService
}

// NewArtifactHandler creates a new artifact handler
func NewArtifactHandler(artifactService *services.ArtifactService) *ArtifactHandler {
	return &ArtifactHandler{
		artifactService: artifactService,
	}
}

// CreateArtifact handles POST /api/v1/artifacts
func (h *ArtifactHandler) CreateArtifact(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateArtifactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	artifact, err := h.artifactService.CreateArtifact(r.Context(), userID, &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create artifact")
		return
	}

	respondWithJSON(w, http.StatusCreated, artifact)
}

// GetArtifact handles GET /api/v1/artifacts/{id}
func (h *ArtifactHandler) GetArtifact(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "artifact id is required")
		return
	}

	artifact, err := h.artifactService.GetArtifact(r.Context(), id)
	if err != nil {
		if err == services.ErrArtifactNotFound {
			respondWithError(w, http.StatusNotFound, "artifact not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to get artifact")
		return
	}

	respondWithJSON(w, http.StatusOK, artifact)
}

// GetUserArtifacts handles GET /api/v1/artifacts (user's artifacts)
func (h *ArtifactHandler) GetUserArtifacts(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, pageSize := getPageParams(r)
	search := r.URL.Query().Get("search")

	response, err := h.artifactService.GetUserArtifacts(r.Context(), userID, search, page, pageSize)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get artifacts")
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// DeleteArtifact handles DELETE /api/v1/artifacts/{id}
func (h *ArtifactHandler) DeleteArtifact(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "artifact id is required")
		return
	}

	if err := h.artifactService.DeleteArtifact(r.Context(), id); err != nil {
		if err == services.ErrArtifactNotFound {
			respondWithError(w, http.StatusNotFound, "artifact not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to delete artifact")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

// GetArtifactStats handles GET /api/v1/artifacts/stats
func (h *ArtifactHandler) GetArtifactStats(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	stats, err := h.artifactService.GetArtifactStats(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get artifact stats")
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}
