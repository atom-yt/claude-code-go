package permissions

import (
	"context"
	"fmt"
)

// AskFn is called when the checker needs interactive approval from the user.
// Returns (true, "") to allow, (false, reason) to deny.
type AskFn func(ctx context.Context, req AskRequest) (bool, string)

// Checker evaluates permission rules for tool calls.
type Checker struct {
	Mode       Mode
	AllowRules []Rule
	DenyRules  []Rule
	AskRules   []Rule
	// AskFn is required when Mode == ModeManual or AskRules are configured.
	// If nil, unknown calls are denied.
	AskFn AskFn
}

// New creates a Checker with the given mode and empty rule lists.
func New(mode Mode) *Checker {
	return &Checker{Mode: mode}
}

// Check evaluates whether a tool call is permitted.
// Priority: Deny > Allow > Ask > Mode default.
func (c *Checker) Check(ctx context.Context, toolName string, input map[string]any) (Decision, error) {
	// trust-all: skip all checks
	if c.Mode == ModeTrustAll {
		return Decision{Allowed: true}, nil
	}

	// 1. Deny rules (highest priority)
	for _, rule := range c.DenyRules {
		if matchRule(rule, toolName, input) {
			return Decision{
				Allowed: false,
				Reason:  fmt.Sprintf("denied by rule: tool=%q path=%q command=%q", rule.Tool, rule.Path, rule.Command),
			}, nil
		}
	}

	// 2. Allow rules
	for _, rule := range c.AllowRules {
		if matchRule(rule, toolName, input) {
			return Decision{Allowed: true}, nil
		}
	}

	// 3. Ask rules
	for _, rule := range c.AskRules {
		if matchRule(rule, toolName, input) {
			return c.ask(ctx, toolName, input, ruleDesc(rule))
		}
	}

	// 4. Mode default
	switch c.Mode {
	case ModeTrustAll:
		return Decision{Allowed: true}, nil
	case ModeManual:
		return c.ask(ctx, toolName, input, "manual mode")
	default: // ModeDefault
		// In default mode, read-only tools are auto-allowed;
		// mutating tools require ask.
		if isReadOnlyTool(toolName) {
			return Decision{Allowed: true}, nil
		}
		return c.ask(ctx, toolName, input, "default mode")
	}
}

func (c *Checker) ask(ctx context.Context, toolName string, input map[string]any, reason string) (Decision, error) {
	if c.AskFn == nil {
		return Decision{
			Allowed: false,
			Reason:  fmt.Sprintf("permission required for %q (%s) but no interactive prompt available", toolName, reason),
		}, nil
	}
	allowed, denyReason := c.AskFn(ctx, AskRequest{
		ToolName:  toolName,
		Input:     input,
		RuleMatch: reason,
	})
	if !allowed {
		if denyReason == "" {
			denyReason = "denied by user"
		}
		return Decision{Allowed: false, Reason: denyReason}, nil
	}
	return Decision{Allowed: true}, nil
}

// isReadOnlyTool returns true for tools that are known to be read-only
// and can be auto-approved in default mode.
func isReadOnlyTool(name string) bool {
	switch name {
	case "Read", "Glob", "Grep":
		return true
	}
	return false
}

func ruleDesc(r Rule) string {
	return fmt.Sprintf("ask rule: tool=%q path=%q command=%q", r.Tool, r.Path, r.Command)
}
