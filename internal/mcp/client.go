package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const initTimeout = 10 * time.Second

// Client connects to one MCP server and wraps its tools as Tool objects.
type Client struct {
	name    string
	trust   string // Trust level: "full", "limited", "untrusted"
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  *bufio.Reader

	mu      sync.Mutex
	nextID  atomic.Int64
	pending map[int]chan response

	Tools []ToolDef // populated after Connect()
}

// ConnectStdio starts an MCP server via stdio and performs the handshake.
func ConnectStdio(ctx context.Context, name, trust string, command string, args []string, env []string) (*Client, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = append(cmd.Environ(), env...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("mcp stdin: %w", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("mcp stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("mcp start %q: %w", command, err)
	}

	c := &Client{
		name:    name,
		trust:   trust,
		cmd:     cmd,
		stdin:   stdin,
		stdout:  bufio.NewReader(stdoutPipe),
		pending: make(map[int]chan response),
	}

	// Read loop.
	go c.readLoop()

	// Handshake.
	initCtx, cancel := context.WithTimeout(ctx, initTimeout)
	defer cancel()

	if err := c.initialize(initCtx); err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("mcp init %q: %w", name, err)
	}

	if err := c.listTools(initCtx); err != nil {
		return nil, fmt.Errorf("mcp list tools %q: %w", name, err)
	}

	return c, nil
}

// Close terminates the MCP server process.
func (c *Client) Close() {
	_ = c.stdin.Close()
	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
}

// Name returns the client's name.
func (c *Client) Name() string {
	return c.name
}

// TrustLevel returns the client's trust level.
func (c *Client) TrustLevel() string {
	return c.trust
}

// GetTools returns the list of tools exposed by the MCP server.
func (c *Client) GetTools() []ToolDef {
	return c.Tools
}

// ListResources returns all resources exposed by the MCP server.
func (c *Client) ListResources(ctx context.Context) ([]ResourceDef, error) {
	var result resourcesListResult
	if err := c.call(ctx, "resources/list", nil, &result); err != nil {
		return nil, err
	}
	return result.Resources, nil
}

// CallTool invokes a tool on the MCP server and returns its text output.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (string, bool, error) {
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
	return strings.Join(parts, "\n"), result.IsError, nil
}

// ---- internal ----

func (c *Client) initialize(ctx context.Context) error {
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

func (c *Client) listTools(ctx context.Context) error {
	var result toolsListResult
	if err := c.call(ctx, "tools/list", nil, &result); err != nil {
		return err
	}
	c.Tools = result.Tools
	return nil
}

// ReadResource reads the content of a resource by URI.
func (c *Client) ReadResource(ctx context.Context, uri string) (string, error) {
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

// call sends a JSON-RPC request and waits for the response.
func (c *Client) call(ctx context.Context, method string, params any, out any) error {
	id := int(c.nextID.Add(1))

	req := request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	replyCh := make(chan response, 1)
	c.mu.Lock()
	c.pending[id] = replyCh
	c.mu.Unlock()

	if err := c.send(req); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return err
	}

	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return ctx.Err()
	case resp := <-replyCh:
		if resp.Error != nil {
			return fmt.Errorf("rpc %s: %s (code %d)", method, resp.Error.Message, resp.Error.Code)
		}
		if out != nil && resp.Result != nil {
			return json.Unmarshal(resp.Result, out)
		}
		return nil
	}
}

// notify sends a JSON-RPC notification (no id, no reply expected).
func (c *Client) notify(method string, params any) error {
	type notif struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params,omitempty"`
	}
	return c.send(notif{JSONRPC: "2.0", Method: method, Params: params})
}

func (c *Client) send(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	c.mu.Lock()
	_, err = c.stdin.Write(data)
	c.mu.Unlock()
	return err
}

// readLoop reads newline-delimited JSON responses from the server.
func (c *Client) readLoop() {
	for {
		line, err := c.stdout.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var resp response
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			continue
		}
		c.mu.Lock()
		ch, ok := c.pending[resp.ID]
		if ok {
			delete(c.pending, resp.ID)
		}
		c.mu.Unlock()
		if ok {
			ch <- resp
		}
	}
}
