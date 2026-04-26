package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// NewTestResponseWriter creates a new test response writer
func NewTestResponseWriter() *TestResponseWriter {
	return &TestResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

// TestResponseWriter is a custom response writer for testing
type TestResponseWriter struct {
	*httptest.ResponseRecorder
}

// BuildTestRequest builds a test request with optional body and user ID
func BuildTestRequest(method, path string, body interface{}, userID string) *http.Request {
	var req *http.Request
	if body != nil {
		reqBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	if userID != "" {
		req = req.WithContext(setUserID(req.Context(), userID))
	}

	return req
}
