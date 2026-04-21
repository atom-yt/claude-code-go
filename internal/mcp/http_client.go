package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// TransportType indicates the transport protocol used for MCP.
type TransportType string

const (
	TransportStdio TransportType = "stdio"
	TransportHTTP  TransportType = "http"
)

// HTTPClient extends Client for HTTP transport.
type HTTPClient struct {
	name      string
	trust     string
	transport TransportType
	baseURL   string
	httpClient *http.Client

	mu      sync.Mutex
	nextID  atomic.Int64
	pending map[int]chan response

	Tools []ToolDef // populated after Connect()

	// Health tracking
	healthMu          sync.RWMutex
	healthStatus       HealthStatus
	consecutiveErrors int
	lastError         error
	lastErrorTime     time.Time
}

// ConnectHTTP connects to an MCP server via HTTP/SSE and performs the handshake.
func ConnectHTTP(ctx context.Context, name, trust, baseURL string) (*HTTPClient, error) {
	c := &HTTPClient{
		name:      name,
		trust:     trust,
		transport: TransportHTTP,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		pending:    make(map[int]chan response),
		healthStatus: HealthHealthy,
	}

	// Handshake.
	initCtx, cancel := context.WithTimeout(ctx, initTimeout)
	defer cancel()

	if err := c.initialize(initCtx); err != nil {
		return nil, fmt.Errorf("mcp http init %q: %w", name, err)
	}

	if err := c.listTools(initCtx); err != nil {
		return nil, fmt.Errorf("mcp http list tools %q: %w", name, err)
	}

	return c, nil
}

// Close is a no-op for HTTP clients.
func (c *HTTPClient) Close() {
	// HTTP clients don't need explicit cleanup
}

// Name returns the client's name.
func (c *HTTPClient) Name() string {
	return c.name
}

// TrustLevel returns the client's trust level.
func (c *HTTPClient) TrustLevel() string {
	return c.trust
}

// GetTools returns the list of tools exposed by the MCP server.
func (c *HTTPClient) GetTools() []ToolDef {
	return c.Tools
}

// CallTool invokes a tool on the MCP server and returns its text output.
func (c *HTTPClient) CallTool(ctx context.Context, name string, args map[string]any) (string, bool, error) {
	params := toolsCallParams{Name: name, Arguments: args}
	var result toolsCallResult
	if err := c.call(ctx, "tools/call", params, &result); err != nil {
		return "", true, err
	}

	var parts []string
	for _, b := range result.Content {
		if b.Type == "text" {
			parts = append(parts, b.Text)
		}
	}
	return joinStrings(parts, "\n"), result.IsError, nil
}

// ListResources returns all resources exposed by the MCP server.
func (c *HTTPClient) ListResources(ctx context.Context) ([]ResourceDef, error) {
	var result resourcesListResult
	if err := c.call(ctx, "resources/list", nil, &result); err != nil {
		return nil, err
	}
	return result.Resources, nil
}

// ReadResource reads the content of a resource by URI.
func (c *HTTPClient) ReadResource(ctx context.Context, uri string) (string, error) {
	params := resourceReadParams{URI: uri}
	var result resourceReadResult
	if err := c.call(ctx, "resources/read", params, &result); err != nil {
		return "", err
	}
	if len(result.Contents) == 0 {
		return "", fmt.Errorf("empty resource content")
	}
	return result.Contents[0].Text, nil
}

// ---- internal ----

func (c *HTTPClient) initialize(ctx context.Context) error {
	params := initializeParams{
		ProtocolVersion: "2024-11-05",
		ClientInfo:      clientInfo{Name: "claude-code-go", Version: "0.1.0"},
	}
	var result initializeResult
	if err := c.call(ctx, "initialize", params, &result); err != nil {
		return err
	}
	// Send initialized notification.
	return c.notify("notifications/initialized", nil)
}

func (c *HTTPClient) listTools(ctx context.Context) error {
	var result toolsListResult
	if err := c.call(ctx, "tools/list", nil, &result); err != nil {
		return err
	}
	c.Tools = result.Tools
	return nil
}

// call sends a JSON-RPC request via HTTP and waits for the response.
func (c *HTTPClient) call(ctx context.Context, method string, params any, out any) error {
	id := int(c.nextID.Add(1))

	req := request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/mcp", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http error %d: %s", resp.StatusCode, string(respBody))
	}

	var rpcResp response
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("rpc %s: %s (code %d)", method, rpcResp.Error.Message, rpcResp.Error.Code)
	}

	if out != nil && rpcResp.Result != nil {
		return json.Unmarshal(rpcResp.Result, out)
	}

	return nil
}

// notify sends a JSON-RPC notification via HTTP.
func (c *HTTPClient) notify(method string, params any) error {
	type notif struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params,omitempty"`
	}

	body, err := json.Marshal(notif{JSONRPC: "2.0", Method: method, Params: params})
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/mcp", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create notification request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http notification error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
// HealthStatus returns the current health status of the MCP server.
func (c *HTTPClient) HealthStatus() HealthStatus {
	c.healthMu.RLock()
	defer c.healthMu.RUnlock()
	return c.healthStatus
}

// HealthError returns the last error if unhealthy, nil if healthy.
func (c *HTTPClient) HealthError() error {
	c.healthMu.RLock()
	defer c.healthMu.RUnlock()
	if c.healthStatus == HealthHealthy {
		return nil
	}
	return c.lastError
}

// CheckHealth performs a health check by attempting to list tools.
// Returns the current health status and any error encountered.
func (c *HTTPClient) CheckHealth(ctx context.Context) (HealthStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, HealthCheckTimeout)
	defer cancel()

	// Try to list tools as a simple health check
	var result toolsListResult
	err := c.call(ctx, "tools/list", nil, &result)

	c.healthMu.Lock()
	defer c.healthMu.Unlock()

	if err != nil {
		c.consecutiveErrors++
		c.lastError = err
		c.lastErrorTime = time.Now()

		// Mark unhealthy after consecutive failures
		if c.consecutiveErrors >= MaxHealthCheckFailures {
			c.healthStatus = HealthUnhealthy
		}
		return c.healthStatus, err
	}

	// Success - reset error counter and mark healthy
	c.consecutiveErrors = 0
	c.lastError = nil
	c.healthStatus = HealthHealthy
	return c.healthStatus, nil
}

// ResetHealth resets the health status to healthy.
// Useful after reconnection attempts.
func (c *HTTPClient) ResetHealth() {
	c.healthMu.Lock()
	defer c.healthMu.Unlock()
	c.healthStatus = HealthHealthy
	c.consecutiveErrors = 0
	c.lastError = nil
}
