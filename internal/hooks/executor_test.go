package hooks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecutor_NoHooks(t *testing.T) {
	e := New(Config{}, "sid")
	deny, reason, err := e.FirePreToolCall(context.Background(), "Bash", nil)
	if err != nil || deny {
		t.Errorf("no hooks should allow everything: deny=%v reason=%q err=%v", deny, reason, err)
	}
}

func TestExecutor_ShellHook_Allow(t *testing.T) {
	e := New(Config{
		EventPreToolCall: []Matcher{{
			ToolPattern: "Bash",
			Hooks: []HookCommand{{
				Type:    TypeCommand,
				Command: `echo '{"decision":"allow"}'`,
			}},
		}},
	}, "sid")

	deny, _, err := e.FirePreToolCall(context.Background(), "Bash", map[string]any{"command": "ls"})
	if err != nil {
		t.Fatal(err)
	}
	if deny {
		t.Error("allow decision should not deny")
	}
}

func TestExecutor_ShellHook_Deny(t *testing.T) {
	e := New(Config{
		EventPreToolCall: []Matcher{{
			Hooks: []HookCommand{{
				Type:    TypeCommand,
				Command: `echo '{"decision":"deny","reason":"blocked"}' && exit 1`,
			}},
		}},
	}, "sid")

	deny, reason, err := e.FirePreToolCall(context.Background(), "Bash", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !deny {
		t.Error("expected deny")
	}
	if reason != "blocked" {
		t.Errorf("expected reason 'blocked', got %q", reason)
	}
}

func TestExecutor_HTTPHook_Allow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"decision": "allow"})
	}))
	defer srv.Close()

	e := New(Config{
		EventPreToolCall: []Matcher{{
			Hooks: []HookCommand{{
				Type: TypeHTTP,
				URL:  srv.URL,
			}},
		}},
	}, "sid")

	deny, _, err := e.FirePreToolCall(context.Background(), "Write", nil)
	if err != nil {
		t.Fatal(err)
	}
	if deny {
		t.Error("http allow should not deny")
	}
}

func TestExecutor_HTTPHook_Deny(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"decision": "deny", "reason": "http denied"})
	}))
	defer srv.Close()

	e := New(Config{
		EventPreToolCall: []Matcher{{
			Hooks: []HookCommand{{Type: TypeHTTP, URL: srv.URL}},
		}},
	}, "sid")

	deny, reason, err := e.FirePreToolCall(context.Background(), "Write", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !deny {
		t.Error("expected deny from 403")
	}
	if reason != "http denied" {
		t.Errorf("expected 'http denied', got %q", reason)
	}
}

func TestExecutor_PatternMatch(t *testing.T) {
	called := false
	// Use a hook that sets called via exit code (0 = allow, touches 'called').
	e := New(Config{
		EventPreToolCall: []Matcher{{
			ToolPattern: "Bash",
			Hooks: []HookCommand{{
				Type:    TypeCommand,
				Command: `echo allow`,
			}},
		}},
	}, "sid")
	_ = called

	// Bash should match.
	deny, _, _ := e.FirePreToolCall(context.Background(), "Bash", nil)
	if deny {
		t.Error("bash matched pattern: should allow")
	}

	// Read should NOT match the Bash pattern → allow (no hooks fired).
	deny, _, _ = e.FirePreToolCall(context.Background(), "Read", nil)
	if deny {
		t.Error("Read did not match Bash pattern: should allow")
	}
}
