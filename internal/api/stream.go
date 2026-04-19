package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// pendingToolUse tracks a tool_use block being assembled from deltas.
type pendingToolUse struct {
	id        string
	name      string
	inputJSON strings.Builder
}

// parseSSE reads an SSE stream and sends parsed APIEvents to ch.
func parseSSE(ctx context.Context, r io.Reader, ch chan<- APIEvent) {
	scanner := bufio.NewScanner(r)
	// Increase scanner buffer for large JSON payloads.
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var eventType string
	var dataLine string

	// Tracks tool_use blocks in progress, keyed by content block index.
	pending := map[int]*pendingToolUse{}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			ch <- APIEvent{Type: EventError, Error: ctx.Err()}
			return
		default:
		}

		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "event:"):
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))

		case strings.HasPrefix(line, "data:"):
			dataLine = strings.TrimSpace(strings.TrimPrefix(line, "data:"))

		case line == "":
			if eventType == "" && dataLine == "" {
				continue
			}
			if done := dispatchEvent(eventType, dataLine, pending, ch); done {
				return
			}
			eventType = ""
			dataLine = ""
		}
	}

	if err := scanner.Err(); err != nil {
		ch <- APIEvent{Type: EventError, Error: fmt.Errorf("SSE scan: %w", err)}
	}
}

// dispatchEvent converts a raw SSE event to APIEvent(s) and sends them.
// Returns true when the stream should stop.
func dispatchEvent(evType, data string, pending map[int]*pendingToolUse, ch chan<- APIEvent) bool {
	switch evType {
	case "content_block_start":
		var payload struct {
			Index        int `json:"index"`
			ContentBlock struct {
				Type string `json:"type"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"content_block"`
		}
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return false
		}
		if payload.ContentBlock.Type == "tool_use" {
			pending[payload.Index] = &pendingToolUse{
				id:   payload.ContentBlock.ID,
				name: payload.ContentBlock.Name,
			}
		}

	case "content_block_delta":
		var payload struct {
			Index int `json:"index"`
			Delta struct {
				Type      string `json:"type"`
				Text      string `json:"text"`
				PartialJSON string `json:"partial_json"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return false
		}
		switch payload.Delta.Type {
		case "text_delta":
			if payload.Delta.Text != "" {
				ch <- APIEvent{Type: EventTextDelta, Text: payload.Delta.Text}
			}
		case "input_json_delta":
			if p, ok := pending[payload.Index]; ok {
				p.inputJSON.WriteString(payload.Delta.PartialJSON)
			}
		}

	case "content_block_stop":
		var payload struct {
			Index int `json:"index"`
		}
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return false
		}
		if p, ok := pending[payload.Index]; ok {
			var input map[string]any
			_ = json.Unmarshal([]byte(p.inputJSON.String()), &input)
			ch <- APIEvent{
				Type: EventToolUse,
				ToolUse: &ToolUse{
					ID:    p.id,
					Name:  p.name,
					Input: input,
				},
			}
			delete(pending, payload.Index)
		}

	case "message_delta":
		var payload struct {
			Delta struct {
				StopReason string `json:"stop_reason"`
			} `json:"delta"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return false
		}
		if payload.Delta.StopReason != "" {
			ch <- APIEvent{Type: EventMessageStop, StopReason: payload.Delta.StopReason}
			return true
		}
		// Emit usage if available
		if payload.Usage.InputTokens > 0 || payload.Usage.OutputTokens > 0 {
			ch <- APIEvent{
				Usage: &Usage{
					InputTokens:  payload.Usage.InputTokens,
					OutputTokens: payload.Usage.OutputTokens,
				},
			}
		}

	case "message_stop":
		ch <- APIEvent{Type: EventMessageStop}
		return true

	case "error":
		var payload struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		msg := data
		if err := json.Unmarshal([]byte(data), &payload); err == nil && payload.Error.Message != "" {
			msg = payload.Error.Message
		}
		ch <- APIEvent{Type: EventError, Error: fmt.Errorf("API error: %s", msg)}
		return true
	}
	return false
}
