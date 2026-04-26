package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/atom-yt/claude-code-go/pkg/agent"

	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
)

// Router manages all HTTP routes
type Router struct {
	router           *mux.Router
	authService      *auth.Service
	authHandler      *auth.Handler
	agentHandler     *AgentHandler
	sessionHandler   *SessionHandler
	messageHandler   *MessageHandler
	chatHandler      *ChatHandler
	healthHandler    *HealthHandler
	skillHandler     *SkillHandler
	artifactHandler  *ArtifactHandler
	scheduleHandler  *ScheduleHandler
	knowledgeHandler *KnowledgeHandler
}

// Config holds router configuration
type Config struct {
	AuthService      *auth.Service
	AgentService     *services.AgentService
	SessionService   *services.SessionService
	MessageService   *services.MessageService
	SkillService     *services.SkillService
	ArtifactService  *services.ArtifactService
	ScheduleService  *services.ScheduleService
	KnowledgeService *services.KnowledgeService
	AgentFactory     *agent.ConfigFactory
}

// NewRouter creates a new router with all routes configured
func NewRouter(cfg *Config) *Router {
	r := mux.NewRouter()

	router := &Router{
		router:           r,
		authService:      cfg.AuthService,
		authHandler:      auth.NewHandler(cfg.AuthService),
		agentHandler:     NewAgentHandler(cfg.AgentService),
		sessionHandler:   NewSessionHandler(cfg.SessionService),
		messageHandler:   NewMessageHandler(cfg.MessageService),
		chatHandler:      NewChatHandler(cfg.AgentFactory),
		healthHandler:    NewHealthHandler(),
		skillHandler:     NewSkillHandler(cfg.SkillService),
		artifactHandler:  NewArtifactHandler(cfg.ArtifactService),
		scheduleHandler:  NewScheduleHandler(cfg.ScheduleService),
		knowledgeHandler: NewKnowledgeHandler(cfg.KnowledgeService),
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
	protected.Use(r.authService.AuthMiddleware)

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

	// Chat routes (streaming and WebSocket)
	protected.HandleFunc("/chat", r.chatHandler.HandleChat).Methods("POST")
	protected.HandleFunc("/chat/stream", r.chatHandler.HandleChat).Methods("POST")
	r.router.HandleFunc("/ws/chat", r.chatHandler.HandleWebSocket)

	// Skill routes
	protected.HandleFunc("/skills", r.skillHandler.GetUserSkills).Methods("GET")
	protected.HandleFunc("/skills", r.skillHandler.CreateSkill).Methods("POST")
	protected.HandleFunc("/skills/{id}", r.skillHandler.GetSkill).Methods("GET")
	protected.HandleFunc("/skills/{id}", r.skillHandler.UpdateSkill).Methods("PUT")
	protected.HandleFunc("/skills/{id}", r.skillHandler.DeleteSkill).Methods("DELETE")
	protected.HandleFunc("/skills/{id}/toggle", r.skillHandler.ToggleSkill).Methods("PUT")

	// Artifact routes
	protected.HandleFunc("/artifacts", r.artifactHandler.GetUserArtifacts).Methods("GET")
	protected.HandleFunc("/artifacts", r.artifactHandler.CreateArtifact).Methods("POST")
	protected.HandleFunc("/artifacts/stats", r.artifactHandler.GetArtifactStats).Methods("GET")
	protected.HandleFunc("/artifacts/{id}", r.artifactHandler.GetArtifact).Methods("GET")
	protected.HandleFunc("/artifacts/{id}", r.artifactHandler.DeleteArtifact).Methods("DELETE")

	// Schedule routes
	protected.HandleFunc("/schedules", r.scheduleHandler.GetUserSchedules).Methods("GET")
	protected.HandleFunc("/schedules", r.scheduleHandler.CreateSchedule).Methods("POST")
	protected.HandleFunc("/schedules/{id}", r.scheduleHandler.GetSchedule).Methods("GET")
	protected.HandleFunc("/schedules/{id}", r.scheduleHandler.UpdateSchedule).Methods("PUT")
	protected.HandleFunc("/schedules/{id}", r.scheduleHandler.DeleteSchedule).Methods("DELETE")
	protected.HandleFunc("/schedules/{id}/toggle", r.scheduleHandler.ToggleSchedule).Methods("PUT")

	// Knowledge routes
	protected.HandleFunc("/knowledge", r.knowledgeHandler.GetUserKnowledge).Methods("GET")
	protected.HandleFunc("/knowledge", r.knowledgeHandler.CreateKnowledge).Methods("POST")
	protected.HandleFunc("/knowledge/{id}", r.knowledgeHandler.GetKnowledge).Methods("GET")
	protected.HandleFunc("/knowledge/{id}", r.knowledgeHandler.UpdateKnowledge).Methods("PUT")
	protected.HandleFunc("/knowledge/{id}", r.knowledgeHandler.DeleteKnowledge).Methods("DELETE")
}

// GetRouter returns the underlying mux router
func (r *Router) GetRouter() *mux.Router {
	return r.router
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
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