package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
)

// AgentHandler handles agent-related HTTP requests
type AgentHandler struct {
	agentService *services.AgentService
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(agentService *services.AgentService) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
	}
}

// CreateAgent handles POST /api/v1/agents
func (h *AgentHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	agent, err := h.agentService.CreateAgent(r.Context(), userID, &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create agent")
		return
	}

	respondWithJSON(w, http.StatusCreated, agent)
}

// GetAgent handles GET /api/v1/agents/{id}
func (h *AgentHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "agent id is required")
		return
	}

	agent, err := h.agentService.GetAgent(r.Context(), id)
	if err != nil {
		if err == services.ErrAgentNotFound {
			respondWithError(w, http.StatusNotFound, "agent not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to get agent")
		return
	}

	respondWithJSON(w, http.StatusOK, agent)
}

// GetUserAgents handles GET /api/v1/agents (user's agents)
func (h *AgentHandler) GetUserAgents(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, pageSize := getPageParams(r)

	response, err := h.agentService.GetUserAgents(r.Context(), userID, page, pageSize)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get agents")
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// ListAgents handles GET /api/v1/agents/list (all agents)
func (h *AgentHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	page, pageSize := getPageParams(r)

	response, err := h.agentService.ListAgents(r.Context(), page, pageSize)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to list agents")
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// UpdateAgent handles PUT /api/v1/agents/{id}
func (h *AgentHandler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "agent id is required")
		return
	}

	var req models.UpdateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	agent, err := h.agentService.UpdateAgent(r.Context(), id, &req)
	if err != nil {
		if err == services.ErrAgentNotFound {
			respondWithError(w, http.StatusNotFound, "agent not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to update agent")
		return
	}

	respondWithJSON(w, http.StatusOK, agent)
}

// DeleteAgent handles DELETE /api/v1/agents/{id}
func (h *AgentHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "agent id is required")
		return
	}

	if err := h.agentService.DeleteAgent(r.Context(), id); err != nil {
		if err == services.ErrAgentNotFound {
			respondWithError(w, http.StatusNotFound, "agent not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to delete agent")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

// GetDefaultAgent handles GET /api/v1/agents/default
func (h *AgentHandler) GetDefaultAgent(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	agent, err := h.agentService.GetDefaultAgent(r.Context(), userID)
	if err != nil {
		if err == services.ErrAgentNotFound {
			respondWithError(w, http.StatusNotFound, "no default agent found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to get default agent")
		return
	}

	respondWithJSON(w, http.StatusOK, agent)
}