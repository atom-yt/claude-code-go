package apiserver

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/atom-yt/claude-code-go/internal/api"
)

// API response types
type (
	// ErrorResponse represents an API error response
	ErrorResponse struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	// CreateSessionRequest represents a request to create a session
	CreateSessionRequest struct {
		SystemPrompt string            `json:"system_prompt,omitempty"`
		Model       string            `json:"model,omitempty"`
		MaxHistory  int               `json:"max_history,omitempty"`
	}

	// CreateSessionResponse represents a created session
	CreateSessionResponse struct {
		SessionID string `json:"session_id"`
		CreatedAt string `json:"created_at"`
	}

	// ChatCompletionRequest represents a chat request
	ChatCompletionRequest struct {
		SessionID  string            `json:"session_id"`
		Message    string            `json:"message"`
		Model      string            `json:"model,omitempty"`
		Stream     bool              `json:"stream,omitempty"`
		MaxTokens  int               `json:"max_tokens,omitempty"`
		Temperature float64           `json:"temperature,omitempty"`
	}

	// ChatCompletionResponse represents a chat response (non-streaming)
	ChatCompletionResponse struct {
		SessionID string `json:"session_id"`
		MessageID string `json:"message_id"`
		Content   string `json:"content"`
		Role      string `json:"role"`
		Usage     *api.Usage `json:"usage,omitempty"`
		CreatedAt string `json:"created_at"`
	}

	// ChatChunk represents a streaming chunk
	ChatChunk struct {
		SessionID string `json:"session_id"`
		Type      string `json:"type"` // "delta", "done", "error"
		Content   string `json:"content,omitempty"`
		Role      string `json:"role,omitempty"`
		Usage     *api.Usage `json:"usage,omitempty"`
		Error     string `json:"error,omitempty"`
	}

	// SessionInfo represents session information
	SessionInfo struct {
		SessionID   string    `json:"session_id"`
		CreatedAt   string    `json:"created_at"`
		LastActive  string    `json:"last_active"`
		MessageCount int       `json:"message_count"`
		UserAgent   string    `json:"user_agent,omitempty"`
		RemoteAddr  string    `json:"remote_addr,omitempty"`
	}

	// SessionsListResponse represents a list of sessions
	SessionsListResponse struct {
		Sessions []SessionInfo `json:"sessions"`
		Total     int          `json:"total"`
	}

	// HealthResponse represents health check response
	HealthResponse struct {
		Status      string `json:"status"`
		Version     string `json:"version,omitempty"`
		Uptime      string `json:"uptime,omitempty"`
		Stats       map[string]any `json:"stats,omitempty"`
	}
)

// Helper functions for responses

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, message string) error {
	return writeJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// writeSSE writes Server-Sent Events.
func writeSSE(w http.ResponseWriter, events <-chan ChatChunk) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		// Can't flush, fall back to JSON
		writeError(w, http.StatusNotImplemented, "Server-Sent Events not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	notify := w.(http.CloseNotifier)

	for {
		select {
		case event, ok := <-events:
			if !ok {
				// Channel closed, send done event
				writeSSEvent(w, flusher, ChatChunk{
					Type: "done",
				})
				return
			}

			if err := writeSSEvent(w, flusher, event); err != nil {
				return
			}

		case <-notify.CloseNotify():
			// Client disconnected
			return
		}
	}
}

// writeSSEvent writes a single SSE event.
func writeSSEvent(w http.ResponseWriter, flusher http.Flusher, event ChatChunk) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Write in SSE format: data: <json>\n\n
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}

	flusher.Flush()
	return nil
}

// readJSON reads and decodes JSON from request body.
func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(dst)
}

// newSessionID generates a new session ID.
func newSessionID() string {
	return fmt.Sprintf("sess_%d_%d", time.Now().Unix(), time.Now().Nanosecond()/1000)
}

// generateMessageID generates a new message ID.
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// compressResponse wraps handler with gzip compression.
func compressResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Wrap writer with gzip
		gz := gzip.NewWriter(w)
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(&gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
	})
}

// gzipResponseWriter wraps http.ResponseWriter for gzip.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}