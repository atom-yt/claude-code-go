package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/hooks"
	"github.com/atom-yt/claude-code-go/internal/permissions"
	"github.com/atom-yt/claude-code-go/internal/tools"
)

// Agent manages conversation history and runs the multi-turn tool loop.
type Agent struct {
	client   api.Streamer
	model    string
	registry *tools.Registry
	checker  *permissions.Checker
	executor *hooks.Executor
	history  []api.Message
}

// New creates an Agent. checker and executor may be nil.
func New(client api.Streamer, model string, registry *tools.Registry, checker *permissions.Checker, executor *hooks.Executor) *Agent {
	return &Agent{
		client:   client,
		model:    model,
		registry: registry,
		checker:  checker,
		executor: executor,
	}
}

// SetModel changes the model used for subsequent queries.
func (a *Agent) SetModel(model string) { a.model = model }

// SetHistory replaces the conversation history (used when resuming a session).
func (a *Agent) SetHistory(msgs []api.Message) {
	a.history = make([]api.Message, len(msgs))
	copy(a.history, msgs)
}

// History returns a copy of the current conversation history.
func (a *Agent) History() []api.Message {
	out := make([]api.Message, len(a.history))
	copy(out, a.history)
	return out
}

// Query appends the user message to history, runs the agent loop, and streams
// events to the returned channel.
func (a *Agent) Query(ctx context.Context, userText string) <-chan StreamEvent {
	ch := make(chan StreamEvent, 64)
	go a.run(ctx, userText, ch)
	return ch
}

func (a *Agent) run(ctx context.Context, userText string, ch chan<- StreamEvent) {
	defer func() {
		if a.executor != nil {
			a.executor.FireStop(ctx)
		}
	}()
	defer close(ch)

	a.history = append(a.history, api.TextMessage(api.RoleUser, userText))

	// Build tool specs for the API.
	var toolSpecs []api.ToolSpec
	if a.registry != nil {
		for _, spec := range a.registry.ToAPISpecs() {
			toolSpecs = append(toolSpecs, api.ToolSpec{
				Name:        spec.Name,
				Description: spec.Description,
				InputSchema: spec.InputSchema,
			})
		}
	}

	for {
		req := api.MessagesRequest{
			Model:    a.model,
			Messages: a.history,
			Tools:    toolSpecs,
		}

		apiCh := a.client.StreamMessages(ctx, req)

		var (
			fullText   string
			toolUses   []api.ToolUse
			stopReason string
		)

		for ev := range apiCh {
			switch ev.Type {
			case api.EventTextDelta:
				fullText += ev.Text
				ch <- StreamEvent{Type: EventTextDelta, Text: ev.Text}

			case api.EventToolUse:
				if ev.ToolUse != nil {
					toolUses = append(toolUses, *ev.ToolUse)
				}

			case api.EventMessageStop:
				stopReason = ev.StopReason

			case api.EventError:
				a.history = a.history[:len(a.history)-1]
				ch <- StreamEvent{Type: EventError, Error: ev.Error}
				return
			}
		}

		_ = stopReason

		// Build assistant message with text + tool_use blocks.
		assistantBlocks := []api.ContentBlock{}
		if fullText != "" {
			assistantBlocks = append(assistantBlocks, api.ContentBlock{Type: "text", Text: fullText})
		}
		for _, tu := range toolUses {
			assistantBlocks = append(assistantBlocks, api.ContentBlock{
				Type:  "tool_use",
				ID:    tu.ID,
				Name:  tu.Name,
				Input: tu.Input,
			})
		}
		a.history = append(a.history, api.Message{
			Role:    api.RoleAssistant,
			Content: assistantBlocks,
		})

		if len(toolUses) == 0 {
			ch <- StreamEvent{Type: EventDone}
			return
		}

		// Phase 1 (sequential): pre-hooks + permission checks.
		// Produces an ordered list of approved tool calls ready for execution.
		type approved struct {
			tu             api.ToolUse
			concurrentSafe bool
		}
		var approvedTools []approved
		var toolResults []api.ToolResult

		aborted := false
		for _, tu := range toolUses {
			// Pre-tool hooks.
			if a.executor != nil {
				deny, reason, err := a.executor.FirePreToolCall(ctx, tu.Name, tu.Input)
				if err == nil && deny {
					denyMsg := fmt.Sprintf("Hook denied tool call: %s", reason)
					toolResults = append(toolResults, api.ToolResult{
						ToolUseID: tu.ID,
						Output:    denyMsg,
						IsError:   true,
					})
					ch <- StreamEvent{Type: EventToolResult, ToolName: tu.Name, ToolOutput: denyMsg, ToolIsError: true}
					continue
				}
			}

			// Permission check (may block waiting for user input via AskFn).
			if a.checker != nil {
				decision, err := a.checker.Check(ctx, tu.Name, tu.Input)
				if err != nil {
					ch <- StreamEvent{Type: EventError, Error: fmt.Errorf("permission check error: %w", err)}
					aborted = true
					break
				}
				if !decision.Allowed {
					denyMsg := fmt.Sprintf("Permission denied: %s", decision.Reason)
					toolResults = append(toolResults, api.ToolResult{
						ToolUseID: tu.ID,
						Output:    denyMsg,
						IsError:   true,
					})
					ch <- StreamEvent{Type: EventToolResult, ToolName: tu.Name, ToolOutput: denyMsg, ToolIsError: true}
					continue
				}
			}

			ch <- StreamEvent{Type: EventToolCall, ToolName: tu.Name, ToolInput: tu.Input}

			safe := false
			if a.registry != nil {
				if t, ok := a.registry.GetByName(tu.Name); ok {
					safe = t.IsConcurrencySafe()
				}
			}
			approvedTools = append(approvedTools, approved{tu: tu, concurrentSafe: safe})
		}
		if aborted {
			return
		}

		// Phase 2: execute approved tools.
		// Concurrent-safe tools run in parallel; others run sequentially.
		// Results are collected in original order.
		results := make([]api.ToolResult, len(approvedTools))
		allSafe := true
		for _, a := range approvedTools {
			if !a.concurrentSafe {
				allSafe = false
				break
			}
		}

		if allSafe && len(approvedTools) > 1 {
			var wg sync.WaitGroup
			for i, ap := range approvedTools {
				i, ap := i, ap
				wg.Add(1)
				go func() {
					defer wg.Done()
					results[i] = a.executeTool(ctx, ap.tu)
				}()
			}
			wg.Wait()
		} else {
			for i, ap := range approvedTools {
				results[i] = a.executeTool(ctx, ap.tu)
			}
		}

		for i, ap := range approvedTools {
			result := results[i]
			toolResults = append(toolResults, result)
			ch <- StreamEvent{
				Type:        EventToolResult,
				ToolName:    ap.tu.Name,
				ToolOutput:  result.Output,
				ToolIsError: result.IsError,
			}
			// Post-tool hooks (async, non-blocking).
			if a.executor != nil {
				go a.executor.FirePostToolCall(ctx, ap.tu.Name, ap.tu.Input)
			}
		}

		// Append tool results and continue the loop.
		a.history = append(a.history, api.ToolResultMessage(toolResults))
	}
}

// executeTool finds and calls a tool by name from the registry.
func (a *Agent) executeTool(ctx context.Context, tu api.ToolUse) api.ToolResult {
	if a.registry == nil {
		return api.ToolResult{
			ToolUseID: tu.ID,
			Output:    fmt.Sprintf("tool %q not found: no registry", tu.Name),
			IsError:   true,
		}
	}

	t, ok := a.registry.GetByName(tu.Name)
	if !ok {
		return api.ToolResult{
			ToolUseID: tu.ID,
			Output:    fmt.Sprintf("unknown tool: %q", tu.Name),
			IsError:   true,
		}
	}

	input := tu.Input
	if input == nil {
		input = map[string]any{}
	}

	toolResult, err := t.Call(ctx, input)
	if err != nil {
		return api.ToolResult{
			ToolUseID: tu.ID,
			Output:    fmt.Sprintf("tool execution error: %v", err),
			IsError:   true,
		}
	}

	output := toolResult.Output
	if output == "" && toolResult.IsError {
		output = "tool returned an error with no message"
	}

	// Truncate very large outputs to avoid exceeding token limits.
	const maxOutputChars = 50_000
	if len(output) > maxOutputChars {
		summary, _ := json.Marshal(map[string]any{
			"truncated": true,
			"chars":     len(output),
			"preview":   output[:500],
		})
		output = string(summary)
	}

	return api.ToolResult{
		ToolUseID: tu.ID,
		Output:    output,
		IsError:   toolResult.IsError,
	}
}
