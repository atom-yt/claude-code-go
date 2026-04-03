package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/baidu/claude-code-go/internal/agent"
	"github.com/baidu/claude-code-go/internal/commands"
	"github.com/baidu/claude-code-go/internal/session"
)

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
		m.input += msg.String()
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

	eventCh := m.ag.Query(context.Background(), text)
	return m, waitForEvent(eventCh)
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
		GetCost: func() (int, int) {
			return m.totalInputTokens, m.totalOutputTokens
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
