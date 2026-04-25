package feishu

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/atom-yt/claude-code-go/internal/agent"
	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/tools"
)

// Session represents a conversation session with a Feishu chat.
type Session struct {
	mu          sync.RWMutex
	ID          string
	Agent       *agent.Agent
	History     []api.Message
	LastActive  time.Time
	Context     *SessionContext
}

// SessionContext contains session metadata.
type SessionContext struct {
	UserID     string
	UserName   string
	TenantKey  string
	ChatID     string
	CreatedAt  time.Time
	Model      string
	Provider   string
}

// SessionManager manages Feishu sessions.
type SessionManager struct {
	mu         sync.RWMutex
	sessions   map[string]*Session
	config     *Config
	toolReg    *tools.Registry
	apiClient  api.Streamer
	model      string
	provider   string
	ctx        context.Context
	cancel     context.CancelFunc
	persistence *SessionPersistence
}

// SessionPersistence handles session persistence.
type SessionPersistence struct {
	enabled bool
	path    string
}

// NewSessionManager creates a new session manager.
func NewSessionManager(ctx context.Context, cfg *Config, toolReg *tools.Registry, apiClient api.Streamer, model, provider string) *SessionManager {
	ctx, cancel := context.WithCancel(ctx)

	// Use model from config if not provided
	if model == "" {
		model = cfg.Model
	}

	sm := &SessionManager{
		sessions:  make(map[string]*Session),
		config:    cfg,
		toolReg:   toolReg,
		apiClient: apiClient,
		model:     model,
		provider:  provider,
		ctx:       ctx,
		cancel:    cancel,
		persistence: &SessionPersistence{
			enabled: cfg.PersistSessions,
			path:    ".feishu/sessions",
		},
	}

	return sm
}

// GetOrCreate gets an existing session or creates a new one.
func (sm *SessionManager) GetOrCreate(chatID, userID, userName, tenantKey string) (*Session, error) {
	sm.mu.RLock()
	session, exists := sm.sessions[chatID]
	sm.mu.RUnlock()

	if exists {
		// Update session activity
		sm.mu.Lock()
		session.LastActive = time.Now()
		session.Context.UserID = userID
		session.Context.UserName = userName
		sm.mu.Unlock()

		return session, nil
	}

	// Check if we have too many sessions
	sm.mu.RLock()
	if len(sm.sessions) >= sm.config.MaxSessions {
		sm.mu.RUnlock()
		return nil, ErrMaxSessionsExceeded
	}
	sm.mu.RUnlock()

	// Create new session
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Double-check after acquiring write lock
	if session, exists := sm.sessions[chatID]; exists {
		return session, nil
	}

	// Create agent instance
	ag := agent.New(sm.apiClient, sm.model, sm.provider, sm.toolReg, nil, nil)

	session = &Session{
		ID:         chatID,
		Agent:      ag,
		History:    make([]api.Message, 0, sm.config.MaxHistorySize),
		LastActive: time.Now(),
		Context: &SessionContext{
			UserID:    userID,
			UserName:  userName,
			TenantKey: tenantKey,
			ChatID:    chatID,
			CreatedAt: time.Now(),
			Model:     sm.model,
			Provider:  sm.provider,
		},
	}

	sm.sessions[chatID] = session

	// Load persisted history if enabled
	if sm.persistence.enabled {
		sm.loadSessionHistory(session)
	}

	return session, nil
}

// Get retrieves a session by ID.
func (sm *SessionManager) Get(chatID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, ok := sm.sessions[chatID]
	return session, ok
}

// Remove removes a session.
func (sm *SessionManager) Remove(chatID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[chatID]; exists {
		// Persist history if enabled
		if sm.persistence.enabled {
			sm.saveSessionHistory(session)
		}

		delete(sm.sessions, chatID)
	}
}

// CleanupInactive removes inactive sessions.
func (sm *SessionManager) CleanupInactive() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for chatID, session := range sm.sessions {
		if now.Sub(session.LastActive) > sm.config.SessionTimeout {
			if sm.persistence.enabled {
				sm.saveSessionHistory(session)
			}

			delete(sm.sessions, chatID)
		}
	}
}

// AddMessage adds a message to session history.
func (sm *SessionManager) AddMessage(chatID string, msg api.Message) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[chatID]
	if !exists {
		return ErrSessionNotFound
	}

	session.History = append(session.History, msg)

	// Trim history if needed
	if len(session.History) > sm.config.MaxHistorySize {
		session.History = session.History[len(session.History)-sm.config.MaxHistorySize:]
	}

	return nil
}

// GetHistory retrieves session history.
func (sm *SessionManager) GetHistory(chatID string) ([]api.Message, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[chatID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	// Return a copy to avoid race conditions
	history := make([]api.Message, len(session.History))
	copy(history, session.History)

	return history, nil
}

// ClearHistory clears a session's history.
func (sm *SessionManager) ClearHistory(chatID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[chatID]
	if !exists {
		return ErrSessionNotFound
	}

	session.History = make([]api.Message, 0, sm.config.MaxHistorySize)
	return nil
}

// ActiveCount returns the number of active sessions.
func (sm *SessionManager) ActiveCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// ListSessions returns all active session IDs.
func (sm *SessionManager) ListSessions() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ids := make([]string, 0, len(sm.sessions))
	for id := range sm.sessions {
		ids = append(ids, id)
	}
	return ids
}

// Close closes all sessions and stops the manager.
func (sm *SessionManager) Close() {
	sm.cancel()

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, session := range sm.sessions {
		if sm.persistence.enabled {
			sm.saveSessionHistory(session)
		}
	}

	sm.sessions = make(map[string]*Session)
}

// loadSessionHistory loads history from disk.
func (sm *SessionManager) loadSessionHistory(session *Session) {
	// Implementation would load from JSON file
	// For now, this is a placeholder
}

// saveSessionHistory saves history to disk.
func (sm *SessionManager) saveSessionHistory(session *Session) {
	// Implementation would save to JSON file
	// For now, this is a placeholder
}

// Session errors
var (
	ErrSessionNotFound      = fmt.Errorf("session not found")
	ErrMaxSessionsExceeded = fmt.Errorf("maximum sessions exceeded")
)