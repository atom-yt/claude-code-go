// Package webfetch implements a web fetch tool for HTTP GET requests.
package webfetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/atom-yt/claude-code-go/internal/tools"
	"github.com/atom-yt/claude-code-go/internal/urlutil"
)

// Tool implements the WebFetch tool.
type Tool struct {
	validator *urlutil.URLValidator
}

var _ tools.Tool = (*Tool)(nil)

// NewTool creates a new WebFetch tool with URL validation enabled.
func NewTool() *Tool {
	return &Tool{
		validator: urlutil.NewURLValidator(),
	}
}

func (t *Tool) Name() string            { return "WebFetch" }
func (t *Tool) IsReadOnly() bool        { return true }
func (t *Tool) IsConcurrencySafe() bool { return true }

func (t *Tool) Description() string {
	return "Retrieve the content of a web page or HTTP resource using a GET request. Returns the HTTP response body as text."
}

func (t *Tool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to fetch (HTTP or HTTPS)",
			},
			"timeout_ms": map[string]any{
				"type":        "integer",
				"description": "Timeout in milliseconds (default 15000)",
			},
			"max_bytes": map[string]any{
				"type":        "integer",
				"description": "Maximum bytes to read from response (default 1MB)",
			},
		},
		"required": []string{"url"},
	}
}

func (t *Tool) Call(ctx context.Context, input map[string]any) (tools.ToolResult, error) {
	url, _ := input["url"].(string)
	if url == "" {
		return tools.ToolResult{Output: "Error: url parameter is required", IsError: true}, nil
	}

	// Validate URL for security (SSRF prevention, internal IP blocking)
	if err := t.validator.Validate(url); err != nil {
		return tools.ToolResult{Output: fmt.Sprintf("URL validation failed: %v", err), IsError: true}, nil
	}

	// Parse timeout
	timeout := 15 * time.Second
	if tm, ok := input["timeout_ms"].(float64); ok && tm > 0 {
		timeout = time.Duration(tm) * time.Millisecond
		if timeout > 60*time.Second {
			timeout = 60 * time.Second // Cap at 60 seconds
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Parse max bytes
	maxBytes := int64(1 * 1024 * 1024) // 1MB default
	if mb, ok := input["max_bytes"].(float64); ok && mb > 0 {
		maxBytes = int64(mb)
		if maxBytes > 10*1024*1024 {
			maxBytes = 10 * 1024 * 1024 // Cap at 10MB
		}
	}

	content, err := fetch(ctx, url, maxBytes)
	if err != nil {
		return tools.ToolResult{Output: fmt.Sprintf("Fetch error: %v", err), IsError: true}, nil
	}

	return tools.ToolResult{Output: content}, nil
}

// fetch performs an HTTP GET request and returns the response body.
func fetch(ctx context.Context, url string, maxBytes int64) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; claude-code-go/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{
		Timeout: 0, // Timeout handled by context
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Limit response body size
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	content := string(body)

	// If content looks like HTML, note it in output
	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/html") || strings.Contains(content, "<!DOCTYPE") || strings.Contains(content, "<html") {
		content = fmt.Sprintf("[HTML content]\n%s", content)
	}

	return content, nil
}
