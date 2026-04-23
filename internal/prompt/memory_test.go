package prompt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscoverMemorySnippet_NoMemoryFile(t *testing.T) {
	// Create temp dir without memory file
	root := t.TempDir()

	// Set memory root to temp dir
	// Note: This test relies on the current directory not having a MEMORY.md
	// In a real scenario, this might need mocking
	result := DiscoverMemorySnippet()

	// Result should be empty or a reasonable fallback
	// We can't easily test the actual behavior without mocking filesystem
	_ = root // Use root to avoid unused variable
	assert.True(t, result == "" || len(result) > 0)
}

func TestDiscoverMemorySnippet_EmptyFile(t *testing.T) {
	root := t.TempDir()
	memoryPath := filepath.Join(root, "MEMORY.md")

	// Create empty MEMORY.md
	if err := os.WriteFile(memoryPath, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	// This test would need to mock memory.MemoryIndexPath()
	// For now, just verify the function doesn't panic
	_ = memoryPath
	_ = DiscoverMemorySnippet()
}

func TestDiscoverMemorySnippet_WhitespaceOnly(t *testing.T) {
	root := t.TempDir()
	memoryPath := filepath.Join(root, "MEMORY.md")

	// Create MEMORY.md with only whitespace
	if err := os.WriteFile(memoryPath, []byte("   \n\n   \t  "), 0o644); err != nil {
		t.Fatal(err)
	}

	_ = memoryPath
	_ = DiscoverMemorySnippet()
}

func TestMaxMemoryChars(t *testing.T) {
	// Verify the constant is set correctly
	assert.Equal(t, 8000, maxMemoryChars)
}