package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// Context key for user ID
type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
)

// AuthMiddleware validates JWT token and adds user context
func (s *Service) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			RespondWithError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		tokenString, err := extractBearerToken(authHeader)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		claims, err := s.ValidateToken(tokenString)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Add user context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware validates JWT token if present
func (s *Service) OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		tokenString, err := extractBearerToken(authHeader)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := s.ValidateToken(tokenString)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Add user context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleMiddleware checks if user has required role
func (s *Service) RoleMiddleware(requiredRole Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := ctxValue[Role](r.Context(), UserRoleKey)
			if !ok {
				RespondWithError(w, http.StatusUnauthorized, "user not authenticated")
				return
			}

			if !hasRole(userRole, requiredRole) {
				RespondWithError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractBearerToken extracts => bearer token from Authorization header
func extractBearerToken(authHeader string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("invalid authorization header format")
	}
	return strings.TrimPrefix(authHeader, prefix), nil
}

// hasRole checks if a user role has sufficient permissions
func hasRole(userRole, requiredRole Role) bool {
	roleHierarchy := map[Role]int{
		RoleAdmin: 3,
		RoleUser:  2,
		RoleGuest: 1,
	}

	return roleHierarchy[userRole] >= roleHierarchy[requiredRole]
}

// ctxValue is a helper function to get typed values from context
func ctxValue[T any](ctx context.Context, key contextKey) (T, bool) {
	value, ok := ctx.Value(key).(T)
	return value, ok
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	return ctxValue[string](ctx, UserIDKey)
}

// GetUserEmail extracts user email from context
func GetUserEmail(ctx context.Context) (string, bool) {
	return ctxValue[string](ctx, UserEmailKey)
}

// GetUserRole extracts user role from context
func GetUserRole(ctx context.Context) (Role, bool) {
	return ctxValue[Role](ctx, UserRoleKey)
}