package permissions

import "testing"

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
