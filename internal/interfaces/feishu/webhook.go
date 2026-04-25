package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// WebhookServer handles Feishu webhook callbacks.
type WebhookServer struct {
	config    *Config
	server    *http.Server
	router    *mux.Router
	handler   *EventHandler
	verifier  *SignatureVerifier
}

// NewWebhookServer creates a new webhook server.
func NewWebhookServer(cfg *Config, handler *EventHandler) *WebhookServer {
	verifier := NewSignatureVerifier(cfg.VerificationToken, cfg.WebhookSecret)

	ws := &WebhookServer{
		config:   cfg,
		handler:  handler,
		verifier: verifier,
	}

	ws.setupRouter()

	return ws
}

// setupRouter sets up HTTP routes.
func (ws *WebhookServer) setupRouter() {
	ws.router = mux.NewRouter()

	// Main webhook endpoint
	ws.router.HandleFunc(ws.config.WebhookPath, ws.handleWebhook).Methods("POST")

	// Health check endpoint
	ws.router.HandleFunc("/health", ws.handleHealth).Methods("GET")

	// Ready endpoint
	ws.router.HandleFunc("/ready", ws.handleReady).Methods("GET")
}

// Start starts the webhook server.
func (ws *WebhookServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", ws.config.WebhookPort)

	ws.server = &http.Server{
		Addr:         addr,
		Handler:      ws.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)

	// Start server in goroutine
	go func() {
		if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("webhook server error: %w", err)
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		return ws.Shutdown(ctx)
	case err := <-errCh:
		return err
	}
}

// Shutdown gracefully shuts down the webhook server.
func (ws *WebhookServer) Shutdown(ctx context.Context) error {
	if ws.server == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return ws.server.Shutdown(shutdownCtx)
}

// handleWebhook handles incoming webhook events.
func (ws *WebhookServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Verify signature
	if valid, err := ws.verifier.VerifySignature(r); !valid {
		if err != nil {
			ws.logError("signature verification failed: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Read request body
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		ws.logError("failed to decode event: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Handle different event types
	switch event.EventType {
	case "url_verification":
		ws.handleURLVerification(w, &event)
	case "event_callback":
		ws.handleEventCallback(ctx, w, &event)
	default:
		ws.logError("unknown event type: %s", event.EventType)
		w.WriteHeader(http.StatusBadRequest)
	}
}

// handleURLVerification handles Feishu URL verification challenge.
func (ws *WebhookServer) handleURLVerification(w http.ResponseWriter, event *Event) {
	// Verify token
	if !ws.verifier.VerifyToken(event.Token) {
		ws.logError("token verification failed")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Return challenge to complete verification
	response := map[string]string{
		"challenge": event.Challenge,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	ws.logInfo("URL verification completed")
}

// handleEventCallback handles Feishu event callbacks.
func (ws *WebhookServer) handleEventCallback(ctx context.Context, w http.ResponseWriter, event *Event) {
	if event.Event.Type != "message" {
		// Non-message event, acknowledge without processing
		w.WriteHeader(http.StatusOK)
		return
	}

	// Convert to incoming message
	msg, err := ws.parseMessageEvent(event)
	if err != nil {
		ws.logError("failed to parse message event: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Process message asynchronously
	go ws.handler.HandleMessage(ctx, msg)

	// Acknowledge receipt
	w.WriteHeader(http.StatusOK)
}

// parseMessageEvent parses a message event into IncomingMessage.
func (ws *WebhookServer) parseMessageEvent(event *Event) (*IncomingMessage, error) {
	if event.Event.Message == nil {
		return nil, fmt.Errorf("no message in event")
	}

	msgEvent := event.Event.Message

	// Parse message content
	var content string
	var images []ImageAttachment
	var files []FileAttachment

	switch msgEvent.Message.MessageType {
	case "text":
		var textContent TextMessageContent
		if err := json.Unmarshal(msgEvent.Message.Content, &textContent); err != nil {
			return nil, fmt.Errorf("failed to parse text content: %w", err)
		}
		content = textContent.Text

	case "image":
		var imgContent ImageMessageContent
		if err := json.Unmarshal(msgEvent.Message.Content, &imgContent); err != nil {
			return nil, fmt.Errorf("failed to parse image content: %w", err)
		}
		images = append(images, ImageAttachment{
			ImageKey: imgContent.ImageKey,
		})

	case "file":
		var fileContent FileMessageContent
		if err := json.Unmarshal(msgEvent.Message.Content, &fileContent); err != nil {
			return nil, fmt.Errorf("failed to parse file content: %w", err)
		}
		files = append(files, FileAttachment{
			FileKey: fileContent.FileKey,
		})

	default:
		return nil, fmt.Errorf("unsupported message type: %s", msgEvent.Message.MessageType)
	}

	// Convert mentions
	mentions := make([]Mention, 0)
	for _, m := range msgEvent.Mentions {
		mentions = append(mentions, m)
	}

	// Build incoming message
	msg := &IncomingMessage{
		ChatID:     msgEvent.ChatID,
		MessageID:   msgEvent.MessageID,
		RootID:      msgEvent.RootID,
		ParentID:    msgEvent.ParentID,
		UserID:      msgEvent.Sender.UserID,
		UserName:    msgEvent.Sender.UserID, // Could fetch actual name
		TenantKey:   msgEvent.Sender.TenantKey,
		Content:     content,
		Images:      images,
		Files:       files,
		Timestamp:   time.Unix(msgEvent.CreateTime, 0),
	}

	return msg, nil
}

// handleHealth handles health check requests.
func (ws *WebhookServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// handleReady handles readiness check requests.
func (ws *WebhookServer) handleReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
}

// Address returns the server's address.
func (ws *WebhookServer) Address() string {
	if ws.server == nil {
		return ""
	}
	return ws.server.Addr
}

// logError logs an error message.
func (ws *WebhookServer) logError(format string, args ...any) {
	// In production, use proper logger
	// For now, just print
	fmt.Printf("[WebhookServer] ERROR: "+format+"\n", args...)
}

// logInfo logs an info message.
func (ws *WebhookServer) logInfo(format string, args ...any) {
	// In production, use proper logger
	fmt.Printf("[WebhookServer] INFO: "+format+"\n", args...)
}