package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/atom-yt/claude-code-go/internal/agent"
	"github.com/atom-yt/claude-code-go/internal/commands"
	"github.com/atom-yt/claude-code-go/internal/memory"
	"github.com/atom-yt/claude-code-go/internal/session"
	"github.com/atom-yt/claude-code-go/internal/subagent"
	"github.com/atom-yt/claude-code-go/internal/taskstore"
)

// spinnerFrames are the animation frames for the waiting spinner.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// spinnerTickMsg triggers a spinner frame advance.
type spinnerTickMsg struct{}

// clearStatusMsg triggers clearing of transient status messages.
type clearStatusMsg struct{}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case spinnerTickMsg:
		if m.status == StatusThinking {
			m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
			return m, spinnerTick()
		}
		return m, nil

	case streamEventMsg:
		return m.handleStreamEvent(msg.event)

	case tea.KeyMsg:
		if m.askPending {
			return m.handleAskKey(msg)
		}
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey routes normal (non-ask) keyboard events.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 如果自动补全处于活动状态，首先处理补全特定的按键
	if m.isAutocompleteActive() {
		switch msg.Type {
		case tea.KeyTab:
			m.cycleAutocomplete()
			return m, nil
		case tea.KeyUp:
			m.selectPrevAutocomplete()
			return m, nil
		case tea.KeyDown:
			m.selectNextAutocomplete()
			return m, nil
		case tea.KeyEnter:
			m.acceptAutocomplete()
			return m.handleSubmit()
		case tea.KeyEsc:
			m.hideAutocomplete()
			return m, nil
		case tea.KeyBackspace, tea.KeyDelete:
			// 在自动补全激活时处理退格键
			runes := []rune(m.input)
			if len(runes) > 0 {
				m.input = string(runes[:len(runes)-1])
				// 更新自动补全查询
				if strings.HasPrefix(m.input, "/") {
					m.updateAutocompleteQuery(m.input)
				} else {
					m.hideAutocomplete()
				}
			}
			return m, nil
		}
	}

	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyCtrlD:
		return m, tea.Quit

	case tea.KeyCtrlL:
		// Clear screen: remove all messages.
		m.messages = nil
		m.scrollOffset = 0
		return m, nil

	case tea.KeyEnter:
		return m.handleSubmit()

	// Alt+Enter / Ctrl+J inserts a newline (Shift+Enter terminal encoding varies).
	case tea.KeyCtrlJ:
		m.input += "\n"
		return m, nil

	case tea.KeyBackspace, tea.KeyDelete:
		runes := []rune(m.input)
		if len(runes) > 0 {
			m.input = string(runes[:len(runes)-1])
		}
		return m, nil

	case tea.KeyUp:
		// Navigate input history upward.
		if len(m.inputHistory) > 0 && m.historyIdx < len(m.inputHistory)-1 {
			// Save current draft on first navigation.
			if m.historyIdx == -1 {
				m.inputDraft = m.input
			}
			m.historyIdx++
			m.input = m.inputHistory[len(m.inputHistory)-1-m.historyIdx]
		}
		return m, nil

	case tea.KeyDown:
		if m.historyIdx > 0 {
			m.historyIdx--
			m.input = m.inputHistory[len(m.inputHistory)-1-m.historyIdx]
		} else if m.historyIdx == 0 {
			m.historyIdx = -1
			m.input = m.inputDraft
		}
		return m, nil

	case tea.KeyPgUp:
		m.scrollOffset += m.historyHeight() / 2
		m.clampScroll()
		return m, nil

	case tea.KeyPgDown:
		m.scrollOffset -= m.historyHeight() / 2
		m.clampScroll()
		return m, nil

	case tea.KeyRunes:
		s := msg.String()
		// Filter out terminal control sequences and non-printable characters.
		// Also filter out OSC sequences like ]11;rgb:... and ANSI escape sequences
		if s != "" && !strings.ContainsAny(s, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0e\x0f\x1b\x7f") {
			// Filter out OSC sequences (] followed by digits and ;)
			if strings.Contains(s, "]\x1b") || strings.Contains(s, "]11;") || strings.Contains(s, "rgb:") {
				return m, nil
			}
			// Filter out ANSI CSI sequences (e.g., alt+\[24;1R, \[A, \[24;1R)
			if strings.HasPrefix(s, "alt+\\[") || strings.Contains(s, "\\[") {
				return m, nil
			}
			// Filter out bracket sequences like [24;1R (device status reports)
			if matchesBracketSequence(s) {
				return m, nil
			}
			clean := true
			for _, r := range s {
				if r < 32 && r != '\t' {
					clean = false
					break
				}
			}
			if clean {
				m.input += s

				// 如果用户输入了 /，触发自动补全
				if strings.HasPrefix(m.input, "/") {
					// 检查是否在第一行（不包含换行符）
					if !strings.Contains(strings.TrimLeft(m.input, "/"), "\n") {
						if m.isAutocompleteActive() {
							m.updateAutocompleteQuery(m.input)
						} else {
							m.showAutocomplete(m.input)
						}
					}
				} else {
					m.hideAutocomplete()
				}
			}
		}
		return m, nil
	}

	return m, nil
}

// handleMouse handles scroll wheel events.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.scrollOffset += 3
		m.clampScroll()
	case tea.MouseButtonWheelDown:
		m.scrollOffset -= 3
		m.clampScroll()
	}
	return m, nil
}

// handleSubmit processes Enter when input is non-empty.
func (m Model) handleSubmit() (tea.Model, tea.Cmd) {
	// Close autocomplete menu
	m.hideAutocomplete()

	// Clear status messages when user submits new input
	m.compactMessage = ""
	m.consolidateMessage = ""

	text := trimSpace(m.input)
	if text == "" {
		return m, nil
	}

	// Save to input history.
	if len(m.inputHistory) == 0 || m.inputHistory[len(m.inputHistory)-1] != text {
		m.inputHistory = append(m.inputHistory, text)
		if len(m.inputHistory) > 100 {
			m.inputHistory = m.inputHistory[1:]
		}
	}
	m.historyIdx = -1
	m.inputDraft = ""
	m.input = ""
	m.scrollOffset = 0

	// Slash command?
	if strings.HasPrefix(text, "/") {
		return m.handleSlashCommand(text)
	}

	m.messages = append(m.messages, ChatMessage{Role: RoleUser, Content: text})

	if m.ag == nil {
		m.messages = append(m.messages, ChatMessage{
			Role:    RoleError,
			Content: "No API key configured. Set ANTHROPIC_API_KEY or use --api-key.",
		})
		return m, nil
	}

	m.messages = append(m.messages, ChatMessage{Role: RoleAssistant, Content: "", Streaming: true})
	m.status = StatusThinking
	m.spinnerIdx = 0

	eventCh := m.ag.Query(context.Background(), text)
	return m, tea.Batch(waitForEvent(eventCh), spinnerTick())
}

// handleSlashCommand parses and executes a "/" prefixed command.
func (m Model) handleSlashCommand(text string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(text)
	name := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	if m.cmdRegistry == nil {
		m.messages = append(m.messages, ChatMessage{Role: RoleError, Content: "command registry not initialised"})
		return m, nil
	}

	cmd, ok := m.cmdRegistry.Get(name)
	if !ok {
		m.messages = append(m.messages, ChatMessage{
			Role:    RoleError,
			Content: fmt.Sprintf("Unknown command: /%s  (try /help)", name),
		})
		return m, nil
	}

	cmdCtx := m.buildCommandContext()
	output, err := cmd.Execute(context.Background(), args, cmdCtx)
	if err != nil {
		m.messages = append(m.messages, ChatMessage{Role: RoleError, Content: err.Error()})
		return m, nil
	}
	if output != "" {
		m.messages = append(m.messages, ChatMessage{Role: RoleAssistant, Content: output})
	}
	return m, nil
}

// buildCommandContext wires Command.Context callbacks to the Model.
func (m *Model) buildCommandContext() *commands.Context {
	return &commands.Context{
		ClearMessages: func() {
			m.messages = nil
			m.scrollOffset = 0
		},
		GetModel: func() string { return m.cfg.Model },
		SetModel: func(model string) {
			m.cfg.Model = model
			if m.ag != nil {
				m.ag.SetModel(model)
			}
		},
		GetProvider: func() string { return m.cfg.Provider },
		SetProvider: func(provider string) {
			m.cfg.Provider = provider
			// Rebuild the API client with the new provider.
			client := buildClient(m.cfg)
			if m.ag != nil {
				m.ag.SetClient(client)
			}
		},
		GetCost: func() (int, int) {
			return m.totalInputTokens, m.totalOutputTokens
		},
		GetTaskManager: func() *taskstore.Store {
			return m.taskManager
		},
		GetSubagentRuntime: func() *subagent.Runtime {
			return m.subagentRuntime
		},
		GetTaskCount: func() int {
			if m.subagentRuntime != nil {
				return m.subagentRuntime.GetSubagentCount()
			}
			return 0
		},
		CompactHistory: func(ctx context.Context) error {
			return m.compactHistory(ctx)
		},
		ConsolidateMemory: func(ctx context.Context) (string, error) {
			return m.consolidateMemory(ctx)
		},
		ConsolidateStatus: func(ctx context.Context) (string, error) {
			return m.consolidateStatus(ctx)
		},
		GetConfig: func() map[string]any {
			return map[string]any{
				"model":                    m.cfg.Model,
				"provider":                 m.cfg.Provider,
				"baseURL":                  m.cfg.BaseURL,
				"verbose":                  m.cfg.Verbose,
				"permissions":              m.cfg.Permissions,
				"hooks":                    m.cfg.Hooks,
				"mcpServers":               m.cfg.MCPServers,
				"autoCompact":              m.cfg.AutoCompact,
				"compactThreshold":         m.cfg.CompactThreshold,
				"compactCooldown":          m.cfg.CompactCooldown,
				"compactKeepRecent":        m.cfg.CompactKeepRecent,
				"contextWindow":            m.cfg.ContextWindow,
				"autoDreamEnabled":        m.cfg.AutoDreamEnabled,
				"autoMemoryDirectory":      m.cfg.AutoMemoryDirectory,
				"minConsolidateHours":     m.cfg.MinConsolidateHours,
				"minConsolidateSessions":  m.cfg.MinConsolidateSessions,
			}
		},
	}
}

// handleStreamEvent processes one agent.StreamEvent.
func (m Model) handleStreamEvent(ev agent.StreamEvent) (tea.Model, tea.Cmd) {
	switch ev.Type {

	case agent.EventTextDelta:
		if len(m.messages) > 0 {
			last := &m.messages[len(m.messages)-1]
			if last.Role == RoleAssistant && last.Streaming {
				last.Content += ev.Text
				last.rendered = "" // invalidate cached render
			}
		}
		if ev.NextCmd != nil {
			return m, ev.NextCmd
		}
		return m, nil

	case agent.EventToolCall:
		summary := summariseInput(ev.ToolInput)
		label := fmt.Sprintf("Running %s%s", ev.ToolName, summary)
		m.messages = append(m.messages, ChatMessage{
			Role:      RoleToolProgress,
			Content:   label,
			Streaming: true,
		})
		if ev.NextCmd != nil {
			return m, ev.NextCmd
		}
		return m, nil

	case agent.EventToolResult:
		for i := len(m.messages) - 1; i >= 0; i-- {
			if m.messages[i].Role == RoleToolProgress && m.messages[i].Streaming {
				m.messages[i].Streaming = false
				if ev.ToolIsError {
					m.messages[i].Content += " ✗"
				} else {
					m.messages[i].Content += " ✓"
				}
				break
			}
		}
		last := m.messages[len(m.messages)-1]
		if last.Role != RoleAssistant {
			m.messages = append(m.messages, ChatMessage{
				Role:      RoleAssistant,
				Content:   "",
				Streaming: true,
			})
		}
		if ev.NextCmd != nil {
			return m, ev.NextCmd
		}
		return m, nil

	case agent.EventDone:
		for i := len(m.messages) - 1; i >= 0; i-- {
			if m.messages[i].Role == RoleAssistant {
				m.messages[i].Streaming = false
				if m.messages[i].Content == "" {
					m.messages = append(m.messages[:i], m.messages[i+1:]...)
				}
				break
			}
		}
		m.status = StatusReady
		m.scrollOffset = 0
		// Persist session asynchronously.
		go m.saveSession()

		// Check if we should auto-compact
		if m.shouldAutoCompact() {
			go m.triggerAutoCompact(context.Background())
		}

		// Trigger auto-dream if enabled
		if m.autoDreamEnabled {
			go m.triggerAutoDream(context.Background())
		}
		return m, nil

	case agent.EventError:
		for len(m.messages) > 0 && m.messages[len(m.messages)-1].Streaming {
			m.messages = m.messages[:len(m.messages)-1]
		}
		m.messages = append(m.messages, ChatMessage{
			Role:    RoleError,
			Content: fmt.Sprintf("Error: %v", ev.Error),
		})
		m.status = StatusReady
		return m, nil
	}

	// Handle usage events (forwarded from agent)
	if ev.Usage != nil {
		m.sessionInputTokens += ev.Usage.InputTokens
		m.sessionOutputTokens += ev.Usage.OutputTokens
		m.totalInputTokens += ev.Usage.InputTokens
		m.totalOutputTokens += ev.Usage.OutputTokens
	}

	return m, nil
}

// handleAskKey processes y/n while a permission ask is pending.
func (m Model) handleAskKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var answer string
	switch msg.Type {
	case tea.KeyRunes:
		answer = strings.ToLower(msg.String())
	case tea.KeyEnter:
		answer = "y"
	case tea.KeyCtrlC, tea.KeyCtrlD:
		answer = "n"
	}
	switch answer {
	case "y", "yes":
		m.resolveAsk(true)
	case "n", "no":
		m.resolveAsk(false)
	}
	return m, nil
}

// resolveAsk delivers the user's decision to the waiting AskFn goroutine.
func (m *Model) resolveAsk(allowed bool) {
	if m.askReplyCh != nil {
		m.askReplyCh <- allowed
		m.askReplyCh = nil
	}
	m.askPending = false
	m.status = StatusThinking
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == RoleAsk {
			if allowed {
				m.messages[i].Content += " → allowed"
			} else {
				m.messages[i].Content += " → denied"
			}
			m.messages[i].Role = RoleToolProgress
			break
		}
	}
}

// waitForEvent returns a tea.Cmd that reads the next StreamEvent from ch.
func waitForEvent(ch <-chan agent.StreamEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return streamEventMsg{event: agent.StreamEvent{Type: agent.EventDone}}
		}
		if ev.Type == agent.EventTextDelta ||
			ev.Type == agent.EventToolCall ||
			ev.Type == agent.EventToolResult {
			ev.NextCmd = waitForEvent(ch)
		}
		return streamEventMsg{event: ev}
	}
}

// summariseInput creates a short display string for a tool call.
func summariseInput(input map[string]any) string {
	if input == nil {
		return ""
	}
	for _, key := range []string{"command", "file_path", "path", "pattern"} {
		if v, ok := input[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				if len(s) > 50 {
					s = s[:47] + "..."
				}
				return fmt.Sprintf("(%s)", s)
			}
		}
	}
	return ""
}

// truncateOutput truncates tool output for display.
func truncateOutput(output string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 500
	}
	if len(output) <= maxLen {
		return output
	}
	return output[:maxLen] + fmt.Sprintf("\n... [truncated, %d more bytes]", len(output)-maxLen)
}

// saveSession persists the current agent history to disk.
func (m *Model) saveSession() {
	if m.ag == nil || m.sessionID == "" {
		return
	}
	rec := session.Record{
		ID:           m.sessionID,
		Model:        m.cfg.Model,
		Messages:     m.ag.History(),
		InputTokens:  m.totalInputTokens,
		OutputTokens: m.totalOutputTokens,
	}
	_ = session.Save(rec) // ignore errors silently
}

func trimSpace(s string) string {
	runes := []rune(s)
	start := 0
	for start < len(runes) && isSpace(runes[start]) {
		start++
	}
	end := len(runes)
	for end > start && isSpace(runes[end-1]) {
		end--
	}
	return string(runes[start:end])
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// spinnerTick returns a tea.Cmd that fires a spinnerTickMsg after a short delay.
func spinnerTick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(_ time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

// consolidateMemory performs memory consolidation.
func (m *Model) consolidateMemory(ctx context.Context) (string, error) {
	if m.ag == nil {
		return "", fmt.Errorf("no agent available")
	}

	m.consolidateRunning = true

	// Create a consolidator
	consolidator, err := memory.NewConsolidator(m.cfg, m.ag.GetClient())
	if err != nil {
		m.consolidateRunning = false
		return "", fmt.Errorf("failed to create consolidator: %w", err)
	}

	// Perform consolidation
	result, err := consolidator.Consolidate(ctx)
	if err != nil {
		m.consolidateRunning = false
		return "", fmt.Errorf("consolidation failed: %w", err)
	}

	m.consolidateRunning = false
	m.consolidateMessage = "Memory consolidated"
	return result, nil
}

// consolidateStatus returns the current consolidation status.
func (m *Model) consolidateStatus(ctx context.Context) (string, error) {
	return memory.BuildStatusPrompt()
}

// triggerAutoDream triggers auto-dream consolidation in the background.
func (m *Model) triggerAutoDream(ctx context.Context) {
	if m.ag == nil {
		return
	}

	// Create a consolidator
	consolidator, err := memory.NewConsolidator(m.cfg, m.ag.GetClient())
	if err != nil {
		return
	}

	// Check if consolidation should run
	should, reason := consolidator.ShouldConsolidate(ctx)
	if !should {
		return
	}

	// Show a brief message in the UI
	// Note: We can't directly modify m.messages from a goroutine safely,
	// so we skip the UI notification for now
	_ = reason

	// Run consolidation in the background
	consolidator.RunBackgroundConsolidation(ctx)
}

// matchesBracketSequence detects ANSI CSI sequences like [24;1R, [A, [K, etc.
// These are terminal escape sequences that should be filtered out.
func matchesBracketSequence(s string) bool {
	// Pattern: [ followed by optional digits and semicolons, ending with a single letter
	// Examples: [24;1R, [A, [K, [1;24r, [?1049l
	if len(s) == 0 {
		return false
	}

	// Check if it starts with [
	if s[0] == '[' {
		// Must have at least 2 characters: [ and a letter
		if len(s) < 2 {
			return false
		}

		// Last character must be a letter
		lastChar := s[len(s)-1]
		if !((lastChar >= 'a' && lastChar <= 'z') || (lastChar >= 'A' && lastChar <= 'Z')) {
			return false
		}

		// Characters between [ and last letter must be: ?, ;, or digits
		for i := 1; i < len(s)-1; i++ {
			c := s[i]
			if c != '?' && c != ';' && !(c >= '0' && c <= '9') {
				return false
			}
		}
		return true
	}

	return false
}
