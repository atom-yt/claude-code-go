package commands

import (
	"context"
	"fmt"
	"strings"
)

// ---- /hooks ----

type hooksCmd struct{}

func (c *hooksCmd) Name() string      { return "hooks" }
func (c *hooksCmd) Aliases() []string { return nil }
func (c *hooksCmd) Description() string {
	return "Show current hook settings"
}

func (c *hooksCmd) Execute(_ context.Context, _ []string, cmdCtx *Context) (string, error) {
	if cmdCtx.GetConfig == nil {
		return "Config not available", nil
	}

	cfg := cmdCtx.GetConfig()
	hooks, ok := cfg["hooks"].(map[string]any)
	if !ok || len(hooks) == 0 {
		return "No hooks configured.\n\nConfigure hooks in ~/.claude/settings.json or .claude/settings.json", nil
	}

	lines := []string{"Hook settings:"}

	// Event descriptions
	eventDesc := map[string]string{
		"pre_tool_call":      "Runs before tool execution, can block",
		"post_tool_call":     "Runs after tool execution",
		"session_start":      "Runs when session starts",
		"stop":               "Runs when session stops",
		"user_prompt_submit": "Runs when user submits prompt",
	}

	// Order events consistently
	eventOrder := []string{"pre_tool_call", "post_tool_call", "session_start", "stop", "user_prompt_submit"}

	for _, event := range eventOrder {
		matchers, ok := hooks[event].([]any)
		if !ok || len(matchers) == 0 {
			continue
		}

		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("%s (%d matchers):", event, len(matchers)))
		if desc, ok := eventDesc[event]; ok {
			lines = append(lines, fmt.Sprintf("  %s", desc))
		}

		for i, m := range matchers {
			if matcher, ok := m.(map[string]any); ok {
				lines = append(lines, c.formatMatcher(i+1, matcher))
			}
		}
	}

	// Show other events not in the standard order
	for event, matchers := range hooks {
		found := false
		for _, eo := range eventOrder {
			if eo == event {
				found = true
				break
			}
		}
		if !found {
			if m, ok := matchers.([]any); ok && len(m) > 0 {
				lines = append(lines, "")
				lines = append(lines, fmt.Sprintf("%s (%d matchers):", event, len(m)))
				for i, matcher := range m {
					if matcherMap, ok := matcher.(map[string]any); ok {
						lines = append(lines, c.formatMatcher(i+1, matcherMap))
					}
				}
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, "Hook types:")
	lines = append(lines, "  shell - Execute shell command")
	lines = append(lines, "  http  - Send HTTP request (url, headers)")

	return "<!-- raw -->\n" + strings.Join(lines, "\n"), nil
}

func (c *hooksCmd) formatMatcher(index int, matcher map[string]any) string {
	lines := []string{fmt.Sprintf("  %d.", index)}

	if pattern, ok := matcher["matcher"].(string); ok && pattern != "" {
		lines[0] += fmt.Sprintf(` matcher: "%s"`, pattern)
	}

	if hookList, ok := matcher["hooks"].([]any); ok {
		for _, h := range hookList {
			if hook, ok := h.(map[string]any); ok {
				lines = append(lines, c.formatHook(hook))
			}
		}
	}

	return strings.Join(lines, "\n")
}

func (c *hooksCmd) formatHook(hook map[string]any) string {
	hookType, _ := hook["type"].(string)
	if hookType == "" {
		hookType = "unknown"
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("    - type: %s", hookType))

	if cmd, ok := hook["command"].(string); ok && cmd != "" {
		parts = append(parts, fmt.Sprintf("command: %s", cmd))
	}
	if url, ok := hook["url"].(string); ok && url != "" {
		parts = append(parts, fmt.Sprintf("url: %s", url))
	}
	if timeout, ok := hook["timeout"].(int); ok && timeout > 0 {
		parts = append(parts, fmt.Sprintf("timeout: %dms", timeout))
	}
	if headers, ok := hook["headers"].(map[string]any); ok && len(headers) > 0 {
		parts = append(parts, fmt.Sprintf("headers: %d defined", len(headers)))
	}

	return strings.Join(parts, ", ")
}