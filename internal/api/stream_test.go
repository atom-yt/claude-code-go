package api

import (
	"context"
	"strings"
	"testing"
)

// collectEvents reads from ch until EventMessageStop or EventError is received.
func collectEvents(ch <-chan APIEvent) []APIEvent {
	var events []APIEvent
	for ev := range ch {
		events = append(events, ev)
		if ev.Type == EventMessageStop || ev.Type == EventError {
			return events
		}
	}
	return events
}

func TestParseSSE_TextDelta(t *testing.T) {
	raw := `event: message_start
data: {"type":"message_start","message":{}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":", world"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_stop
data: {"type":"message_stop"}

`
	ch := make(chan APIEvent, 16)
	go func() {
		parseSSE(context.Background(), strings.NewReader(raw), ch)
		close(ch)
	}()

	events := collectEvents(ch)

	if len(events) != 3 {
		t.Fatalf("want 3 events (2 text_delta + 1 message_stop), got %d: %+v", len(events), events)
	}
	if events[0].Type != EventTextDelta || events[0].Text != "Hello" {
		t.Errorf("event[0] want text_delta 'Hello', got %+v", events[0])
	}
	if events[1].Type != EventTextDelta || events[1].Text != ", world" {
		t.Errorf("event[1] want text_delta ', world', got %+v", events[1])
	}
	if events[2].Type != EventMessageStop {
		t.Errorf("event[2] want message_stop, got %+v", events[2])
	}
}

func TestParseSSE_Error(t *testing.T) {
	raw := `event: error
data: {"error":{"type":"overloaded_error","message":"server overloaded"}}

`
	ch := make(chan APIEvent, 4)
	go func() {
		parseSSE(context.Background(), strings.NewReader(raw), ch)
		close(ch)
	}()

	events := collectEvents(ch)

	if len(events) != 1 || events[0].Type != EventError {
		t.Fatalf("want 1 error event, got %+v", events)
	}
	if events[0].Error == nil {
		t.Fatal("error event should have non-nil Error")
	}
}
