package agent_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/atom-yt/claude-code-go/internal/agent"
	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/tools"
)

// mockTool is a simple in-memory tool for testing.
type mockTool struct {
	name   string
	output string
}

func (t *mockTool) Name() string        { return t.name }
func (t *mockTool) Description() string { return "mock tool" }
func (t *mockTool) InputSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}
func (t *mockTool) IsReadOnly() bool        { return true }
func (t *mockTool) IsConcurrencySafe() bool { return true }
func (t *mockTool) Call(_ context.Context, _ map[string]any) (tools.ToolResult, error) {
	return tools.ToolResult{Output: t.output}, nil
}

type mockStreamer struct {
	lastReq api.MessagesRequest
	events  []api.APIEvent
}

func (m *mockStreamer) StreamMessages(_ context.Context, req api.MessagesRequest) <-chan api.APIEvent {
	m.lastReq = req
	ch := make(chan api.APIEvent, len(m.events))
	for _, ev := range m.events {
		ch <- ev
	}
	close(ch)
	return ch
}

// sseResponse builds a minimal SSE response body for the Anthropic streaming format.
func sseTextResponse(text string) string {
	return fmt.Sprintf(`event: message_start
data: {"type":"message_start","message":{"id":"msg_test","type":"message","role":"assistant","content":[],"model":"claude-test","stop_reason":null,"usage":{"input_tokens":10,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":%q}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":5}}

event: message_stop
data: {"type":"message_stop"}

`, text)
}

// sseToolThenTextResponse returns SSE for a tool_use block followed by a text response.
// It serves two responses: first with tool_use, then with final text.
func newMockServer(t *testing.T, responses []string) *httptest.Server {
	t.Helper()
	callCount := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		idx := callCount
		if idx >= len(responses) {
			idx = len(responses) - 1
		}
		callCount++
		fmt.Fprint(w, responses[idx])
	}))
}

func collectAgentEvents(ch <-chan agent.StreamEvent) []agent.StreamEvent {
	var events []agent.StreamEvent
	for ev := range ch {
		events = append(events, ev)
		if ev.Type == agent.EventDone || ev.Type == agent.EventError {
			break
		}
	}
	return events
}

// TestAgent_PlainTextResponse verifies that a plain text API response is
// delivered as EventTextDelta events followed by EventDone.
func TestAgent_PlainTextResponse(t *testing.T) {
	srv := newMockServer(t, []string{sseTextResponse("Hello from Claude")})
	defer srv.Close()

	client := api.New("test-key")
	client.BaseURL = srv.URL

	ag := agent.New(client, "claude-test", nil, nil, nil)
	ch := ag.Query(context.Background(), "Hi")
	events := collectAgentEvents(ch)

	var text strings.Builder
	var gotDone bool
	for _, ev := range events {
		switch ev.Type {
		case agent.EventTextDelta:
			text.WriteString(ev.Text)
		case agent.EventDone:
			gotDone = true
		case agent.EventError:
			t.Fatalf("unexpected error event: %v", ev.Error)
		}
	}

	if !gotDone {
		t.Error("expected EventDone, never received it")
	}
	if text.String() != "Hello from Claude" {
		t.Errorf("want text %q, got %q", "Hello from Claude", text.String())
	}
}

// TestAgent_ToolCall verifies the agent executes a tool and continues.
func TestAgent_ToolCall(t *testing.T) {
	// First response: tool_use block.
	toolUseResp := `event: message_start
data: {"type":"message_start","message":{"id":"msg1","type":"message","role":"assistant","content":[],"model":"claude-test","usage":{"input_tokens":10,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"tu_abc","name":"echo_tool","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":10}}

event: message_stop
data: {"type":"message_stop"}

`
	// Second response: final text after seeing tool result.
	finalResp := sseTextResponse("Tool result received.")

	srv := newMockServer(t, []string{toolUseResp, finalResp})
	defer srv.Close()

	client := api.New("test-key")
	client.BaseURL = srv.URL

	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "echo_tool", output: "echo output"})

	ag := agent.New(client, "claude-test", registry, nil, nil)
	ch := ag.Query(context.Background(), "Use the tool")
	events := collectAgentEvents(ch)

	var toolCalls, toolResults, textDeltas int
	var gotDone bool
	for _, ev := range events {
		switch ev.Type {
		case agent.EventToolCall:
			toolCalls++
			if ev.ToolName != "echo_tool" {
				t.Errorf("want tool name 'echo_tool', got %q", ev.ToolName)
			}
		case agent.EventToolResult:
			toolResults++
			if ev.ToolOutput != "echo output" {
				t.Errorf("want tool output 'echo output', got %q", ev.ToolOutput)
			}
		case agent.EventTextDelta:
			textDeltas++
		case agent.EventDone:
			gotDone = true
		case agent.EventError:
			t.Fatalf("unexpected error event: %v", ev.Error)
		}
	}

	if toolCalls != 1 {
		t.Errorf("want 1 EventToolCall, got %d", toolCalls)
	}
	if toolResults != 1 {
		t.Errorf("want 1 EventToolResult, got %d", toolResults)
	}
	if textDeltas == 0 {
		t.Error("want at least 1 EventTextDelta in final response")
	}
	if !gotDone {
		t.Error("expected EventDone, never received it")
	}
}

// TestAgent_APIError verifies that an HTTP error response produces EventError.
func TestAgent_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":{"type":"authentication_error","message":"invalid api key"}}`)
	}))
	defer srv.Close()

	client := api.New("bad-key")
	client.BaseURL = srv.URL

	ag := agent.New(client, "claude-test", nil, nil, nil)
	ch := ag.Query(context.Background(), "hello")
	events := collectAgentEvents(ch)

	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	last := events[len(events)-1]
	if last.Type != agent.EventError {
		t.Errorf("want EventError, got %v", last.Type)
	}
	if last.Error == nil {
		t.Error("EventError should have non-nil Error field")
	}
}

// TestAgent_UnknownTool verifies that calling an unregistered tool produces a
// tool_result error without crashing the agent loop.
func TestAgent_UnknownTool(t *testing.T) {
	toolUseResp := `event: message_start
data: {"type":"message_start","message":{"id":"msg2","type":"message","role":"assistant","content":[],"model":"claude-test","usage":{"input_tokens":10,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"tu_xyz","name":"nonexistent_tool","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":5}}

event: message_stop
data: {"type":"message_stop"}

`
	srv := newMockServer(t, []string{toolUseResp, sseTextResponse("I see the tool failed.")})
	defer srv.Close()

	client := api.New("test-key")
	client.BaseURL = srv.URL

	// Empty registry — tool won't be found.
	registry := tools.NewRegistry()

	ag := agent.New(client, "claude-test", registry, nil, nil)
	ch := ag.Query(context.Background(), "use unknown tool")
	events := collectAgentEvents(ch)

	var gotToolError bool
	for _, ev := range events {
		if ev.Type == agent.EventToolResult && ev.ToolIsError {
			gotToolError = true
		}
	}
	if !gotToolError {
		t.Error("expected an EventToolResult with ToolIsError=true for unknown tool")
	}
}

// TestAgent_HistoryPreserved verifies that SetHistory seeds the conversation
// and Query appends to it correctly.
func TestAgent_HistoryPreserved(t *testing.T) {
	srv := newMockServer(t, []string{sseTextResponse("continued")})
	defer srv.Close()

	client := api.New("test-key")
	client.BaseURL = srv.URL

	ag := agent.New(client, "claude-test", nil, nil, nil)

	prior := []api.Message{
		api.TextMessage(api.RoleUser, "prior user message"),
		{Role: api.RoleAssistant, Content: []api.ContentBlock{{Type: "text", Text: "prior assistant reply"}}},
	}
	ag.SetHistory(prior)

	ch := ag.Query(context.Background(), "follow-up")
	collectAgentEvents(ch)

	history := ag.History()
	// 2 prior + 1 user (follow-up) + 1 assistant reply = 4
	if len(history) != 4 {
		t.Fatalf("want 4 history entries (2 prior + user + assistant), got %d", len(history))
	}
	if history[2].Role != api.RoleUser {
		t.Errorf("want 3rd entry to be user, got %q", history[2].Role)
	}
	if history[3].Role != api.RoleAssistant {
		t.Errorf("want 4th entry to be assistant, got %q", history[3].Role)
	}
}

// TestAgent_SystemPromptInjected verifies the agent forwards its system prompt
// through the request payload.
func TestAgent_SystemPromptInjected(t *testing.T) {
	streamer := &mockStreamer{
		events: []api.APIEvent{
			{Type: api.EventTextDelta, Text: "done"},
			{Type: api.EventMessageStop, StopReason: "end_turn"},
		},
	}

	ag := agent.New(streamer, "claude-test", nil, nil, nil)
	ag.SetSystemPrompt("use project instructions")

	ch := ag.Query(context.Background(), "hello")
	collectAgentEvents(ch)

	blocks, ok := streamer.lastReq.System.([]api.ContentBlock)
	if !ok || len(blocks) != 1 {
		t.Fatalf("want one system block, got %#v", streamer.lastReq.System)
	}
	if blocks[0].Text != "use project instructions" {
		t.Fatalf("want injected system prompt, got %#v", blocks[0].Text)
	}
}
