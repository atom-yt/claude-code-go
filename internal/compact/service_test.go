package compact

import (
	"context"
	"testing"

	"github.com/atom-yt/claude-code-go/internal/api"
)

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

func TestCompactRewritesHistoryWithSummary(t *testing.T) {
	streamer := &mockStreamer{
		events: []api.APIEvent{
			{Type: api.EventTextDelta, Text: "- user goal\n- open issue"},
			{Type: api.EventMessageStop, StopReason: "end_turn"},
		},
	}

	history := []api.Message{
		api.TextMessage(api.RoleUser, "first"),
		api.TextMessage(api.RoleAssistant, "reply"),
		api.TextMessage(api.RoleUser, "recent"),
	}

	svc := NewService(streamer, "claude-test", "project rules")
	result, err := svc.Compact(context.Background(), history, 1)
	if err != nil {
		t.Fatalf("Compact: %v", err)
	}
	if result.Noop {
		t.Fatal("expected non-noop result")
	}
	if len(result.History) != 2 {
		t.Fatalf("want 2 history messages, got %d", len(result.History))
	}
	if result.History[0].Role != api.RoleAssistant {
		t.Fatalf("want summary role assistant, got %q", result.History[0].Role)
	}
	if got := result.History[0].Content[0].Text; got != "## Conversation Summary\n\n- user goal\n- open issue" {
		t.Fatalf("unexpected summary text: %q", got)
	}
	if result.History[1].Content[0].Text != "recent" {
		t.Fatalf("want recent tail preserved, got %q", result.History[1].Content[0].Text)
	}
	if streamer.lastReq.System == nil {
		t.Fatal("want system prompt in compact request")
	}
}

func TestCompactNoopWhenHistoryAlreadyShort(t *testing.T) {
	streamer := &mockStreamer{}
	history := []api.Message{api.TextMessage(api.RoleUser, "short")}

	svc := NewService(streamer, "claude-test", "")
	result, err := svc.Compact(context.Background(), history, 4)
	if err != nil {
		t.Fatalf("Compact: %v", err)
	}
	if !result.Noop {
		t.Fatal("expected noop result")
	}
	if len(result.History) != 1 {
		t.Fatalf("want 1 message, got %d", len(result.History))
	}
	if streamer.lastReq.Model != "" {
		t.Fatalf("expected no request to be made, got %#v", streamer.lastReq)
	}
}
