package bash

import (
	"context"
	"strings"
	"testing"
)

func TestBash_Success(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{"command": "echo hello"})
	if err != nil || result.IsError {
		t.Fatalf("unexpected error: %v / %s", err, result.Output)
	}
	if !strings.Contains(result.Output, "hello") {
		t.Errorf("expected 'hello' in output, got: %s", result.Output)
	}
}

func TestBash_NonZeroExit(t *testing.T) {
	tool := &Tool{}
	result, _ := tool.Call(context.Background(), map[string]any{"command": "exit 42"})
	if !result.IsError {
		t.Fatal("expected error for non-zero exit")
	}
	if !strings.Contains(result.Output, "42") {
		t.Errorf("expected exit code in output, got: %s", result.Output)
	}
}

func TestBash_Timeout(t *testing.T) {
	tool := &Tool{}
	result, _ := tool.Call(context.Background(), map[string]any{
		"command": "sleep 10",
		"timeout": float64(1),
	})
	if !result.IsError {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(result.Output, "timed out") {
		t.Errorf("expected timeout message, got: %s", result.Output)
	}
}
