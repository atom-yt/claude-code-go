package api

import "testing"

func TestConvertSystemToMessagesString(t *testing.T) {
	msgs := convertSystemToMessages("follow project rules")
	if len(msgs) != 1 {
		t.Fatalf("want 1 message, got %d", len(msgs))
	}
	if msgs[0].Role != "system" {
		t.Fatalf("want role system, got %q", msgs[0].Role)
	}
	if msgs[0].Content != "follow project rules" {
		t.Fatalf("unexpected content: %#v", msgs[0].Content)
	}
}

func TestConvertSystemToMessagesBlocks(t *testing.T) {
	msgs := convertSystemToMessages([]ContentBlock{
		{Type: "text", Text: "part one"},
		{Type: "text", Text: "part two"},
	})
	if len(msgs) != 1 {
		t.Fatalf("want 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "part one\n\npart two" {
		t.Fatalf("unexpected content: %#v", msgs[0].Content)
	}
}
