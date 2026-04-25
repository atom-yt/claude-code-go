// Package apiserver provides HTTP API server for external Agent access.
//
// This allows external services (web apps, mobile apps, etc.) to interact
// with Claude Code agents via REST API.
//
// Architecture:
//
//   External Services → API Server → Agent Manager → Agent Instances
//
// The API Server:
// - Handles REST API requests
// - Manages session lifecycle
// - Routes requests to appropriate agent instance
// - Supports multi-instance deployment (load balancing)
package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/atom-yt/claude-code-go/internal/agent"
	"github.com/atom-yt/claude-code-go/internal/config"
	"github.com/gorilla/mux"
)

// DeploymentMode specifies how agents are deployed.
type DeploymentMode string

const (
	// ModeSingle runs a single agent instance for all requests.
	ModeSingle DeploymentMode = "single"

	// ModePerSession runs a new agent per session.
	ModePerSession DeploymentMode = "per-session"

	// ModePool uses a pool of agent instances with load balancing.
	ModePool DeploymentMode = "pool"
)

// Config holds API Server configuration.
type Config struct {
	// Server configuration
	Addr           string        // Listen address (e.g., ":8080")
	ReadTimeout    time.Duration // Request read timeout
	WriteTimeout   time.Duration // Response write timeout
	IdleTimeout    time.Duration // Connection idle timeout

	// Agent configuration
	Model          string   // AI model to use
	Provider       string   // API provider
	BaseURL        string   // Custom base URL
	APIKey         string   // API key
	SystemPrompt   string   // System prompt for all sessions

	// MCP configuration
	MCPServers     map[string]config.MCPServerConfig // MCP server configs
	MCPTrustLevels  map[string]string                // MCP trust levels

	// Permissions configuration
	Permissions    config.PermissionsConfig // Permission rules

	// Hooks configuration
	Hooks          map[string][]config.HookMatcherConfig // Hook definitions

	// Deployment mode
	DeploymentMode DeploymentMode // "single", "per-session", or "pool"
	PoolSize       int         // Number of agents in pool (for pool mode)

	// Features
	EnableSSE      bool // Enable Server-Sent Events for streaming
	EnableWS      bool // Enable WebSocket support
	EnableAuth     bool // Enable API authentication
	APIKeyAuth     string // Required API key if auth enabled (different from APIKey above)

	// CORS
	AllowedOrigins []string // CORS allowed origins

	// Working directory (for task management)
	WorkingDir     string // Working directory for tools
}

// Server is the API server instance.
type Server struct {
	config        *Config
	router        *mux.Router
	agentManager  *AgentManager
	httpServer    *http.Server
	startTime     time.Time

	mu            sync.RWMutex
	sessions      map[string]*SessionHandle
}

// SessionHandle represents an active API session.
type SessionHandle struct {
	ID         string
	Agent      *agent.Agent
	CreatedAt  time.Time
	LastActive time.Time
	UserAgent  string      // HTTP User-Agent
	RemoteAddr string      // Client address
}

// NewServer creates a new API server.
func NewServer(cfg *Config) *Server {
	router := mux.NewRouter()

	s := &Server{
		config:   cfg,
		router:   router,
		sessions: make(map[string]*SessionHandle),
	}

	s.setupRoutes()
	s.setupAgentManager()

	s.httpServer = &http.Server{
		Addr:         s.config.Addr,
		Handler:      s.handlerWithMiddleware(s.router),
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	return s
}

// setupRoutes sets up all API routes.
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Session management
	s.router.HandleFunc("/api/v1/sessions", s.handleListSessions).Methods("GET")
	s.router.HandleFunc("/api/v1/sessions", s.handleCreateSession).Methods("POST")
	s.router.HandleFunc("/api/v1/sessions/{id}", s.handleGetSession).Methods("GET")
	s.router.HandleFunc("/api/v1/sessions/{id}", s.handleDeleteSession).Methods("DELETE")

	// Stats
	s.router.HandleFunc("/api/v1/stats", s.handleGetStats).Methods("GET")

	// Chat (REST API)
	s.router.HandleFunc("/api/v1/chat/completions", s.handleChatCompletion).Methods("POST")

	// Chat (Server-Sent Events for streaming)
	if s.config.EnableSSE {
		s.router.HandleFunc("/api/v1/chat/stream", s.handleChatSSE).Methods("POST")
	}

	// Chat (WebSocket)
	if s.config.EnableWS {
		s.router.HandleFunc("/api/v1/chat/ws", s.handleChatWebSocket)
	}
}

// setupAgentManager initializes agent manager.
func (s *Server) setupAgentManager() {
	s.agentManager = &AgentManager{
		mode:    s.config.DeploymentMode,
		poolIdx:  0,
		factory: s.createAgent,
	}

	switch s.config.DeploymentMode {
	case ModePool:
		// Create pool of agents
		if s.config.PoolSize <= 0 {
			s.config.PoolSize = 4 // Default pool size
		}
		// Initialize pool
		if err := s.agentManager.InitializePool(context.Background(), s.config.PoolSize); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to initialize agent pool: %v\n", err)
		}
	}
}

// handlerWithMiddleware wraps router with common middleware.
func (s *Server) handlerWithMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS handling
		s.handleCORS(w, r)

		// Authentication
		if s.config.EnableAuth {
			if !s.authenticateRequest(r) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Logging
		s.logRequest(r)

		// Call handler
		h.ServeHTTP(w, r)
	})
}

// handleCORS handles CORS headers.
func (s *Server) handleCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}

	// Check if origin is allowed
	allowed := false
	for _, allowedOrigin := range s.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			allowed = true
			break
		}
	}

	if allowed {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
}

// authenticateRequest validates API key.
func (s *Server) authenticateRequest(r *http.Request) bool {
	apiKey := r.Header.Get("Authorization")
	if apiKey == "" {
		apiKey = r.URL.Query().Get("api_key")
	}

	// Strip "Bearer " prefix if present
	if len(apiKey) > 7 && apiKey[:7] == "Bearer " {
		apiKey = apiKey[7:]
	}

	return apiKey == s.config.APIKeyAuth
}

// logRequest logs incoming requests.
func (s *Server) logRequest(r *http.Request) {
	// In production, use proper logging
	fmt.Printf("[API] %s %s from %s\n", r.Method, r.URL.Path, r.RemoteAddr)
}

// Start starts API server.
func (s *Server) Start(ctx context.Context) error {
	s.startTime = time.Now()
	fmt.Printf("Starting API server on %s\n", s.config.Addr)
	fmt.Printf("Deployment mode: %s\n", s.config.DeploymentMode)

	errCh := make(chan error, 1)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return s.Shutdown(ctx)
	case err := <-errCh:
		return err
	}
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down API server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	return nil
}

// GetAgent gets an agent instance based on deployment mode.
func (s *Server) GetAgent(ctx context.Context, sessionID string) (*agent.Agent, error) {
	return s.agentManager.GetAgent(ctx, sessionID)
}

// ReleaseAgent releases an agent back to the pool.
func (s *Server) ReleaseAgent(agent *agent.Agent, sessionID string) {
	s.agentManager.ReleaseAgent(agent, sessionID)
}

// Stats returns server statistics.
func (s *Server) Stats() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]any{
		"active_sessions": len(s.sessions),
		"deployment_mode": s.config.DeploymentMode,
		"pool_size":     s.config.PoolSize,
	}
}
