package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const inputPrefix = "> "

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	histH := m.historyHeight()
	history, maxScroll := m.renderHistory(histH)
	_ = maxScroll

	divider := m.styles.divider.Render(strings.Repeat("─", m.width))
	input := m.renderInput()
	status := m.renderStatusBar()

	return strings.Join([]string{history, divider, input, status}, "\n")
}

// renderHistory renders the message log, applying scrollOffset from the bottom.
// Returns the rendered string and the total number of logical lines (for scroll capping).
func (m Model) renderHistory(maxLines int) (string, int) {
	// Build all logical lines from all messages.
	var all []string
	for i := range m.messages {
		all = append(all, m.renderMessage(i)...)
	}

	totalLines := len(all)

	// Cap scrollOffset now that we know total lines.
	offset := m.scrollOffset
	maxScroll := totalLines - maxLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if offset > maxScroll {
		offset = maxScroll
	}

	// Slice the window: offset=0 → bottom, offset=N → scroll up N lines.
	start := totalLines - maxLines - offset
	if start < 0 {
		start = 0
	}
	end := start + maxLines
	if end > totalLines {
		end = totalLines
	}

	window := all[start:end]

	// Pad top to fill the history area.
	for len(window) < maxLines {
		window = append([]string{""}, window...)
	}

	// Scroll hint in top-right corner when scrolled up.
	if offset > 0 {
		hint := m.styles.scrollHint.Render(fmt.Sprintf(" ↑ scrolled (%d lines) PgUp/PgDn ", offset))
		if len(window) > 0 {
			// Overlay hint at the end of the first line.
			first := window[0]
			firstW := lipgloss.Width(first)
			hintW := lipgloss.Width(hint)
			pad := m.width - firstW - hintW
			if pad > 0 {
				first += strings.Repeat(" ", pad) + hint
			} else {
				first = hint
			}
			window[0] = first
		}
	}

	return strings.Join(window, "\n"), maxScroll
}

// renderMessage converts one ChatMessage into display lines.
func (m Model) renderMessage(idx int) []string {
	msg := &m.messages[idx]
	var lines []string

	switch msg.Role {
	case RoleUser:
		lines = append(lines, m.styles.userLabel.Render("You:"))
		for _, l := range wordWrap(msg.Content, m.width-4) {
			lines = append(lines, "  "+m.styles.messageText.Render(l))
		}

	case RoleAssistant:
		lines = append(lines, m.styles.assistantLabel.Render("Claude:"))
		content := msg.Content
		if content == "" && msg.Streaming {
			lines = append(lines, "  "+m.styles.toolText.Render("▋"))
		} else {
			rendered := m.renderMarkdown(msg, m.width-4)
			for _, l := range strings.Split(strings.TrimRight(rendered, "\n"), "\n") {
				// Stream cursor on last line while streaming.
				if msg.Streaming && l == "" {
					continue
				}
				lines = append(lines, "  "+l)
			}
			if msg.Streaming {
				// Append cursor to last line.
				if len(lines) > 0 {
					lines[len(lines)-1] += m.styles.toolText.Render("▋")
				}
			}
		}

	case RoleToolProgress:
		icon := "  >"
		if msg.Streaming {
			icon = "  ●"
		}
		lines = append(lines, m.styles.toolLabel.Render(icon)+" "+m.styles.toolText.Render(msg.Content))

	case RoleError:
		lines = append(lines, m.styles.errorLabel.Render("Error:"))
		for _, l := range wordWrap(msg.Content, m.width-4) {
			lines = append(lines, "  "+m.styles.errorText.Render(l))
		}

	case RoleAsk:
		lines = append(lines, m.styles.askLabel.Render("  ?")+" "+m.styles.askText.Render(msg.Content))
	}

	lines = append(lines, "") // blank separator
	return lines
}

// renderMarkdown renders assistant content through glamour for Markdown formatting.
// Falls back to plain word-wrapped text on error.
func (m Model) renderMarkdown(msg *ChatMessage, width int) string {
	// Use per-message render cache when not streaming.
	if !msg.Streaming && msg.rendered != "" {
		return msg.rendered
	}

	if width < 20 {
		width = 20
	}

	r := m.mdRenderer.get(width)
	if r == nil {
		return msg.Content
	}

	out, err := r.Render(msg.Content)
	if err != nil {
		return msg.Content
	}

	// Cache for non-streaming messages.
	if !msg.Streaming {
		msg.rendered = out
	}

	return out
}

// renderInput renders the single-line input area.
func (m Model) renderInput() string {
	if m.askPending {
		return m.styles.askLabel.Render("  ? ") +
			m.styles.askText.Render("Allow? [y/n]: ") +
			m.styles.toolText.Render("█")
	}

	prefix := m.styles.inputPrefix.Render(inputPrefix)

	// Show last line of multi-line input in the input box.
	display := m.input
	if strings.Contains(display, "\n") {
		parts := strings.Split(display, "\n")
		// Show line count hint.
		prefix = m.styles.inputPrefix.Render(fmt.Sprintf("[%dL] ", len(parts)))
		display = parts[len(parts)-1]
	}

	cursor := m.styles.inputPrefix.Render("█")
	text := m.styles.inputText.Render(display) + cursor
	return prefix + text
}

// renderStatusBar renders the bottom status line.
func (m Model) renderStatusBar() string {
	model := m.cfg.Model
	if model == "" {
		model = "no model"
	}

	cwd, _ := getCWD()

	var parts []string
	parts = append(parts, "model:"+model)
	if cwd != "" {
		if len(cwd) > 30 {
			cwd = "…" + cwd[len(cwd)-29:]
		}
		parts = append(parts, "cwd:"+cwd)
	}
	if m.totalOutputTokens > 0 {
		parts = append(parts, fmt.Sprintf("tokens:%d↑%d↓", m.totalInputTokens, m.totalOutputTokens))
	}
	parts = append(parts, string(m.status))

	content := " " + strings.Join(parts, "  │  ") + " "
	padding := m.width - lipgloss.Width(content)
	if padding > 0 {
		content += strings.Repeat(" ", padding)
	}

	return m.styles.statusBar.Render(content)
}

// wordWrap wraps text at width columns.
func wordWrap(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	var lines []string
	for _, para := range strings.Split(text, "\n") {
		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		cur := ""
		for _, w := range words {
			if cur == "" {
				cur = w
			} else if len(cur)+1+len(w) <= width {
				cur += " " + w
			} else {
				lines = append(lines, cur)
				cur = w
			}
		}
		if cur != "" {
			lines = append(lines, cur)
		}
	}
	return lines
}
