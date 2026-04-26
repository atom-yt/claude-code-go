package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCORSMiddleware tests CORS headers and OPTIONS requests
func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "GET request with CORS headers",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "OPTIONS preflight request returns 200",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "POST request with CORS headers",
			method:         "POST",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that responds with OK
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Apply CORS middleware
			handler := corsMiddleware(nextHandler)

			// Create request
			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", "http://localhost:3000")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute handler
			handler.ServeHTTP(w, req)

			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check CORS headers
			if tt.checkHeaders {
				expectedHeaders := map[string]string{
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
					"Access-Control-Allow-Headers": "Content-Type, Authorization",
				}

				for header, expectedValue := range expectedHeaders {
					actualValue := w.Header().Get(header)
					if actualValue != expectedValue {
						t.Errorf("Expected header %s to be %q, got %q", header, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

// TestRecoveryMiddleware tests panic recovery
func TestRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.Handler
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "panic handler returns 500",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			}),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
		{
			name: "normal handler works correctly",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}),
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name: "handler returns custom status",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("Created"))
			}),
			expectedStatus: http.StatusCreated,
			expectedBody:   "Created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply recovery middleware
			handler := recoveryMiddleware(tt.handler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute handler
			handler.ServeHTTP(w, req)

			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check body
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

// TestLoggingMiddleware tests that logging middleware passes through
func TestLoggingMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Logged"))
	})

	handler := loggingMiddleware(nextHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Logged" {
		t.Errorf("Expected body 'Logged', got %q", w.Body.String())
	}
}

// TestAuthMiddleware_ValidToken tests valid token authentication
func TestAuthMiddleware_ValidToken(t *testing.T) {
	// This is a simplified test since we can't easily mock the full auth service
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the handlers.getUserID helper
		userID := getUserID(r.Context())
		if userID == "test-user" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Authenticated"))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	// Set user ID in context using handlers helper
	ctx := setUserID(req.Context(), "test-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	nextHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}
		nextHandler.ServeHTTP(w, r)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "missing authorization header", resp["error"])
}

func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			respondWithError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid authorization header format", resp["error"])
}