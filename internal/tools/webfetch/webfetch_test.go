package webfetch

import (
	"context"
	"testing"
)

func TestWebFetch_Name(t *testing.T) {
	tool := &Tool{}
	if tool.Name() != "WebFetch" {
		t.Errorf("expected name WebFetch, got %s", tool.Name())
	}
}

func TestWebFetch_ReadOnly(t *testing.T) {
	tool := &Tool{}
	if !tool.IsReadOnly() {
		t.Error("WebFetch should be read-only")
	}
}

func TestWebFetch_ConcurrencySafe(t *testing.T) {
	tool := &Tool{}
	if !tool.IsConcurrencySafe() {
		t.Error("WebFetch should be concurrency safe")
	}
}

func TestWebFetch_MissingURL(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing URL")
	}
}

func TestWebFetch_InvalidURL(t *testing.T) {
	tool := &Tool{}
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
	tool := &Tool{}
	// Use a very short timeout to trigger timeout error
	result, err := tool.Call(context.Background(), map[string]any{
		"url":        "https://httpbin.org/delay/5",
		"timeout_ms": 10, // 10ms timeout
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected timeout error")
	}
}