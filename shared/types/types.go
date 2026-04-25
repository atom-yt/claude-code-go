package types

// User represents a platform user.
type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"` // Never expose in JSON
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// Session represents an AI conversation session.
type Session struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Title     string `json:"title"`
	Model     string `json:"model"`
	Provider  string `json:"provider"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Message represents a single message in a session.
type Message struct {
	ID        string      `json:"id"`
	SessionID string      `json:"sessionId"`
	Role      string      `json:"role"` // "user", "assistant", "tool"
	Content   string      `json:"content"`
	ToolCalls []ToolCall  `json:"toolCalls,omitempty"`
	CreatedAt string      `json:"createdAt"`
}

// ToolCall represents a tool invocation.
type ToolCall struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Input   map[string]any `json:"input"`
	Output  string         `json:"output,omitempty"`
	IsError bool           `json:"isError,omitempty"`
}

// Agent represents an agent configuration.
type Agent struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"userId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	SystemPrompt string                `json:"systemPrompt"`
	Tools       []string               `json:"tools"`
	Config      map[string]any         `json:"config,omitempty"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
}

// AuthToken represents authentication tokens.
type AuthToken struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"` // Seconds until expiration
	TokenType    string `json:"tokenType"` // Always "Bearer"
}

// LoginRequest is the request body for login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is the request body for registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateSessionRequest is the request body for creating a session.
type CreateSessionRequest struct {
	Title    string `json:"title,omitempty"`
	Model    string `json:"model,omitempty"`
	Provider string `json:"provider,omitempty"`
}

// SendMessageRequest is the request body for sending a message.
type SendMessageRequest struct {
	Content   string `json:"content"`
	Stream    bool   `json:"stream,omitempty"`
}

// ErrorResponse represents an API error.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Database string `json:"database"`
}