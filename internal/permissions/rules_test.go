package permissions

import (
	"path/filepath"
	"testing"
)

func TestMatchRule_ToolOnly(t *testing.T) {
	r := Rule{Tool: "Bash"}
	if !matchRule(r, "Bash", nil) {
		t.Error("expected match for Bash")
	}
	if matchRule(r, "Read", nil) {
		t.Error("should not match Read")
	}
}

func TestMatchRule_PathGlob(t *testing.T) {
	r := Rule{Tool: "Write", Path: "*.go"}
	if !matchRule(r, "Write", map[string]any{"file_path": "main.go"}) {
		t.Error("expected *.go to match main.go")
	}
	if matchRule(r, "Write", map[string]any{"file_path": "main.py"}) {
		t.Error("should not match main.py")
	}
}

func TestMatchRule_CommandPrefix(t *testing.T) {
	r := Rule{Command: "git "}
	if !matchRule(r, "Bash", map[string]any{"command": "git status"}) {
		t.Error("expected 'git ' prefix to match 'git status'")
	}
	if matchRule(r, "Bash", map[string]any{"command": "rm -rf /"}) {
		t.Error("should not match rm -rf /")
	}
}

func TestMatchRule_Wildcard(t *testing.T) {
	r := Rule{} // empty rule matches everything
	if !matchRule(r, "Bash", map[string]any{"command": "anything"}) {
		t.Error("empty rule should match everything")
	}
}

func TestGlobMatch_DoubleStar(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		// Basic double star patterns
		{"src/**/*.go", "src/file.go", true},
		{"src/**/*.go", "src/sub/file.go", true},
		{"src/**/*.go", "src/sub/deep/file.go", true},
		{"src/**/*.go", "src/file.txt", false},
		{"src/**/*.go", "test/file.go", false},

		// Double star at beginning
		{"**/*.go", "file.go", true},
		{"**/*.go", "src/file.go", true},
		{"**/*.go", "src/sub/file.go", true},
		{"**/*.go", "src/file.txt", false},

		// Double star at end
		{"src/**", "src", true},
		{"src/**", "src/file.go", true},
		{"src/**", "src/sub/file.go", true},
		{"src/**", "test/file.go", false},

		// Double star in middle
		{"src/**/test/*.go", "src/test/file.go", true},
		{"src/**/test/*.go", "src/sub/test/file.go", true},
		{"src/**/test/*.go", "src/sub/deep/test/file.go", true},
		{"src/**/test/*.go", "src/test/file.txt", false},

		// Multiple segments with double star
		{"project/src/**/*.go", "project/src/file.go", true},
		{"project/src/**/*.go", "project/src/sub/file.go", true},
		{"project/src/**/*.go", "project/src/sub/deep/file.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"::"+tt.path, func(t *testing.T) {
			got := globMatch(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("globMatch(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestGlobMatch_WithSeparator(t *testing.T) {
	sep := string(filepath.Separator)

	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{sep + "tmp" + sep + "**", sep + "tmp" + sep + "file.txt", true},
		{sep + "tmp" + sep + "**", sep + "tmp" + sep + "sub" + sep + "file.txt", true},
		{"src" + sep + "**" + sep + "*.go", "src" + sep + "file.go", true},
		{"src" + sep + "**" + sep + "*.go", "src" + sep + "sub" + sep + "file.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"::"+tt.path, func(t *testing.T) {
			got := globMatch(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("globMatch(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}
