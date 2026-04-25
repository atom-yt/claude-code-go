package apiserver

import (
	"context"
	"fmt"
	"os"

	"github.com/atom-yt/claude-code-go/internal/agent"
	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/config"
	"github.com/atom-yt/claude-code-go/internal/hooks"
	"github.com/atom-yt/claude-code-go/internal/mcp"
	"github.com/atom-yt/claude-code-go/internal/mcpresource"
	"github.com/atom-yt/claude-code-go/internal/permissions"
	"github.com/atom-yt/claude-code-go/internal/providers"
	"github.com/atom-yt/claude-code-go/internal/tools"
	"github.com/atom-yt/claude-code-go/internal/tools/ask"
	"github.com/atom-yt/claude-code-go/internal/tools/bash"
	"github.com/atom-yt/claude-code-go/internal/tools/brief"
	"github.com/atom-yt/claude-code-go/internal/tools/edit"
	"github.com/atom-yt/claude-code-go/internal/tools/glob"
	"github.com/atom-yt/claude-code-go/internal/tools/grep"
	"github.com/atom-yt/claude-code-go/internal/tools/read"
	"github.com/atom-yt/claude-code-go/internal/tools/task"
	"github.com/atom-yt/claude-code-go/internal/tools/todo"
	"github.com/atom-yt/claude-code-go/internal/tools/webfetch"
	"github.com/atom-yt/claude-code-go/internal/tools/websearch"
	"github.com/atom-yt/claude-code-go/internal/tools/write"
)

// buildClient creates API client based on config.
func (s *Server) buildClient() api.Streamer {
	provider := s.config.Provider
	baseURL := s.config.BaseURL
	apiKey := s.config.APIKey

	if provider == "" && baseURL == "" {
		return api.New(apiKey)
	}
	if provider == "anthropic" && baseURL == "" {
		return api.New(apiKey)
	}

	protocol := providers.ResolveProtocol(provider)
	if protocol == providers.ProtocolAnthropic {
		return api.NewWithBaseURL(apiKey, baseURL)
	}
	return api.NewOpenAI(apiKey, baseURL)
}

// buildRegistry creates tool registry.
func (s *Server) buildRegistry() *tools.Registry {
	r := tools.NewRegistry()
	r.Register(&read.Tool{})
	r.Register(&write.Tool{})
	r.Register(&edit.Tool{})
	r.Register(&bash.Tool{})
	r.Register(&glob.Tool{})
	r.Register(&grep.Tool{})
	r.Register(webfetch.NewTool())
	r.Register(websearch.NewTool())
	r.Register(&ask.Tool{})
	r.Register(&brief.Tool{})
	r.Register(&todo.Tool{})
	r.Register(&task.TaskCreateTool{})
	r.Register(&task.TaskGetTool{})
	r.Register(&task.TaskListTool{})
	r.Register(&task.TaskUpdateTool{})
	r.Register(&task.TaskDeleteTool{})
	r.Register(&task.TaskOutputTool{})
	return r
}

// connectMCPServers connects to MCP servers and registers their tools.
func (s *Server) connectMCPServers(ctx context.Context, registry *tools.Registry) *map[string]mcp.MCPCallTool {
	if s.config.MCPServers == nil || len(s.config.MCPServers) == 0 {
		return nil
	}
	type result struct {
		name   string
		client mcp.MCPCallTool
		err    error
	}
	ch := make(chan result, len(s.config.MCPServers))
	clients := make(map[string]mcp.MCPCallTool)

	for name, cfg := range s.config.MCPServers {
		name, cfg := name, cfg
		go func() {
			trust := cfg.Trust
			if trust == "" {
				trust = config.TrustUntrusted
			}

			var client mcp.MCPCallTool
			var err error

			switch cfg.Type {
			case "", "stdio":
				client, err = mcp.ConnectStdio(ctx, name, trust, cfg.Command, cfg.Args, cfg.Env)
			case "http", "sse":
				if cfg.URL == "" {
					err = fmt.Errorf("http/sse transport requires 'url' field")
				} else {
					client, err = mcp.ConnectHTTP(ctx, name, trust, cfg.URL)
				}
			default:
				err = fmt.Errorf("unsupported MCP transport: %q", cfg.Type)
			}

			ch <- result{name: name, client: client, err: err}
		}()
	}

	for range s.config.MCPServers {
		r := <-ch
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "MCP server %q failed: %v\n", r.name, r.err)
			continue
		}
		mcp.RegisterTools(registry, r.client)
		clients[r.name] = r.client
	}
	if len(clients) == 0 {
		return nil
	}
	return &clients
}

// buildExecutor creates hooks executor.
func (s *Server) buildExecutor() *hooks.Executor {
	if s.config.Hooks == nil || len(s.config.Hooks) == 0 {
		return nil
	}

	hookCfg := make(hooks.Config)
	for eventStr, matchers := range s.config.Hooks {
		event := hooks.Event(eventStr)
		for _, m := range matchers {
			hm := hooks.Matcher{ToolPattern: m.Matcher}
			for _, cmd := range m.Hooks {
				hm.Hooks = append(hm.Hooks, hooks.HookCommand{
					Type:    hooks.CommandType(cmd.Type),
					Command: cmd.Command,
					URL:     cmd.URL,
					Headers: cmd.Headers,
					Timeout: cmd.Timeout,
				})
			}
			hookCfg[event] = append(hookCfg[event], hm)
		}
	}

	return hooks.New(hookCfg, "apiserver-"+fmt.Sprintf("%d", os.Getpid()))
}

// buildChecker creates permission checker.
func (s *Server) buildChecker() *permissions.Checker {
	cfg := s.config.Permissions
	if cfg.DefaultMode == "" {
		cfg.DefaultMode = string(permissions.ModeDefault)
	}

	mode := permissions.Mode(cfg.DefaultMode)
	checker := permissions.New(mode)

	// Populate MCP trust levels
	if s.config.MCPTrustLevels != nil {
		checker.MCPTrustLevels = s.config.MCPTrustLevels
	}

	for _, r := range cfg.Allow {
		checker.AllowRules = append(checker.AllowRules, permissions.Rule{Tool: r.Tool, Path: r.Path, Command: r.Command})
	}
	for _, r := range cfg.Deny {
		checker.DenyRules = append(checker.DenyRules, permissions.Rule{Tool: r.Tool, Path: r.Path, Command: r.Command})
	}
	for _, r := range cfg.Ask {
		checker.AskRules = append(checker.AskRules, permissions.Rule{Tool: r.Tool, Path: r.Path, Command: r.Command})
	}

	return checker
}

// createAgent creates a new agent instance with all dependencies.
func (s *Server) createAgent(ctx context.Context, sessionID string) (*agent.Agent, error) {
	// Build API client
	client := s.buildClient()

	// Build tool registry
	registry := s.buildRegistry()

	// Connect MCP servers
	mcpClients := s.connectMCPServers(ctx, registry)
	if mcpClients != nil && len(*mcpClients) > 0 {
		listResTool := &mcpresource.ListMcpResourcesTool{}
		listResTool.SetClients(mcpClients)
		registry.Register(listResTool)

		readResTool := &mcpresource.ReadMcpResourceTool{}
		readResTool.SetClients(mcpClients)
		registry.Register(readResTool)
	}

	// Build permission checker
	checker := s.buildChecker()

	// Build hooks executor
	executor := s.buildExecutor()

	// Create agent
	ag := agent.New(client, s.config.Model, s.config.Provider, registry, checker, executor)

	// Set system prompt if provided
	if s.config.SystemPrompt != "" {
		ag.SetSystemPrompt(s.config.SystemPrompt)
	}

	// Fire session_start hook
	if executor != nil {
		go executor.FireSessionStart(ctx)
	}

	return ag, nil
}