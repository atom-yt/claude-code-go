package apiserver

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/atom-yt/claude-code-go/internal/agent"
	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/gorilla/mux"
)

// handleHealth handles health check requests.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	stats := s.agentManager.Stats()

	response := HealthResponse{
		Status: "healthy",
		Uptime: time.Since(s.startTime).String(),
		Stats:  stats,
	}

	writeJSON(w, http.StatusOK, response)
}

var startTime = time.Now()

// handleCreateSession creates a new session.
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sessionID := newSessionID()
	createdAt := time.Now().Format(time.RFC3339)

	// Get or create agent
	agentInstance, err := s.GetAgent(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create agent: %v", err))
		return
	}

	s.mu.Lock()
	s.sessions[sessionID] = &SessionHandle{
		ID:        sessionID,
		Agent:     agentInstance,
		CreatedAt: time.Now(),
		LastActive: time.Now(),
		UserAgent: r.UserAgent(),
		RemoteAddr: r.RemoteAddr,
	}
	s.mu.Unlock()

	response := CreateSessionResponse{
		SessionID: sessionID,
		CreatedAt: createdAt,
	}

	writeJSON(w, http.StatusCreated, response)
}

// handleGetSession gets session information.
func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	response := SessionInfo{
		SessionID:   session.ID,
		CreatedAt:   session.CreatedAt.Format(time.RFC3339),
		LastActive:  session.LastActive.Format(time.RFC3339),
		MessageCount: 0,
		UserAgent:   session.UserAgent,
		RemoteAddr:  session.RemoteAddr,
	}

	writeJSON(w, http.StatusOK, response)
}

// handleDeleteSession deletes a session.
func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	s.mu.Lock()
	session, exists := s.sessions[sessionID]
	if !exists {
		s.mu.Unlock()
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	delete(s.sessions, sessionID)
	s.mu.Unlock()

	// Release agent back to pool
	s.ReleaseAgent(session.Agent, sessionID)

	w.WriteHeader(http.StatusNoContent)
}

// handleChatCompletion handles chat completion requests (REST API).
func (s *Server) handleChatCompletion(w http.ResponseWriter, r *http.Request) {
	var req ChatCompletionRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SessionID == "" {
		writeError(w, http.StatusBadRequest, "session_id is required")
		return
	}

	// Get agent
	agentInstance, err := s.GetAgent(r.Context(), req.SessionID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	// Update last active time
	s.mu.Lock()
	if session, exists := s.sessions[req.SessionID]; exists {
		session.LastActive = time.Now()
	}
	s.mu.Unlock()

	// Build messages
	eventCh := agentInstance.Query(r.Context(), req.Message)

	var fullResponse strings.Builder
	var usage *api.Usage

	for event := range eventCh {
		switch event.Type {
		case agent.EventTextDelta:
			fullResponse.WriteString(event.Text)

		case agent.EventToolCall:
			// Could include tool use in response

		case agent.EventToolResult:
			// Could include tool result

		case agent.EventError:
			writeError(w, http.StatusInternalServerError, event.Error.Error())
			return

		case agent.EventDone:
			usage = event.Usage
			// Done, exit loop
		}
	}

	response := ChatCompletionResponse{
		SessionID: req.SessionID,
		MessageID: generateMessageID(),
		Content:   fullResponse.String(),
		Role:      "assistant",
		Usage:     usage,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleChatSSE handles chat requests with Server-Sent Events for streaming.
func (s *Server) handleChatSSE(w http.ResponseWriter, r *http.Request) {
	var req ChatCompletionRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Force stream mode for SSE endpoint
	req.Stream = true

	if req.SessionID == "" {
		writeError(w, http.StatusBadRequest, "session_id is required")
		return
	}

	// Get agent
	agentInstance, err := s.GetAgent(r.Context(), req.SessionID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	// Create response channel
	responseCh := make(chan ChatChunk, 100)

	// Start streaming in goroutine
	go func() {
		defer close(responseCh)

		eventCh := agentInstance.Query(r.Context(), req.Message)

		for event := range eventCh {
			chunk := ChatChunk{
				SessionID: req.SessionID,
			}

			switch event.Type {
			case agent.EventTextDelta:
				chunk.Type = "delta"
				chunk.Content = event.Text

			case agent.EventToolCall:
				chunk.Type = "tool_call"
				chunk.Content = fmt.Sprintf("Tool: %s", event.ToolName)

			case agent.EventToolResult:
				chunk.Type = "tool_result"
				chunk.Content = event.ToolOutput

			case agent.EventError:
				chunk.Type = "error"
				chunk.Error = event.Error.Error()

			case agent.EventDone:
				chunk.Type = "done"
				chunk.Usage = event.Usage

			default:
				continue
			}

			responseCh <- chunk

			if event.Type == agent.EventDone || event.Type == agent.EventError {
				return
			}
		}
	}()

	// Write SSE response
	writeSSE(w, responseCh)
}

// handleChatWebSocket handles WebSocket chat connections.
func (s *Server) handleChatWebSocket(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket support
	writeError(w, http.StatusNotImplemented, "WebSocket not yet implemented")
}

// handleListSessions lists all active sessions.
func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]SessionInfo, 0, len(s.sessions))

	for _, session := range s.sessions {
		sessions = append(sessions, SessionInfo{
			SessionID:   session.ID,
			CreatedAt:   session.CreatedAt.Format(time.RFC3339),
			LastActive:  session.LastActive.Format(time.RFC3339),
			MessageCount: 0,
			UserAgent:   session.UserAgent,
			RemoteAddr:  session.RemoteAddr,
		})
	}

	response := SessionsListResponse{
		Sessions: sessions,
		Total:    len(sessions),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleGetStats returns server statistics.
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	stats := s.Stats()
	s.mu.RUnlock()

	response := map[string]any{
		"server":  stats,
		"agents":  s.agentManager.Stats(),
		"sessions": len(s.sessions),
	}

	writeJSON(w, http.StatusOK, response)
}