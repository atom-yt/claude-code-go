package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAIClient is an OpenAI-compatible streaming client.
// Works with Kimi (api.moonshot.cn), OpenAI, DeepSeek, and other compatible providers.
type OpenAIClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewOpenAI returns an OpenAIClient pointing at the given base URL.
func NewOpenAI(apiKey, baseURL string) *OpenAIClient {
	return &OpenAIClient{
		APIKey:     apiKey,
		BaseURL:    strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{},
	}
}

// StreamMessages implements Streamer using the OpenAI /v1/chat/completions endpoint.
func (c *OpenAIClient) StreamMessages(ctx context.Context, req MessagesRequest) <-chan APIEvent {
	ch := make(chan APIEvent, 64)
	go c.stream(ctx, req, ch)
	return ch
}

// ---- request conversion ----

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Tools    []openAITool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    any              `json:"content,omitempty"` // string or []openAIContentPart
	ToolCallID string           `json:"tool_call_id,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
}

type openAIContentPart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type openAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openAIToolCallFunc `json:"function"`
}

type openAIToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAITool struct {
	Type     string        `json:"type"`
	Function openAIToolDef `json:"function"`
}

type openAIToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// convertMessages translates Anthropic-style messages to OpenAI format.
func convertMessages(msgs []Message) []openAIMessage {
	var out []openAIMessage
	for _, m := range msgs {
		switch m.Role {
		case RoleUser:
			// Split: tool_result blocks → separate "tool" role messages;
			// text blocks → single "user" message.
			var textParts []string
			for _, b := range m.Content {
				switch b.Type {
				case "tool_result":
					out = append(out, openAIMessage{
						Role:       "tool",
						ToolCallID: b.ToolUseID,
						Content:    b.Content,
					})
				case "text":
					textParts = append(textParts, b.Text)
				}
			}
			if len(textParts) > 0 {
				out = append(out, openAIMessage{
					Role:    "user",
					Content: strings.Join(textParts, "\n"),
				})
			}

		case RoleAssistant:
			msg := openAIMessage{Role: "assistant"}
			var textParts []string
			for _, b := range m.Content {
				switch b.Type {
				case "text":
					textParts = append(textParts, b.Text)
				case "tool_use":
					argsJSON, _ := json.Marshal(b.Input)
					msg.ToolCalls = append(msg.ToolCalls, openAIToolCall{
						ID:   b.ID,
						Type: "function",
						Function: openAIToolCallFunc{
							Name:      b.Name,
							Arguments: string(argsJSON),
						},
					})
				}
			}
			if len(textParts) > 0 {
				msg.Content = strings.Join(textParts, "\n")
			}
			out = append(out, msg)
		}
	}
	return out
}

func convertTools(specs []ToolSpec) []openAITool {
	out := make([]openAITool, len(specs))
	for i, s := range specs {
		out[i] = openAITool{
			Type: "function",
			Function: openAIToolDef{
				Name:        s.Name,
				Description: s.Description,
				Parameters:  s.InputSchema,
			},
		}
	}
	return out
}

// ---- streaming ----

func (c *OpenAIClient) stream(ctx context.Context, req MessagesRequest, ch chan<- APIEvent) {
	defer close(ch)

	oaiReq := openAIRequest{
		Model:    req.Model,
		Messages: convertMessages(req.Messages),
		Tools:    convertTools(req.Tools),
		Stream:   true,
	}

	body, err := json.Marshal(oaiReq)
	if err != nil {
		ch <- APIEvent{Type: EventError, Error: fmt.Errorf("marshal request: %w", err)}
		return
	}

	buildReq := func() (*http.Request, error) {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
			c.BaseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "text/event-stream")
		return httpReq, nil
	}

	resp, err := doWithRetry(ctx, c.HTTPClient, buildReq, DefaultRetryConfig)
	if err != nil {
		ch <- APIEvent{Type: EventError, Error: err}
		return
	}
	defer resp.Body.Close()

	parseOpenAISSE(ctx, resp.Body, ch)
}

// ---- SSE parser ----

// openAIChunk is a single chunk from the OpenAI SSE stream.
type openAIChunk struct {
	Choices []struct {
		Delta struct {
			Content   string           `json:"content"`
			ToolCalls []openAIToolCall `json:"tool_calls"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func parseOpenAISSE(ctx context.Context, r io.Reader, ch chan<- APIEvent) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Accumulate partial tool call arguments per index.
	type pendingCall struct {
		id        string
		name      string
		argsAccum strings.Builder
	}
	pending := map[int]*pendingCall{}

	// Track if we've sent a stop event
	stopSent := false

	for scanner.Scan() {
		if ctx.Err() != nil {
			if !stopSent {
				ch <- APIEvent{Type: EventError, Error: ctx.Err()}
			}
			return
		}
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			stopSent = true
			// Flush any accumulated tool calls.
			for _, p := range pending {
				var input map[string]any
				_ = json.Unmarshal([]byte(p.argsAccum.String()), &input)
				if input == nil {
					input = map[string]any{}
				}
				ch <- APIEvent{
					Type: EventToolUse,
					ToolUse: &ToolUse{
						ID:    p.id,
						Name:  p.name,
						Input: input,
					},
				}
			}
			ch <- APIEvent{Type: EventMessageStop, StopReason: "end_turn"}
			return
		}

		var chunk openAIChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 && chunk.Usage == nil {
			continue
		}

		// Handle usage if present (typically in final chunk)
		if chunk.Usage != nil {
			ch <- APIEvent{
				Usage: &Usage{
					InputTokens:  chunk.Usage.PromptTokens,
					OutputTokens: chunk.Usage.CompletionTokens,
				},
			}
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta

		if delta.Content != "" {
			ch <- APIEvent{Type: EventTextDelta, Text: delta.Content}
		}

		for _, tc := range delta.ToolCalls {
			idx := 0 // OpenAI includes index field; use map key
			// Infer index from position in slice (tc.ID is only set on first chunk).
			if _, exists := pending[idx]; !exists || tc.ID != "" {
				if tc.ID != "" {
					pending[len(pending)] = &pendingCall{id: tc.ID, name: tc.Function.Name}
					idx = len(pending) - 1
				}
			} else {
				idx = len(pending) - 1
			}
			if p, ok := pending[idx]; ok {
				p.argsAccum.WriteString(tc.Function.Arguments)
			}
		}

		if chunk.Choices[0].FinishReason != nil {
			stopSent = true
			reason := *chunk.Choices[0].FinishReason
			if reason == "tool_calls" {
				// Flush tool calls.
				for _, p := range pending {
					var input map[string]any
					_ = json.Unmarshal([]byte(p.argsAccum.String()), &input)
					if input == nil {
						input = map[string]any{}
					}
					ch <- APIEvent{
						Type: EventToolUse,
						ToolUse: &ToolUse{
							ID:    p.id,
							Name:  p.name,
							Input: input,
						},
					}
				}
				pending = map[int]*pendingCall{}
			}
			ch <- APIEvent{Type: EventMessageStop, StopReason: reason}
			return
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		if !stopSent {
			ch <- APIEvent{Type: EventError, Error: fmt.Errorf("SSE scan error: %w", err)}
		}
		return
	}

	// Stream closed normally but without explicit stop event
	if !stopSent {
		ch <- APIEvent{Type: EventMessageStop, StopReason: "stream_closed"}
	}
}
