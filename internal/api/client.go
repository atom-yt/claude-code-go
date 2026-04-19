package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

// New returns a Client pointing at the default Anthropic API.
func New(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{},
	}
}

// NewWithBaseURL returns a Client pointing at a custom Anthropic-compatible base URL.
func NewWithBaseURL(apiKey, baseURL string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    strings.TrimRight(baseURL, "/"),
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

	buildReq := func() (*http.Request, error) {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
			c.BaseURL+"/v1/messages", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("x-api-key", c.APIKey)
		httpReq.Header.Set("anthropic-version", anthropicVersion)
		httpReq.Header.Set("content-type", "application/json")
		httpReq.Header.Set("accept", "text/event-stream")
		return httpReq, nil
	}

	resp, err := doWithRetry(ctx, c.HTTPClient, buildReq, DefaultRetryConfig)
	if err != nil {
		ch <- APIEvent{Type: EventError, Error: err}
		return
	}
	defer resp.Body.Close()

	parseSSE(ctx, resp.Body, ch)
}
