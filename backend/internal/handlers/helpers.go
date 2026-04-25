package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Helper functions for handlers

// getURLParam extracts a URL parameter from the request
func getURLParam(r *http.Request, key string) string {
	vars := mux.Vars(r)
	return vars[key]
}

// getQueryInt extracts an integer query parameter with a default value
func getQueryInt(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// getPageParams extracts page and pageSize from query parameters
func getPageParams(r *http.Request) (page, pageSize int) {
	page = getQueryInt(r, "page", 1)
	pageSize = getQueryInt(r, "page_size", 10)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// userIDKey is the key used to store user ID in the request context
type contextKey string

const userIDKey contextKey = "userID"

// getUserID extracts the user ID from the request context
func getUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}

// setUserID sets the user ID in the request context
func setUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}