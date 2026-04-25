package feishu

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/interfaces"
)

// Formatter converts Agent output to Feishu message formats.
type Formatter struct {
	config     *Config
	markdown   bool
	cards      bool
	truncate   int
}

// NewFormatter creates a new formatter.
func NewFormatter(cfg *Config) *Formatter {
	return &Formatter{
		config:   cfg,
		markdown: cfg.EnableMarkdown,
		cards:    cfg.EnableCards,
		truncate: cfg.TruncateLength,
	}
}

// FormatText formats text content.
func (f *Formatter) FormatText(content string) string {
	content = strings.TrimSpace(content)

	// Truncate if needed
	if f.truncate > 0 && len(content) > f.truncate {
		content = content[:f.truncate] + "..."
	}

	return content
}

// FormatMarkdown formats markdown content for Feishu.
func (f *Formatter) FormatMarkdown(content string) string {
	if !f.markdown {
		return f.FormatText(content)
	}

	content = strings.TrimSpace(content)

	// Feishu markdown support
	// Convert common markdown patterns
	content = f.convertMarkdown(content)

	// Truncate
	if f.truncate > 0 && len(content) > f.truncate {
		content = content[:f.truncate] + "..."
	}

	return content
}

// convertMarkdown converts common markdown to Feishu format.
func (f *Formatter) convertMarkdown(content string) string {
	// Feishu supports most common markdown
	// Main conversions needed:

	// Code blocks ``` -> [code]...[/code]
	content = strings.ReplaceAll(content, "```", "[code]")

	// Inline code ` -> [code]...[/code] (simplified)
	// Feishu handles this natively in most cases

	// Bold ** -> **
	// Feishu supports this natively

	// Italic * -> *
	// Feishu supports this natively

	// Links [text](url) -> <url|text>
	content = f.convertLinks(content)

	return content
}

// convertLinks converts markdown links to Feishu format.
func (f *Formatter) convertLinks(content string) string {
	// Simple markdown link detection
	// [text](url) -> <url|text>

	start := 0
	result := strings.Builder{}

	for i := 0; i < len(content); i++ {
		if content[i] == '[' && i+1 < len(content) {
			// Found potential link start
			endBracket := strings.Index(content[i:], "]")
			if endBracket == -1 {
				continue
			}

			text := content[i+1 : i+endBracket]

			// Check for (...)
			if i+endBracket+1 < len(content) && content[i+endBracket+1] == '(' {
				endParen := strings.Index(content[i+endBracket+2:], ")")
				if endParen == -1 {
					continue
				}

				url := content[i+endBracket+2 : i+endBracket+2+endParen]

				// Convert to Feishu format
				result.WriteString(content[start:i])
				result.WriteString(fmt.Sprintf("<%s|%s>", url, text))

				i += endBracket + 2 + endParen
				start = i + 1
			}
		}
	}

	if start < len(content) {
		result.WriteString(content[start:])
	}

	return result.String()
}

// FormatCard creates an interactive card from content.
func (f *Formatter) FormatCard(title, content string) *Card {
	if !f.cards {
		return nil
	}

	card := &Card{
		Elements: make([]*CardElement, 0),
	}

	// Add header if title provided
	if title != "" {
		card.Header = &CardHeader{
			Title: &CardText{
				Tag:     "plain_text",
				Content: title,
			},
			Template: "blue",
		}
	}

	// Add content as text element
	content = f.FormatText(content)

	card.Elements = append(card.Elements, &CardElement{
		Tag: "div",
		Text: &CardText{
			Tag:     "lark_md",
			Content: content,
		},
	})

	return card
}

// FormatCardWithActions creates a card with action buttons.
func (f *Formatter) FormatCardWithActions(title, content string, actions []*CardActionElement) *Card {
	card := f.FormatCard(title, content)
	if card == nil {
		return nil
	}

	// Add action element
	card.Elements = append(card.Elements, &CardElement{
		Tag:     "action",
		Actions: actions,
	})

	return card
}

// FormatAgentStream formats streaming agent events.
func (f *Formatter) FormatAgentStream(event string, data string) (any, interfaces.MessageFormat) {
	switch event {
	case "text", "message":
		if f.markdown {
			return f.FormatMarkdown(data), interfaces.FormatMarkdown
		}
		return f.FormatText(data), interfaces.FormatText

	case "error":
		// Format as card for errors
		return f.FormatCard("Error", data), interfaces.FormatCard

	case "tool_use":
		// Format as markdown for tool calls
		return fmt.Sprintf("**Tool:** %s", data), interfaces.FormatMarkdown

	case "tool_result":
		// Format as markdown for tool results
		return fmt.Sprintf("**Tool Result:**\n%s", f.FormatText(data)), interfaces.FormatMarkdown

	default:
		return f.FormatText(data), interfaces.FormatText
	}
}

// FormatCodeBlock formats code blocks for Feishu.
func (f *Formatter) FormatCodeBlock(language, code string) string {
	// Feishu markdown supports code blocks
	return fmt.Sprintf("```%s\n%s\n```", language, code)
}

// FormatTable formats a simple table as markdown.
func (f *Formatter) FormatTable(headers []string, rows [][]string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("| ")
	for _, h := range headers {
		sb.WriteString(h)
		sb.WriteString(" | ")
	}
	sb.WriteString("\n")

	// Separator
	sb.WriteString("| ")
	for range headers {
		sb.WriteString("--- | ")
	}
	sb.WriteString("\n")

	// Rows
	for _, row := range rows {
		sb.WriteString("| ")
		for _, cell := range row {
			sb.WriteString(cell)
			sb.WriteString(" | ")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatList formats a list as markdown.
func (f *Formatter) FormatList(items []string, ordered bool) string {
	if ordered {
		return f.formatNumberedList(items)
	}
	return f.formatUnorderedList(items)
}

// formatUnorderedList formats an unordered list.
func (f *Formatter) formatUnorderedList(items []string) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- %s\n", item))
	}
	return sb.String()
}

// formatNumberedList formats a numbered list.
func (f *Formatter) formatNumberedList(items []string) string {
	var sb strings.Builder
	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}
	return sb.String()
}

// FormatJSON formats JSON data as a code block.
func (f *Formatter) FormatJSON(data any) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("```json\nError formatting JSON: %v\n```", err)
	}
	return fmt.Sprintf("```json\n%s\n```", string(jsonBytes))
}

// CreateSimpleButton creates a simple button for cards.
func (f *Formatter) CreateSimpleButton(text, value string) *CardActionElement {
	return &CardActionElement{
		Tag: "button",
		Text: CardText{
			Tag:     "plain_text",
			Content: text,
		},
		Type: "default",
		Value: map[string]any{
			"action": value,
		},
	}
}

// CreatePrimaryButton creates a primary (highlighted) button.
func (f *Formatter) CreatePrimaryButton(text, value string) *CardActionElement {
	return &CardActionElement{
		Tag: "button",
		Text: CardText{
			Tag:     "plain_text",
			Content: text,
		},
		Type: "primary",
		Value: map[string]any{
			"action": value,
		},
	}
}

// FormatMessage formats a complete message for sending to Feishu.
func (f *Formatter) FormatMessage(content string, format interfaces.MessageFormat) (msgType string, msgContent string, err error) {
	switch format {
	case interfaces.FormatText:
		contentJSON, _ := json.Marshal(TextMessageContent{Text: content})
		return "text", string(contentJSON), nil

	case interfaces.FormatMarkdown:
		// Use post type for markdown
		contentMap := map[string]string{"text": content}
		sections := []map[string]string{contentMap}
		postContent := map[string][]map[string]string{"zh_cn": sections}
		postJSON, _ := json.Marshal(postContent)
		return "post", fmt.Sprintf(`{"post":%s}`, string(postJSON)), nil

	case interfaces.FormatCard:
		card := f.FormatCard("", content)
		if card == nil {
			return f.FormatMessage(content, interfaces.FormatText)
		}
		cardJSON, _ := json.Marshal(card)
		return "interactive", string(cardJSON), nil

	default:
		return f.FormatMessage(content, interfaces.FormatText)
	}
}

// ShouldUseCard determines if content should be formatted as a card.
func (f *Formatter) ShouldUseCard(content string) bool {
	if !f.cards {
		return false
	}

	// Use cards for structured content
	indicators := []string{
		"Error:", "Warning:", "Success:", "Info:",
		"```", "```json", "```go", "```bash",
		"| ", " | ", " |", // Tables
	}

	contentLower := strings.ToLower(content)
	for _, indicator := range indicators {
		if strings.Contains(contentLower, strings.ToLower(indicator)) {
			return true
		}
	}

	return false
}