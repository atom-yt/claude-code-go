package handlers

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
)

// Router manages all HTTP routes
type Router struct {
	router        *mux.Router
	authHandler   *auth.Handler
	agentHandler  *AgentHandler
	sessionHandler *SessionHandler
	messageHandler *MessageHandler
	healthHandler *HealthHandler
}

// Config holds router configuration
type Config struct {
	AuthService   *auth.Service
	AgentService  *services.AgentService
	SessionService *services.SessionService
	MessageService *services.MessageService
}

// NewRouter creates a new router with all routes configured
func NewRouter(cfg *Config) *Router {
	r := mux.NewRouter()

	router := &Router{
		router:        r,
		authHandler:   auth.NewHandler(cfg.AuthService),
		agentHandler:  NewAgentHandler(cfg.AgentService),
		sessionHandler: NewSessionHandler(cfg.SessionService),
		messageHandler: NewMessageHandler(cfg.MessageService),
		healthHandler: NewHealthHandler(),
	}

	router.setupRoutes()
	router.setupMiddleware()

	return router
}

// setupMiddleware configures middleware
func (r *Router) setupMiddleware() {
	r.router.Use(loggingMiddleware)
	r.router.Use(corsMiddleware)
	r.router.Use(recoveryMiddleware)
}

// setupRoutes configures all routes
func (r *Router) setupRoutes() {
	// Health check
	r.router.HandleFunc("/health", r.healthHandler.Health).Methods("GET")

	// API routes
	api := r.router.PathPrefix("/api/v1").Subrouter()

	// Auth routes (no authentication required)
	api.HandleFunc("/auth/register", r.authHandler.HandleRegister).Methods("POST")
	api.HandleFunc("/auth/login", r.authHandler.HandleLogin).Methods("POST")
	api.HandleFunc("/auth/refresh", r.authHandler.HandleRefresh).Methods("POST")

	// Protected routes (require authentication)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware)

	// User routes
	protected.HandleFunc("/auth/me", r.authHandler.HandleMe).Methods("GET")

	// Session routes
	protected.HandleFunc("/sessions", r.sessionHandler.GetUserSessions).Methods("GET")
	protected.HandleFunc("/sessions", r.sessionHandler.CreateSession).Methods("POST")
	protected.HandleFunc("/sessions/active", r.sessionHandler.GetActiveSessions).Methods("GET")
	protected.HandleFunc("/sessions/{id}", r.sessionHandler.GetSession).Methods("GET")
	protected.HandleFunc("/sessions/{id}", r.sessionHandler.UpdateSession).Methods("PUT")
	protected.HandleFunc("/sessions/{id}", r.sessionHandler.DeleteSession).Methods("DELETE")
	protected.HandleFunc("/sessions/{id}/archive", r.sessionHandler.ArchiveSession).Methods("POST")

	// Message routes
	protected.HandleFunc("/sessions/{id}/messages", r.messageHandler.GetSessionMessages).Methods("GET")
	protected.HandleFunc("/sessions/{id}/messages", r.messageHandler.CreateMessage).Methods("POST")
	protected.HandleFunc("/sessions/{id}/messages/recent", r.messageHandler.GetRecentMessages).Methods("GET")
	protected.HandleFunc("/messages/{id}", r.messageHandler.GetMessage).Methods("GET")
	protected.HandleFunc("/messages/{id}", r.messageHandler.DeleteMessage).Methods("DELETE")

	// Agent routes
	protected.HandleFunc("/agents", r.agentHandler.GetUserAgents).Methods("GET")
	protected.HandleFunc("/agents", r.agentHandler.CreateAgent).Methods("POST")
	protected.HandleFunc("/agents/list", r.agentHandler.ListAgents).Methods("GET")
	protected.HandleFunc("/agents/default", r.agentHandler.GetDefaultAgent).Methods("GET")
	protected.HandleFunc("/agents/{id}", r.agentHandler.GetAgent).Methods("GET")
	protected.HandleFunc("/agents/{id}", r.agentHandler.UpdateAgent).Methods("PUT")
	protected.HandleFunc("/agents/{id}", r.agentHandler.DeleteAgent).Methods("DELETE")
}

// GetRouter returns the underlying mux router
func (r *Router) GetRouter() *mux.Router {
	return r.router
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// authMiddleware validates JWT tokens and adds user ID to context
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "authorization header required")
			return
		}

		// For now, we'll skip actual validation
		// In production, validate the token using auth.Service.ValidateToken
		// and add the user ID to the context using auth.UserIDKey
		next.ServeHTTP(w, r)
	})
}

// GetUserID extracts user ID from request context
func GetUserID(r *http.Request) string {
	if userID, ok := auth.GetUserID(r.Context()); ok {
		return userID
	}
	return ""
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request
		// In production, use a proper logger
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware recovers from panics
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				respondWithError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}