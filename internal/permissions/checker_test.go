package permissions

import (
	"context"
	"testing"
)

func TestChecker_TrustAll(t *testing.T) {
	c := New(ModeTrustAll)
	d, _ := c.Check(context.Background(), "Bash", map[string]any{"command": "rm -rf /"})
	if !d.Allowed {
		t.Error("trust-all should allow everything")
	}
}

func TestChecker_DenyRule(t *testing.T) {
	c := New(ModeDefault)
	c.DenyRules = []Rule{{Tool: "Bash"}}
	d, _ := c.Check(context.Background(), "Bash", map[string]any{"command": "echo hi"})
	if d.Allowed {
		t.Error("deny rule should block Bash")
	}
}

func TestChecker_AllowRule(t *testing.T) {
	c := New(ModeDefault)
	c.AllowRules = []Rule{{Tool: "Write", Path: "/tmp/*"}}
	d, _ := c.Check(context.Background(), "Write", map[string]any{"file_path": "/tmp/foo.txt"})
	if !d.Allowed {
		t.Errorf("allow rule should permit /tmp/foo.txt: %s", d.Reason)
	}
}

func TestChecker_DenyBeforeAllow(t *testing.T) {
	c := New(ModeDefault)
	c.DenyRules = []Rule{{Tool: "Write"}}
	c.AllowRules = []Rule{{Tool: "Write"}}
	d, _ := c.Check(context.Background(), "Write", map[string]any{})
	if d.Allowed {
		t.Error("deny should take priority over allow")
	}
}

func TestChecker_ReadOnlyAutoAllowed(t *testing.T) {
	c := New(ModeDefault) // no allow rules
	for _, tool := range []string{"Read", "Glob", "Grep"} {
		d, _ := c.Check(context.Background(), tool, map[string]any{})
		if !d.Allowed {
			t.Errorf("read-only tool %s should be auto-allowed in default mode", tool)
		}
	}
}

func TestChecker_AskRule(t *testing.T) {
	c := New(ModeDefault)
	c.AskRules = []Rule{{Tool: "Bash"}}
	asked := false
	c.AskFn = func(_ context.Context, _ AskRequest) (bool, string) {
		asked = true
		return true, ""
	}
	d, _ := c.Check(context.Background(), "Bash", map[string]any{"command": "ls"})
	if !asked {
		t.Error("AskFn should have been called")
	}
	if !d.Allowed {
		t.Errorf("AskFn returned true, expected allowed: %s", d.Reason)
	}
}

func TestChecker_NoAskFn_Denied(t *testing.T) {
	c := New(ModeManual) // requires ask, but AskFn is nil
	d, _ := c.Check(context.Background(), "Bash", map[string]any{})
	if d.Allowed {
		t.Error("without AskFn, manual mode should deny")
	}
}
