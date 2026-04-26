package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/atom-yt/claude-code-go/pkg/agent"
)

const (
	// WebSocket upgrader configuration
	wsReadBufferSize  = 1024
	wsWriteBufferSize = 1024
)

// ChatHandler handles chat requests
type ChatHandler struct {
	agentFactory agent.AgentFactory
	upgrader    *websocket.Upgrader
}

// NewChatHandler creates a new chat handler
func NewChatHandler(agentFactory agent.AgentFactory) *ChatHandler {
	return &ChatHandler{
		agentFactory: agentFactory,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  wsReadBufferSize,
			WriteBufferSize: wsWriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

// ChatRequest represents a chat request via HTTP
type ChatRequest struct {
	SessionID   string          `json:"session_id"`
	Message     string          `json:"message"`
	Stream      bool            `json:"stream"`
	AgentConfig *agent.Config    `json:"agent_config,omitempty"`
}

// HandleChat handles REST chat request
func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Set session ID if not provided
	if req.SessionID == "" {
		req.SessionID = r.URL.Query().Get("session_id")
	}

	// Only streaming is supported
	if !req.Stream {
		respondWithError(w, http.StatusBadRequest, "only streaming mode is supported")
		return
	}

	h.handleStreamChat(w, r, &req)
}

// handleStreamChat handles streaming chat with SSE
func (h *ChatHandler) handleStreamChat(w http.ResponseWriter, r *http.Request, req *ChatRequest) {
	ctx := r.Context()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create agent configuration
	agentCfg := h.buildAgentConfig(req)
	agt, err := h.agentFactory.Create(ctx, agentCfg)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Query agent
	eventCh := agt.Chat(ctx, req.Message)

	// Stream events
	flusher, ok := w.(http.Flusher)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	for event := range eventCh {
		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			continue
		}

		// Write SSE format
		w.Write([]byte("data: "))
		w.Write(data)
		w.Write([]byte("\n\n"))
		flusher.Flush()
	}
}

// HandleWebSocket handles WebSocket chat connection
func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get session ID from query
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		respondWithError(w, http.StatusBadRequest, "session_id required")
		return
	}

	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket connected: user=%s session=%s", userID, sessionID)

	// Handle WebSocket messages
	ctx := r.Context()
	h.handleWSMessages(ctx, conn, sessionID, userID)
}

// handleWSMessages handles incoming WebSocket messages
func (h *ChatHandler) handleWSMessages(ctx context.Context, conn *websocket.Conn, sessionID, userID string) {
	// Send connected event
	sendWSEvent(conn, map[string]interface{}{
		"type": "connected",
		"data": map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
		},
	})

	for {
		select {
		case <-ctx.Done():
			return

		default:
			// Read message from client
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}

			// Only handle text messages
			if messageType != websocket.TextMessage {
				continue
			}

			// Parse request
			var req ChatRequest
			if err := json.Unmarshal(data, &req); err != nil {
				log.Printf("Failed to unmarshal WebSocket message: %v", err)
				sendWSError(conn, "invalid request")
				continue
			}

			// Set session ID from connection
			req.SessionID = sessionID

			// Enable streaming for WebSocket
			req.Stream = true

			// Create agent configuration
			agentCfg := h.buildAgentConfig(&req)

			// Create agent
			agt, err := h.agentFactory.Create(ctx, agentCfg)
			if err != nil {
				log.Printf("Agent creation error: %v", err)
				sendWSError(conn, err.Error())
				continue
			}

			// Process chat
			eventCh := agt.Chat(ctx, req.Message)

			// Stream events to client
			for event := range eventCh {
				sendWSEvent(conn, map[string]interface{}{
					"type": string(event.Type),
					"data": map[string]interface{}{
						"text":   event.Text,
						"error":  event.Error,
						"usage":  event.Usage,
					},
				})

				if event.Type == agent.EventTypeDone || event.Type == agent.EventTypeError {
					break
				}
			}
		}
	}
}

// buildAgentConfig builds agent configuration from request
func (h *ChatHandler) buildAgentConfig(req *ChatRequest) *agent.Config {
	cfg := &agent.Config{}
	if req.AgentConfig != nil {
		cfg = req.AgentConfig
	}
	return cfg
}

// sendWSEvent sends an event to WebSocket client
func sendWSEvent(conn *websocket.Conn, data interface{}) error {
	return conn.WriteJSON(map[string]interface{}{
		"type": "event",
		"data": data,
	})
}

// sendWSError sends an error to WebSocket client
func sendWSError(conn *websocket.Conn, message string) error {
	return conn.WriteJSON(map[string]interface{}{
		"type":  "error",
		"error": message,
	})
}