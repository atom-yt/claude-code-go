package glob

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGlob_Match(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("pkg"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("txt"), 0o644)

	tool := &Tool{}
	result, _ := tool.Call(context.Background(), map[string]any{
		"pattern": "*.go",
		"path":    dir,
	})
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if !strings.Contains(result.Output, "a.go") {
		t.Errorf("expected a.go in output, got: %s", result.Output)
	}
	if strings.Contains(result.Output, "b.txt") {
		t.Errorf("b.txt should not match *.go, got: %s", result.Output)
	}
}

func TestGlob_NoMatch(t *testing.T) {
	dir := t.TempDir()
	tool := &Tool{}
	result, _ := tool.Call(context.Background(), map[string]any{
		"pattern": "*.xyz",
		"path":    dir,
	})
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if !strings.Contains(result.Output, "No files") {
		t.Errorf("expected no-match message, got: %s", result.Output)
	}
}
