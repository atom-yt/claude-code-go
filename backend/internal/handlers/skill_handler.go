package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
)

// SkillHandler handles skill-related HTTP requests
type SkillHandler struct {
	skillService services.SkillServiceInterface
}

// NewSkillHandler creates a new skill handler
func NewSkillHandler(skillService services.SkillServiceInterface) *SkillHandler {
	return &SkillHandler{
		skillService: skillService,
	}
}

// CreateSkill handles POST /api/v1/skills
func (h *SkillHandler) CreateSkill(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	skill, err := h.skillService.CreateSkill(r.Context(), userID, &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create skill")
		return
	}

	respondWithJSON(w, http.StatusCreated, skill)
}

// GetSkill handles GET /api/v1/skills/{id}
func (h *SkillHandler) GetSkill(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "skill id is required")
		return
	}

	skill, err := h.skillService.GetSkill(r.Context(), id)
	if err != nil {
		if err == services.ErrSkillNotFound {
			respondWithError(w, http.StatusNotFound, "skill not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to get skill")
		return
	}

	respondWithJSON(w, http.StatusOK, skill)
}

// GetUserSkills handles GET /api/v1/skills (user's skills)
func (h *SkillHandler) GetUserSkills(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, pageSize := getPageParams(r)
	category := r.URL.Query().Get("category")

	response, err := h.skillService.GetUserSkills(r.Context(), userID, category, page, pageSize)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get skills")
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// UpdateSkill handles PUT /api/v1/skills/{id}
func (h *SkillHandler) UpdateSkill(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "skill id is required")
		return
	}

	var req models.UpdateSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	skill, err := h.skillService.UpdateSkill(r.Context(), id, &req)
	if err != nil {
		if err == services.ErrSkillNotFound {
			respondWithError(w, http.StatusNotFound, "skill not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to update skill")
		return
	}

	respondWithJSON(w, http.StatusOK, skill)
}

// DeleteSkill handles DELETE /api/v1/skills/{id}
func (h *SkillHandler) DeleteSkill(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "skill id is required")
		return
	}

	if err := h.skillService.DeleteSkill(r.Context(), id); err != nil {
		if err == services.ErrSkillNotFound {
			respondWithError(w, http.StatusNotFound, "skill not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to delete skill")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

// ToggleSkill handles PUT /api/v1/skills/{id}/toggle
func (h *SkillHandler) ToggleSkill(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "skill id is required")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	if err := h.skillService.ToggleSkill(r.Context(), id, req.Enabled); err != nil {
		if err == services.ErrSkillNotFound {
			respondWithError(w, http.StatusNotFound, "skill not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to toggle skill")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]bool{"enabled": req.Enabled})
}
