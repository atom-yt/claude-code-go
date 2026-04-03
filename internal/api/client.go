package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultBaseURL        = "https://api.anthropic.com"
	anthropicVersion      = "2023-06-01"
	defaultMaxTokens      = 8096
)

// Client is an Anthropic Messages API client.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// New returns a Client with sensible defaults.
func New(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{},
	}
}

// StreamMessages sends a streaming messages request and returns a channel of
// APIEvent values. The channel is closed after EventMessageStop or EventError.
func (c *Client) StreamMessages(ctx context.Context, req MessagesRequest) <-chan APIEvent {
	ch := make(chan APIEvent, 64)
	req.Stream = true
	if req.MaxTokens == 0 {
		req.MaxTokens = defaultMaxTokens
	}
	go c.stream(ctx, req, ch)
	return ch
}

func (c *Client) stream(ctx context.Context, req MessagesRequest, ch chan<- APIEvent) {
	defer close(ch)

	body, err := json.Marshal(req)
	if err != nil {
		ch <- APIEvent{Type: EventError, Error: fmt.Errorf("marshal request: %w", err)}
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		ch <- APIEvent{Type: EventError, Error: fmt.Errorf("build request: %w", err)}
		return
	}

	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("accept", "text/event-stream")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		ch <- APIEvent{Type: EventError, Error: fmt.Errorf("http request: %w", err)}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		ch <- APIEvent{Type: EventError, Error: fmt.Errorf("API error %d: %s", resp.StatusCode, string(b))}
		return
	}

	parseSSE(ctx, resp.Body, ch)
}
