// Package tui implements the terminal user interface using bubbletea.
package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/atom-yt/claude-code-go/internal/agent"
	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/commands"
	"github.com/atom-yt/claude-code-go/internal/config"
	"github.com/atom-yt/claude-code-go/internal/hooks"
	"github.com/atom-yt/claude-code-go/internal/mcp"
	"github.com/atom-yt/claude-code-go/internal/permissions"
	"github.com/atom-yt/claude-code-go/internal/session"
	"github.com/atom-yt/claude-code-go/internal/tools"
	toolbash "github.com/atom-yt/claude-code-go/internal/tools/bash"
	tooledit "github.com/atom-yt/claude-code-go/internal/tools/edit"
	toolglob "github.com/atom-yt/claude-code-go/internal/tools/glob"
	toolgrep "github.com/atom-yt/claude-code-go/internal/tools/grep"
	toolread "github.com/atom-yt/claude-code-go/internal/tools/read"
	toolwrite "github.com/atom-yt/claude-code-go/internal/tools/write"
)

// Role identifies the sender of a chat message.
type Role string

const (
	RoleUser         Role = "user"
	RoleAssistant    Role = "assistant"
	RoleError        Role = "error"
	RoleToolProgress Role = "tool_progress"
	RoleAsk          Role = "ask"
)

// ChatMessage holds a single conversation turn.
type ChatMessage struct {
	Role      Role
	Content   string
	Streaming bool
	rendered  string // cached rendered output (invalidated on Content change)
}

// Status is the status bar indicator.
type Status string

const (
	StatusReady    Status = "ready"
	StatusThinking Status = "thinking..."
	StatusAsking   Status = "waiting for approval..."
)

// Config carries CLI flag values from cobra.
type Config struct {
	Model   string
	APIKey  string
	Verbose bool
}

// streamEventMsg wraps agent.StreamEvent for bubbletea routing.
type streamEventMsg struct {
	event agent.StreamEvent
}

// Model is the bubbletea application state.
type Model struct {
	cfg      config.Settings
	ag       *agent.Agent
	messages []ChatMessage
	input    string
	status   Status
	width    int
	height   int
	styles   styles

	// Cached glamour renderer (invalidated when width changes).
	mdRenderer    *glamourRenderer

	// Scrolling.
	scrollOffset int

	// Input history navigation.
	inputHistory []string
	historyIdx   int
	inputDraft   string

	// Permission ask state.
	askPending bool
	askReplyCh chan<- bool

	// Cumulative token usage.
	totalInputTokens  int
	totalOutputTokens int

	// Slash commands.
	cmdRegistry *commands.Registry

	// Session persistence.
	sessionID string
}

type styles struct {
	userLabel      lipgloss.Style
	assistantLabel lipgloss.Style
	errorLabel     lipgloss.Style
	toolLabel      lipgloss.Style
	askLabel       lipgloss.Style
	messageText    lipgloss.Style
	errorText      lipgloss.Style
	askText        lipgloss.Style
	toolText       lipgloss.Style
	inputPrefix    lipgloss.Style
	inputText      lipgloss.Style
	statusBar      lipgloss.Style
	divider        lipgloss.Style
	scrollHint     lipgloss.Style
}

// NewModel creates an initialised TUI model.
func NewModel(cliCfg Config, initialPrompt string) Model {
	settings := config.Load(config.CLIFlags{
		Model:   cliCfg.Model,
		APIKey:  cliCfg.APIKey,
		Verbose: cliCfg.Verbose,
	})
	if settings.APIKey == "" {
		settings.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	m := Model{
		cfg:         settings,
		status:      StatusReady,
		styles:      buildStyles(),
		historyIdx:  -1,
		cmdRegistry: commands.NewRegistry(),
		sessionID:   session.NewID(),
		mdRenderer:  &glamourRenderer{},
	}

	if settings.APIKey != "" {
		client := api.New(settings.APIKey)
		registry := buildRegistry()

		// Connect MCP servers and register their tools.
		connectMCPServers(context.Background(), settings.MCPServers, registry)

		checker := buildChecker(settings.Permissions, &m)
		executor := buildExecutor(settings.Hooks)
		m.ag = agent.New(client, settings.Model, registry, checker, executor)

		// Fire session_start hook.
		if executor != nil {
			go executor.FireSessionStart(context.Background())
		}
	}

	if strings.TrimSpace(initialPrompt) != "" {
		m.messages = append(m.messages, ChatMessage{Role: RoleUser, Content: initialPrompt})
		if m.ag == nil {
			m.messages = append(m.messages, ChatMessage{
				Role:    RoleError,
				Content: "No API key configured. Set ANTHROPIC_API_KEY or use --api-key.",
			})
		}
	}

	return m
}

// NewModelWithHistory creates a TUI model pre-loaded with a prior session.
func NewModelWithHistory(cliCfg Config, rec session.Record) Model {
	m := NewModel(cliCfg, "")
	m.sessionID = rec.ID
	m.totalInputTokens = rec.InputTokens
	m.totalOutputTokens = rec.OutputTokens

	// Replay messages into the UI.
	for _, msg := range rec.Messages {
		switch msg.Role {
		case api.RoleUser:
			for _, block := range msg.Content {
				if block.Type == "text" && block.Text != "" {
					m.messages = append(m.messages, ChatMessage{Role: RoleUser, Content: block.Text})
				}
			}
		case api.RoleAssistant:
			var text string
			for _, block := range msg.Content {
				if block.Type == "text" {
					text += block.Text
				}
			}
			if text != "" {
				m.messages = append(m.messages, ChatMessage{Role: RoleAssistant, Content: text})
			}
		}
	}

	// Restore agent history so conversation context is preserved.
	if m.ag != nil {
		m.ag.SetHistory(rec.Messages)
	}

	return m
}

// historyHeight returns the pixel-line height of the history area.
func (m Model) historyHeight() int {
	reserved := 1 + 1 + 1 + 1 // divider + input + statusbar + blank
	h := m.height - reserved
	if h < 1 {
		h = 1
	}
	return h
}

// clampScroll keeps scrollOffset in [0, maxScroll].
func (m *Model) clampScroll() {
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	// Max scroll is computed in View; just cap at a large number.
	if m.scrollOffset > 10000 {
		m.scrollOffset = 10000
	}
}

// buildRegistry registers all built-in tools.
func buildRegistry() *tools.Registry {
	r := tools.NewRegistry()
	r.Register(&toolread.Tool{})
	r.Register(&toolwrite.Tool{})
	r.Register(&tooledit.Tool{})
	r.Register(&toolbash.Tool{})
	r.Register(&toolglob.Tool{})
	r.Register(&toolgrep.Tool{})
	return r
}

// connectMCPServers connects to all configured MCP servers concurrently
// and registers their tools into the registry.
func connectMCPServers(ctx context.Context, servers map[string]config.MCPServerConfig, registry *tools.Registry) {
	if len(servers) == 0 {
		return
	}
	type result struct {
		name   string
		client *mcp.Client
		err    error
	}
	ch := make(chan result, len(servers))

	for name, cfg := range servers {
		name, cfg := name, cfg
		go func() {
			if cfg.Type != "stdio" && cfg.Type != "" {
				ch <- result{name: name, err: fmt.Errorf("unsupported MCP transport: %q", cfg.Type)}
				return
			}
			c, err := mcp.ConnectStdio(ctx, name, cfg.Command, cfg.Args, cfg.Env)
			ch <- result{name: name, client: c, err: err}
		}()
	}

	for range servers {
		r := <-ch
		if r.err != nil {
			// Log to stderr silently; don't crash TUI.
			continue
		}
		mcp.RegisterTools(registry, r.client)
	}
}

// buildExecutor converts config hook definitions into a hooks.Executor.
// Returns nil if no hooks are configured.
func buildExecutor(cfg map[string][]config.HookMatcherConfig) *hooks.Executor {
	if len(cfg) == 0 {
		return nil
	}

	hookCfg := make(hooks.Config)
	for eventStr, matchers := range cfg {
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

	return hooks.New(hookCfg, "session-"+fmt.Sprintf("%d", os.Getpid()))
}

// buildChecker constructs a Checker from config, wiring the AskFn to the TUI.
func buildChecker(cfg config.PermissionsConfig, m *Model) *permissions.Checker {
	mode := permissions.Mode(cfg.DefaultMode)
	if mode == "" {
		mode = permissions.ModeDefault
	}

	checker := permissions.New(mode)
	for _, r := range cfg.Allow {
		checker.AllowRules = append(checker.AllowRules, permissions.Rule{Tool: r.Tool, Path: r.Path, Command: r.Command})
	}
	for _, r := range cfg.Deny {
		checker.DenyRules = append(checker.DenyRules, permissions.Rule{Tool: r.Tool, Path: r.Path, Command: r.Command})
	}
	for _, r := range cfg.Ask {
		checker.AskRules = append(checker.AskRules, permissions.Rule{Tool: r.Tool, Path: r.Path, Command: r.Command})
	}

	checker.AskFn = func(ctx context.Context, req permissions.AskRequest) (bool, string) {
		replyCh := make(chan bool, 1)

		summary := fmt.Sprintf("Allow %s", req.ToolName)
		for _, key := range []string{"command", "file_path", "path"} {
			if v, ok := req.Input[key]; ok {
				if s, _ := v.(string); s != "" {
					if len(s) > 60 {
						s = s[:57] + "..."
					}
					summary += fmt.Sprintf(" (%s)", s)
					break
				}
			}
		}
		summary += "? [y/n] "

		m.askPending = true
		m.askReplyCh = replyCh
		m.status = StatusAsking
		m.messages = append(m.messages, ChatMessage{Role: RoleAsk, Content: summary})

		select {
		case allowed := <-replyCh:
			if !allowed {
				return false, "denied by user"
			}
			return true, ""
		case <-ctx.Done():
			return false, "context cancelled"
		}
	}

	return checker
}

func buildStyles() styles {
	return styles{
		userLabel:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
		assistantLabel: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")),
		errorLabel:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")),
		toolLabel:      lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		askLabel:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")),
		messageText:    lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		errorText:      lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
		askText:        lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
		toolText:       lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		inputPrefix:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")),
		inputText:      lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		statusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("244")).
			PaddingLeft(1).PaddingRight(1),
		divider:    lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
		scrollHint: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}
}
