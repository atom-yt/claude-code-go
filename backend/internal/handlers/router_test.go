package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
	"github.com/stretchr/testify/assert"
)

// MockAuthService is a mock implementation of AuthServiceInterface
type MockAuthService struct {
	validateTokenFunc func(token string) (*auth.Claims, error)
}

func (m *MockAuthService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For testing, we just pass through
		next.ServeHTTP(w, r)
	})
}

func TestRouter_HealthEndpoint(t *testing.T) {
	cfg := &Config{
		AuthService:      &auth.Service{},
		AgentService:     &MockAgentService{},
		SessionService:   &MockSessionService{},
		MessageService:   &MockMessageService{},
		SkillService:     &MockSkillService{},
		ArtifactService:  &MockArtifactService{},
		ScheduleService:  &MockScheduleService{},
		KnowledgeService: &MockKnowledgeService{},
	}

	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "healthy")
}

func TestRouter_CORSOptions(t *testing.T) {
	cfg := &Config{
		AuthService:      &auth.Service{},
		AgentService:     &MockAgentService{},
		SessionService:   &MockSessionService{},
		MessageService:   &MockMessageService{},
		SkillService:     &MockSkillService{},
		ArtifactService:  &MockArtifactService{},
		ScheduleService:  &MockScheduleService{},
		KnowledgeService: &MockKnowledgeService{},
	}

	router := NewRouter(cfg)

	// Test OPTIONS request for CORS preflight
	req := httptest.NewRequest("OPTIONS", "/api/v1/sessions", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Check CORS headers
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
}

func TestRouter_PublicRoutes(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"health check", "GET", "/health"},
		{"auth login", "POST", "/api/v1/auth/login"},
		{"auth register", "POST", "/api/v1/auth/register"},
		{"auth refresh", "POST", "/api/v1/auth/refresh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				AuthService:      &auth.Service{},
				AgentService:     &MockAgentService{},
				SessionService:   &MockSessionService{},
				MessageService:   &MockMessageService{},
				SkillService:     &MockSkillService{},
				ArtifactService:  &MockArtifactService{},
				ScheduleService:  &MockScheduleService{},
				KnowledgeService: &MockKnowledgeService{},
			}

			router := NewRouter(cfg)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			// Public routes should not return 401
			assert.NotEqual(t, http.StatusUnauthorized, rr.Code)
		})
	}
}

func TestRouter_RouteRegistration(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		shouldExist bool
	}{
		// Session routes
		{"GET /sessions", "GET", "/api/v1/sessions", true},
		{"POST /sessions", "POST", "/api/v1/sessions", true},
		{"GET /sessions/active", "GET", "/api/v1/sessions/active", true},
		{"GET /sessions/{id}", "GET", "/api/v1/sessions/test-id", true},
		{"PUT /sessions/{id}", "PUT", "/api/v1/sessions/test-id", true},
		{"DELETE /sessions/{id}", "DELETE", "/api/v1/sessions/test-id", true},

		// Agent routes
		{"GET /agents", "GET", "/api/v1/agents", true},
		{"POST /agents", "POST", "/api/v1/agents", true},
		{"GET /agents/list", "GET", "/api/v1/agents/list", true},
		{"GET /agents/default", "GET", "/api/v1/agents/default", true},

		// Skill routes
		{"GET /skills", "GET", "/api/v1/skills", true},
		{"POST /skills", "POST", "/api/v1/skills", true},

		// Artifact routes
		{"GET /artifacts", "GET", "/api/v1/artifacts", true},
		{"POST /artifacts", "POST", "/api/v1/artifacts", true},
		{"GET /artifacts/stats", "GET", "/api/v1/artifacts/stats", true},

		// Schedule routes
		{"GET /schedules", "GET", "/api/v1/schedules", true},
		{"POST /schedules", "POST", "/api/v1/schedules", true},

		// Knowledge routes
		{"GET /knowledge", "GET", "/api/v1/knowledge", true},
		{"POST /knowledge", "POST", "/api/v1/knowledge", true},

		// Chat routes
		{"POST /chat", "POST", "/api/v1/chat", true},
		{"POST /chat/stream", "POST", "/api/v1/chat/stream", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				AuthService:      &auth.Service{},
				AgentService:     &MockAgentService{},
				SessionService:   &MockSessionService{},
				MessageService:   &MockMessageService{},
				SkillService:     &MockSkillService{},
				ArtifactService:  &MockArtifactService{},
				ScheduleService:  &MockScheduleService{},
				KnowledgeService: &MockKnowledgeService{},
			}

			router := NewRouter(cfg)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			// Route should be registered (not return 404 or 405 for wrong method)
			if tt.shouldExist {
				assert.NotEqual(t, http.StatusNotFound, rr.Code, "Route %s %s should be registered", tt.method, tt.path)
			}
		})
	}
}

func TestRouter_CORSHeaders(t *testing.T) {
	cfg := &Config{
		AuthService:      &auth.Service{},
		AgentService:     &MockAgentService{},
		SessionService:   &MockSessionService{},
		MessageService:   &MockMessageService{},
		SkillService:     &MockSkillService{},
		ArtifactService:  &MockArtifactService{},
		ScheduleService:  &MockScheduleService{},
		KnowledgeService: &MockKnowledgeService{},
	}

	router := NewRouter(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check that CORS headers are set
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestRouter_WebSocketRoute(t *testing.T) {
	cfg := &Config{
		AuthService:      &auth.Service{},
		AgentService:     &MockAgentService{},
		SessionService:   &MockSessionService{},
		MessageService:   &MockMessageService{},
		SkillService:     &MockSkillService{},
		ArtifactService:  &MockArtifactService{},
		ScheduleService:  &MockScheduleService{},
		KnowledgeService: &MockKnowledgeService{},
	}

	router := NewRouter(cfg)

	// WebSocket route should be registered
	req := httptest.NewRequest("GET", "/ws/chat", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// WebSocket without session_id should return 400 (not 404)
	// Note: This test may return 401 (unauthorized) because WebSocket also checks auth
	assert.NotEqual(t, http.StatusNotFound, rr.Code, "WebSocket route should be registered")
}