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

func TestChecker_MCPFullTrust_AutoAllowed(t *testing.T) {
	c := New(ModeDefault)
	c.MCPTrustLevels = map[string]string{"myserver": "full"}

	d, _ := c.Check(context.Background(), "mcp__myserver__tool1", map[string]any{})
	if !d.Allowed {
		t.Errorf("MCP tool from 'full' trust server should be auto-allowed, got allowed=%v, reason=%q", d.Allowed, d.Reason)
	}
}

func TestChecker_MCPLimitedTrust_RequiresAsk(t *testing.T) {
	c := New(ModeDefault)
	c.MCPTrustLevels = map[string]string{"myserver": "limited"}

	asked := false
	c.AskFn = func(_ context.Context, _ AskRequest) (bool, string) {
		asked = true
		return true, ""
	}
	d, _ := c.Check(context.Background(), "mcp__myserver__tool1", map[string]any{})

	if !asked {
		t.Error("MCP tool from 'limited' trust server should ask for permission")
	}
	if !d.Allowed {
		t.Errorf("MCP tool from 'limited' trust should be allowed after ask: %s", d.Reason)
	}
}

func TestChecker_MCPUntrustedTrust_RequiresAsk(t *testing.T) {
	c := New(ModeDefault)
	c.MCPTrustLevels = map[string]string{"myserver": "untrusted"}

	asked := false
	c.AskFn = func(_ context.Context, _ AskRequest) (bool, string) {
		asked = true
		return true, ""
	}
	d, _ := c.Check(context.Background(), "mcp__myserver__tool1", map[string]any{})

	if !asked {
		t.Error("MCP tool from 'untrusted' trust server should ask for permission")
	}
	if !d.Allowed {
		t.Errorf("MCP tool from 'untrusted' trust should be allowed after ask: %s", d.Reason)
	}
}

func TestChecker_MCPDefaultTrust_RequiresAsk(t *testing.T) {
	c := New(ModeDefault)
	c.MCPTrustLevels = map[string]string{"myserver": ""} // empty = default untrusted

	asked := false
	c.AskFn = func(_ context.Context, _ AskRequest) (bool, string) {
		asked = true
		return true, ""
	}
	d, _ := c.Check(context.Background(), "mcp__myserver__tool1", map[string]any{})

	if !asked {
		t.Error("MCP tool with default trust should ask for permission")
	}
	if !d.Allowed {
		t.Errorf("MCP tool with default trust should be allowed after ask: %s", d.Reason)
	}
}

func TestChecker_MCPDenyRuleTakesPriority(t *testing.T) {
	c := New(ModeDefault)
	c.MCPTrustLevels = map[string]string{"myserver": "full"}
	c.DenyRules = []Rule{{Tool: "mcp__myserver__tool1"}}

	d, _ := c.Check(context.Background(), "mcp__myserver__tool1", map[string]any{})
	if d.Allowed {
		t.Error("deny rule should block MCP tool even with full trust")
	}
}

