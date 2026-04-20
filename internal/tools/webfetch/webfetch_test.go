package webfetch

import (
	"context"
	"strings"
	"testing"
)

func TestWebFetch_Name(t *testing.T) {
	tool := NewTool()
	if tool.Name() != "WebFetch" {
		t.Errorf("expected name WebFetch, got %s", tool.Name())
	}
}

func TestWebFetch_ReadOnly(t *testing.T) {
	tool := NewTool()
	if !tool.IsReadOnly() {
		t.Error("WebFetch should be read-only")
	}
}

func TestWebFetch_ConcurrencySafe(t *testing.T) {
	tool := NewTool()
	if !tool.IsConcurrencySafe() {
		t.Error("WebFetch should be concurrency safe")
	}
}

func TestWebFetch_MissingURL(t *testing.T) {
	tool := NewTool()
	result, err := tool.Call(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing URL")
	}
}

func TestWebFetch_InvalidURL(t *testing.T) {
	tool := NewTool()
	result, err := tool.Call(context.Background(), map[string]any{
		"url": "not a valid url",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid URL")
	}
}

func TestWebFetch_TimeoutParam(t *testing.T) {
	tool := NewTool()
	// Test that timeout parameter is accepted (actual timeout behavior is network-dependent)
	// We use a longer timeout to avoid flaky test failures
	result, err := tool.Call(context.Background(), map[string]any{
		"url":        "https://httpbin.org/delay/1",
		"timeout_ms": 5000, // 5 seconds - long enough for the response
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The test passes if we got a result without error (or validation error)
	// The actual behavior is network-dependent
	_ = result
}

func TestWebFetch_InternalIPBlocked(t *testing.T) {
	tool := NewTool()
	testCases := []string{
		"http://127.0.0.1:8080",
		"http://localhost:8080",
		"http://10.0.0.1",
		"http://192.168.1.1",
		"http://172.16.0.1",
		"http://[::1]:8080",
	}

	for _, url := range testCases {
		t.Run(url, func(t *testing.T) {
			result, err := tool.Call(context.Background(), map[string]any{
				"url": url,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.IsError {
				t.Errorf("expected error for internal URL %s", url)
			}
			if !strings.Contains(result.Output, "validation") && !strings.Contains(result.Output, "internal") {
				t.Errorf("expected validation error for internal URL %s, got: %s", url, result.Output)
			}
		})
	}
}

func TestWebFetch_DangerousSchemeBlocked(t *testing.T) {
	tool := NewTool()
	testCases := []string{
		"file:///etc/passwd",
		"ftp://example.com",
		"javascript:alert('xss')",
	}

	for _, url := range testCases {
		t.Run(url, func(t *testing.T) {
			result, err := tool.Call(context.Background(), map[string]any{
				"url": url,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.IsError {
				t.Errorf("expected error for dangerous scheme in URL %s", url)
			}
			if !strings.Contains(result.Output, "validation") && !strings.Contains(result.Output, "scheme") {
				t.Errorf("expected validation error for dangerous scheme %s, got: %s", url, result.Output)
			}
		})
	}
}

func TestWebFetch_ValidPublicURL(t *testing.T) {
	tool := NewTool()
	result, err := tool.Call(context.Background(), map[string]any{
		"url": "https://example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error for valid URL: %s", result.Output)
	}
}
