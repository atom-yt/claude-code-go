package hooks

import (
	"context"
	"path/filepath"
)

// Config holds all hook matchers keyed by event name.
type Config map[Event][]Matcher

// Executor fires hooks for lifecycle events.
type Executor struct {
	config    Config
	sessionID string
}

// New creates an Executor with the given hook config.
func New(cfg Config, sessionID string) *Executor {
	return &Executor{config: cfg, sessionID: sessionID}
}

// FirePreToolCall runs pre_tool_call hooks and returns whether the call
// should be denied. Returns ("", nil) when everything is allowed.
func (e *Executor) FirePreToolCall(ctx context.Context, toolName string, input map[string]any) (deny bool, reason string, err error) {
	result, err := e.fire(ctx, EventPreToolCall, toolName, input)
	if err != nil {
		return false, "", err
	}
	if result.Decision == "deny" {
		return true, result.Reason, nil
	}
	return false, "", nil
}

// FirePostToolCall runs post_tool_call hooks (informational; result is ignored).
func (e *Executor) FirePostToolCall(ctx context.Context, toolName string, input map[string]any) {
	_, _ = e.fire(ctx, EventPostToolCall, toolName, input)
}

// FireSessionStart runs session_start hooks.
func (e *Executor) FireSessionStart(ctx context.Context) {
	_, _ = e.fire(ctx, EventSessionStart, "", nil)
}

// FireStop runs stop hooks.
func (e *Executor) FireStop(ctx context.Context) {
	_, _ = e.fire(ctx, EventStop, "", nil)
}

// FireUserPromptSubmit runs user_prompt_submit hooks when the user submits a prompt.
// This allows logging, validation, or modification of user input before processing.
func (e *Executor) FireUserPromptSubmit(ctx context.Context, userPrompt string) {
	hookInput := Input{
		Event:      EventUserPromptSubmit,
		SessionID:  e.sessionID,
		UserPrompt: userPrompt,
	}
	_, _ = e.fireWithInput(ctx, EventUserPromptSubmit, hookInput)
}

// fireWithInput runs all matching hooks with the provided input.
func (e *Executor) fireWithInput(ctx context.Context, event Event, hookInput Input) (Result, error) {
	matchers, ok := e.config[event]
	if !ok {
		return Result{}, nil
	}

	var last Result
	for _, m := range matchers {
		if !matchTool(m.ToolPattern, "") {
			continue
		}
		for _, cmd := range m.Hooks {
			result, err := run(ctx, cmd, hookInput)
			if err != nil {
				// Hook execution error: log and continue (don't block).
				continue
			}
			last = result
			if result.Decision == "deny" {
				return result, nil
			}
		}
	}
	return last, nil
}

// fire runs all matching hooks for the given event and returns the first
// deny decision, or the last result if none denies.
func (e *Executor) fire(ctx context.Context, event Event, toolName string, toolInput map[string]any) (Result, error) {
	matchers, ok := e.config[event]
	if !ok {
		return Result{}, nil
	}

	hookInput := Input{
		Event:     event,
		ToolName:  toolName,
		ToolInput: toolInput,
		SessionID: e.sessionID,
	}

	var last Result
	for _, m := range matchers {
		if !matchTool(m.ToolPattern, toolName) {
			continue
		}
		for _, cmd := range m.Hooks {
			result, err := run(ctx, cmd, hookInput)
			if err != nil {
				// Hook execution error: log and continue (don't block).
				continue
			}
			last = result
			if result.Decision == "deny" {
				return result, nil
			}
		}
	}
	return last, nil
}

// matchTool returns true if the tool name matches the pattern.
// Empty pattern matches everything.
func matchTool(pattern, toolName string) bool {
	if pattern == "" {
		return true
	}
	matched, err := filepath.Match(pattern, toolName)
	return err == nil && matched
}
