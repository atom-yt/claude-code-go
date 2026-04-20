package pathutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandHome(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"no tilde", "/tmp/file", "/tmp/file"},
		{"with tilde", "~/file", "/"},
		{"with tilde and path", "~/Documents/file", "/Documents/file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandHome(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestIsSensitive(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		isSen bool
	}{
		{"env file", ".env", true},
		{"local env", ".env.local", true},
		{"production env", ".env.production", true},
		{"credentials", "credentials", true},
		{"dot credentials", ".credentials", true},
		{"gitconfig", ".gitconfig", true},
		{"ssh dir", "/home/user/.ssh", true}, // .ssh dir is sensitive
		{"ssh file", "/home/user/.ssh/id_rsa", true},
		{"secrets dir", "secrets", true},
		{"regular file", "README.md", false},
		{"code file", "main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSensitive(tt.path)
			assert.Equal(t, tt.isSen, result, "path: %s", tt.path)
		})
	}
}

func TestIsWithinDir(t *testing.T) {
	tests := []struct {
		name string
		path string
		dir  string
		within bool
	}{
		{"same path", "/tmp", "/tmp", true},
		{"subdir", "/tmp/sub", "/tmp", true},
		{"deep sub", "/tmp/sub/deep/file.txt", "/tmp", true},
		{"outside", "/etc/passwd", "/tmp", false},
		{"parent", "/", "/tmp", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWithinDir(tt.path, tt.dir)
			assert.Equal(t, tt.within, result)
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	// Create a temporary directory as workspace
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		validator  *Validator
		input      string
		wantErr    bool
		errContains string
	}{
		{
			name:       "empty path",
			validator:  NewValidator(tmpDir),
			input:      "",
			wantErr:    true,
			errContains: "path is empty",
		},
		{
			name:      "valid path in workspace",
			validator: NewValidator(tmpDir),
			input:     filepath.Join(tmpDir, "file.txt"),
			wantErr:   false,
		},
		{
			name:       "path outside workspace",
			validator:  NewValidator(tmpDir),
			input:      "/etc/passwd",
			wantErr:    true,
			errContains: "outside workspace",
		},
		{
			name:      "path outside but allowed",
			validator: func() *Validator {
				v := NewValidator(tmpDir)
				v.SetAllowOutside(true)
				return v
			}(),
			input:    "/etc/passwd",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.validator.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				// Result should be absolute path
				assert.True(t, filepath.IsAbs(got) || filepath.IsAbs(tt.input))
			}
		})
	}
}

func TestValidator_ValidateRelativePath(t *testing.T) {
	// Test relative paths when CWD is within workspace
	tmpDir := t.TempDir()
	validator := NewValidator(tmpDir)

	// Temporarily change to workspace dir
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	t.Cleanup(func() { os.Chdir(oldDir) })

	got, err := validator.Validate("file.txt")
	assert.NoError(t, err)
	assert.True(t, filepath.IsAbs(got))
}

func TestValidator_TraversalAttack(t *testing.T) {
	tmpDir := t.TempDir()
	validator := NewValidator(tmpDir)

	// Try path traversal
	_, err := validator.Validate(filepath.Join(tmpDir, "..", "etc", "passwd"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside workspace")
}