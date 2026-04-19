// Package tui implements the terminal user interface using bubbletea.
package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/atom-yt/claude-code-go/internal/agent"
	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/commands"
	"github.com/atom-yt/claude-code-go/internal/config"
	"github.com/atom-yt/claude-code-go/internal/hooks"
	"github.com/atom-yt/claude-code-go/internal/mcp"
	"github.com/atom-yt/claude-code-go/internal/permissions"
	"github.com/atom-yt/claude-code-go/internal/session"
	"github.com/atom-yt/claude-code-go/internal/skills"
	"github.com/atom-yt/claude-code-go/internal/tools"
	toolask "github.com/atom-yt/claude-code-go/internal/tools/ask"
	toolbash "github.com/atom-yt/claude-code-go/internal/tools/bash"
	toolbrief "github.com/atom-yt/claude-code-go/internal/tools/brief"
	tooledit "github.com/atom-yt/claude-code-go/internal/tools/edit"
	toolglob "github.com/atom-yt/claude-code-go/internal/tools/glob"
	toolgrep "github.com/atom-yt/claude-code-go/internal/tools/grep"
	toolplanmode "github.com/atom-yt/claude-code-go/internal/tools/planmode"
	toolread "github.com/atom-yt/claude-code-go/internal/tools/read"
	tasktool "github.com/atom-yt/claude-code-go/internal/tools/task"
	tooltodo "github.com/atom-yt/claude-code-go/internal/tools/todo"
	toolwebfetch "github.com/atom-yt/claude-code-go/internal/tools/webfetch"
	toolwebsearch "github.com/atom-yt/claude-code-go/internal/tools/websearch"
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
	Model    string
	APIKey   string
	Provider string
	BaseURL  string
	Verbose  bool
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
	mdRenderer *glamourRenderer

	// Scrolling.
	scrollOffset int

	// Input history navigation.
	inputHistory []string
	historyIdx   int
	inputDraft   string

	// Permission ask state.
	askPending bool
	askReplyCh chan<- bool

	// Cumulative token usage (for cost tracking, persists across sessions).
	totalInputTokens  int
	totalOutputTokens int

	// Session token usage (for progress bar, reset on compact).
	sessionInputTokens  int
	sessionOutputTokens int

	// Context window for current model.
	contextWindow int

	// Last compact time (to prevent frequent compacts).
	lastCompactTime time.Time

	// Auto-compact configuration.
	autoCompactEnabled bool
	compactThreshold   float64
	compactCooldown    time.Duration
	compactKeepRecent  int

	// Auto-dream configuration.
	autoDreamEnabled       bool
	minConsolidateHours    int
	minConsolidateSessions int

	// Slash commands.
	cmdRegistry *commands.Registry

	// Skills system.
	skillRegistry *skills.Registry

	// Session persistence.
	sessionID string

	// Spinner animation frame index.
	spinnerIdx int

	// Autocomplete state.
	autocomplete *AutocompleteState
}

type styles struct {
	userLabel            lipgloss.Style
	assistantLabel       lipgloss.Style
	errorLabel           lipgloss.Style
	toolLabel            lipgloss.Style
	askLabel             lipgloss.Style
	messageText          lipgloss.Style
	errorText            lipgloss.Style
	askText              lipgloss.Style
	toolText             lipgloss.Style
	inputPrefix          lipgloss.Style
	inputText            lipgloss.Style
	statusBar            lipgloss.Style
	divider              lipgloss.Style
	scrollHint           lipgloss.Style
	logo                 lipgloss.Style
	tagline              lipgloss.Style
	autocompleteHeader   lipgloss.Style
	autocompleteItem     lipgloss.Style
	autocompleteSelected lipgloss.Style
}

// NewModel creates an initialised TUI model.
func NewModel(cliCfg Config, initialPrompt string) Model {
	settings := config.Load(config.CLIFlags{
		Model:    cliCfg.Model,
		APIKey:   cliCfg.APIKey,
		Provider: cliCfg.Provider,
		BaseURL:  cliCfg.BaseURL,
		Verbose:  cliCfg.Verbose,
	})
	if settings.APIKey == "" {
		settings.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	m := Model{
		cfg:           settings,
		status:        StatusReady,
		styles:        buildStyles(),
		historyIdx:    -1,
		cmdRegistry:   commands.NewRegistry(),
		skillRegistry: scanSkills(),
		sessionID:     session.NewID(),
		mdRenderer:    &glamourRenderer{},
		// Initialize compact config
		contextWindow: getContextWindow(settings.Model),
	}

	// Register skill command if skills are available
	if m.skillRegistry != nil && len(m.skillRegistry.List()) > 0 {
		m.cmdRegistry.Register(commands.NewSkillCmd(m.skillRegistry))
	}

	// Load compact configuration
	m.autoCompactEnabled = settings.AutoCompact
	if !m.autoCompactEnabled {
		// Default to true if not explicitly set
		m.autoCompactEnabled = true
	}

	if settings.CompactThreshold > 0 {
		m.compactThreshold = settings.CompactThreshold
	} else {
		m.compactThreshold = 0.8 // Default 80%
	}

	if settings.CompactCooldown > 0 {
		m.compactCooldown = time.Duration(settings.CompactCooldown) * time.Minute
	} else {
		m.compactCooldown = 5 * time.Minute // Default 5 minutes
	}

	if settings.CompactKeepRecent > 0 {
		m.compactKeepRecent = settings.CompactKeepRecent
	} else {
		m.compactKeepRecent = 10 // Default 10 messages
	}

	// Allow context window override from settings
	if settings.ContextWindow > 0 {
		m.contextWindow = settings.ContextWindow
	}

	// Load auto-dream configuration
	m.autoDreamEnabled = settings.AutoDreamEnabled

	if settings.MinConsolidateHours > 0 {
		m.minConsolidateHours = settings.MinConsolidateHours
	} else {
		m.minConsolidateHours = 24 // Default 24 hours
	}

	if settings.MinConsolidateSessions > 0 {
		m.minConsolidateSessions = settings.MinConsolidateSessions
	} else {
		m.minConsolidateSessions = 5 // Default 5 sessions
	}

	if settings.APIKey != "" {
		client := buildClient(settings)
		registry := buildRegistry()

		// Connect MCP servers and register their tools.
		connectMCPServers(context.Background(), settings.MCPServers, registry)

		// Register Skill tool if skills are available.
		if m.skillRegistry != nil && len(m.skillRegistry.List()) > 0 {
			registry.Register(skills.NewSkillTool(m.skillRegistry))
		}

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
	// Reset session tokens to 0 when loading a session
	m.sessionInputTokens = 0
	m.sessionOutputTokens = 0

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
	// Logo: 4 lines + 1 blank line = 5 lines
	reserved := 5 + 1 + 1 + 1 // logo + divider + input + statusbar
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

// providerInfo describes a known provider's default base URL and protocol.
type providerInfo struct {
	baseURL  string
	protocol string // "anthropic" or "openai"
}

// knownProviders maps provider name → connection info.
// BaseURL for OpenAI-compatible providers must include the version prefix (e.g. /v1)
// because the client appends only /chat/completions.
var knownProviders = map[string]providerInfo{
	"openai":   {"https://api.openai.com/v1", "openai"},
	"kimi":     {"https://api.moonshot.cn/v1", "openai"},
	"moonshot": {"https://api.moonshot.cn/v1", "openai"},
	"deepseek": {"https://api.deepseek.com/v1", "openai"},
	"qwen":     {"https://dashscope.aliyuncs.com/compatible-mode/v1", "openai"},
	"codex":    {"https://coder.api.visioncoder.cn/v1", "openai"},
	// ByteDance Ark — OpenAI-compatible: /v3/chat/completions (no extra /v1)
	"ark":        {"https://ark.cn-beijing.volces.com/api/coding/v3", "openai"},
	"ark-openai": {"https://ark.cn-beijing.volces.com/api/coding/v3", "openai"},
	// ByteDance Ark — Anthropic-compatible: appends /v1/messages
	"ark-anthropic": {"https://ark.cn-beijing.volces.com/api/coding", "anthropic"},
}

// modelContextWindows maps model names to their context window sizes (in tokens).
var modelContextWindows = map[string]int{
	// Claude models
	"claude-sonnet-4-6": 200000,
	"claude-opus-4-6":   200000,
	"claude-3-opus":     200000,
	"claude-3-sonnet":   200000,
	"claude-3-haiku":    200000,
	"claude-3.5-sonnet": 200000,
	"claude-3.5-haiku":  200000,
	// OpenAI GPT models
	"gpt-4o":        128000,
	"gpt-4o-mini":   128000,
	"gpt-4-turbo":   128000,
	"gpt-4":         8192,
	"gpt-3.5-turbo": 16385,
	"o1":            200000,
	"o1-mini":       128000,
	"o1-preview":    128000,
	"o3":            200000,
	"o3-mini":       200000,
	// DeepSeek models
	"deepseek-chat":  128000,
	"deepseek-coder": 128000,
	// Qwen models
	"qwen":    128000,
	"qwen2":   128000,
	"qwen-max": 32768,
	// Kimi models
	"moonshot-v1": 128000,
}

// getContextWindow returns the context window size for a given model.
// Returns 128000 for unknown models (safe default for most modern LLMs).
func getContextWindow(model string) int {
	// First check if model matches a known prefix
	for knownModel, window := range modelContextWindows {
		if model == knownModel || len(model) > len(knownModel) && model[:len(knownModel)] == knownModel {
			return window
		}
	}
	// Fallback to 128000 (safe default for most modern LLMs)
	return 128000
}

// buildClient creates the right API client based on provider/baseURL settings.
func buildClient(s config.Settings) api.Streamer {
	provider := strings.ToLower(s.Provider)

	if provider != "" && provider != "anthropic" {
		info, known := knownProviders[provider]

		baseURL := s.BaseURL
		if baseURL == "" && known {
			baseURL = info.baseURL
		}

		protocol := "openai"
		if known {
			protocol = info.protocol
		}

		if protocol == "anthropic" {
			return api.NewWithBaseURL(s.APIKey, baseURL)
		}
		return api.NewOpenAI(s.APIKey, baseURL)
	}

	// Anthropic provider with a custom base URL (e.g. ark-anthropic endpoint via --base-url).
	if provider == "anthropic" && s.BaseURL != "" {
		return api.NewWithBaseURL(s.APIKey, s.BaseURL)
	}

	// Custom base URL only → assume OpenAI-compatible.
	if s.BaseURL != "" {
		return api.NewOpenAI(s.APIKey, s.BaseURL)
	}

	// Default: Anthropic.
	return api.New(s.APIKey)
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
	r.Register(&toolwebfetch.Tool{})
	r.Register(&toolwebsearch.Tool{})
	r.Register(&toolplanmode.EnterPlanModeTool{})
	r.Register(&toolplanmode.ExitPlanModeTool{})
	r.Register(&tooltodo.Tool{})
	r.Register(&tasktool.TaskCreateTool{})
	r.Register(&tasktool.TaskGetTool{})
	r.Register(&tasktool.TaskListTool{})
	r.Register(&tasktool.TaskUpdateTool{})
	r.Register(&tasktool.TaskDeleteTool{})
	r.Register(&tasktool.TaskOutputTool{})
	r.Register(&toolask.Tool{})
	r.Register(&toolbrief.Tool{})
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
		divider:              lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
		scrollHint:           lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		logo:                 lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81")),
		tagline:              lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		autocompleteHeader:   lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("244")).Bold(true),
		autocompleteItem:     lipgloss.NewStyle().Foreground(lipgloss.Color("251")),
		autocompleteSelected: lipgloss.NewStyle().Background(lipgloss.Color("24")).Foreground(lipgloss.Color("231")).Bold(true),
	}
}

// shouldAutoCompact checks if auto-compact should be triggered.
func (m *Model) shouldAutoCompact() bool {
	if !m.autoCompactEnabled {
		return false
	}

	total := m.sessionInputTokens + m.sessionOutputTokens
	if m.contextWindow == 0 {
		return false
	}

	// Check if usage >= threshold
	if float64(total) < float64(m.contextWindow)*m.compactThreshold {
		return false
	}

	// Check cooldown period
	if !m.lastCompactTime.IsZero() && time.Since(m.lastCompactTime) < m.compactCooldown {
		return false
	}

	return true
}

// triggerAutoCompact initiates an auto-compact operation.
func (m *Model) triggerAutoCompact(ctx context.Context) error {
	// Show message in UI
	m.messages = append(m.messages, ChatMessage{
		Role:    RoleAssistant,
		Content: "Auto-compacting conversation history...",
	})

	if err := m.compactHistory(ctx); err != nil {
		m.messages = append(m.messages, ChatMessage{
			Role:    RoleError,
			Content: fmt.Sprintf("Auto-compact failed: %v", err),
		})
		return err
	}

	// Save session
	go m.saveSession()
	return nil
}

// compactHistory compacts conversation history by summarizing old messages.
func (m *Model) compactHistory(ctx context.Context) error {
	if m.ag == nil {
		return fmt.Errorf("no agent available")
	}

	// Get old messages to summarize (excluding recent ones)
	keepRecent := m.compactKeepRecent
	if len(m.messages) <= keepRecent {
		// Not enough messages to compact
		return nil
	}

	// This is a simplified compact that just keeps recent messages
	// The summary generation is a TODO for future implementation

	// Keep recent messages and add a system message about compaction
	newMessages := []ChatMessage{
		{Role: RoleUser, Content: "<!-- History compacted: older messages have been removed to save context. -->"},
	}
	newMessages = append(newMessages, m.messages[len(m.messages)-keepRecent:]...)

	m.messages = newMessages

	// Update agent history to remove old messages
	// Get current history
	history := m.ag.History()
	// Keep only recent messages from history (accounting for user/assistant pairs)
	// This is a simple approach - keep last N messages worth of history
	if len(history) > keepRecent*2 { // *2 because of user/assistant pairs
		history = history[len(history)-keepRecent*2:]
	}
	m.ag.SetHistory(history)

	// Reset session token counters
	m.sessionInputTokens = 0
	m.sessionOutputTokens = 0
	m.lastCompactTime = time.Now()

	return nil
}

// scanSkills scans the default directories for skills and returns a registry.
func scanSkills() *skills.Registry {
	registry := skills.NewRegistry()

	// Add global skills directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalSkillsDir := filepath.Join(homeDir, ".claude", "skills")
		if _, err := os.Stat(globalSkillsDir); err == nil {
			registry.AddDir(globalSkillsDir)
		}
	}

	// Add project skills directory
	projectSkillsDir := filepath.Join(".", ".claude", "skills")
	if _, err := os.Stat(projectSkillsDir); err == nil {
		registry.AddDir(projectSkillsDir)
	}

	// Scan all directories
	_ = registry.Scan()

	return registry
}
