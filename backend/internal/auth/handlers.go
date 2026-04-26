package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Handler methods
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	Refresh(refreshToken string) (*RefreshResponse, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	// Middleware method
	AuthMiddleware(next http.Handler) http.Handler
}

// Handler provides HTTP handlers for authentication
type Handler struct {
	service AuthService
}

// NewHandler creates a new auth handler
func NewHandler(service AuthService) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers authentication routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.HandleRegister)
	mux.HandleFunc("POST /api/v1/auth/login", h.HandleLogin)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.HandleRefresh)
	mux.HandleFunc("GET /api/v1/auth/me", h.HandleMe)
}

// handleRegister handles user registration
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	resp, err := h.service.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			RespondWithError(w, http.StatusConflict, "user already exists")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	RespondWithJSON(w, http.StatusCreated, resp)
}

// handleLogin handles user login
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	resp, err := h.service.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrInvalidCredentials) {
			RespondWithError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	RespondWithJSON(w, http.StatusOK, resp)
}

// handleRefresh handles token refresh
func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	resp, err := h.service.Refresh(req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrExpiredToken) {
			RespondWithError(w, http.StatusUnauthorized, "refresh token expired")
			return
		}
		if errors.Is(err, ErrInvalidToken) {
			RespondWithError(w, http.StatusUnauthorized, "invalid refresh token")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "failed to refresh token")
		return
	}

	RespondWithJSON(w, http.StatusOK, resp)
}

// handleMe handles getting current user info
func (h *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]*User{
		"user": user,
	})
}

// RespondWithError sends an error response
func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}

// RespondWithJSON sends a JSON response
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}