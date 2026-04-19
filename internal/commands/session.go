package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/session"
)

// ---- /session ----

type sessionCmd struct{}

func (c *sessionCmd) Name() string      { return "session" }
func (c *sessionCmd) Aliases() []string { return nil }
func (c *sessionCmd) Description() string {
	return "List saved sessions"
}

func (c *sessionCmd) Execute(_ context.Context, args []string, cmdCtx *Context) (string, error) {
	sessions, err := session.List()
	if err != nil {
		return fmt.Sprintf("Failed to list sessions: %v", err), nil
	}

	if len(sessions) == 0 {
		return "No saved sessions found.", nil
	}

	lines := []string{fmt.Sprintf("Saved sessions (%d):", len(sessions))}
	lines = append(lines, "")

	for i, s := range sessions {
		lines = append(lines, c.formatSession(i+1, s))
	}

	lines = append(lines, "")
	lines = append(lines, "Note: Session restoration requires restarting the CLI with the --resume flag.")

	return "<!-- raw -->\n" + strings.Join(lines, "\n"), nil
}

func (c *sessionCmd) formatSession(index int, s session.Record) string {
	created := s.CreatedAt.Format("2006-01-02 15:04")
	updated := s.UpdatedAt.Format("2006-01-02 15:04")
	duration := time.Since(s.CreatedAt)
	durationStr := ""
	if duration < 24*time.Hour {
		durationStr = fmt.Sprintf(" (%s ago", formatDuration(duration))
	} else {
		durationStr = fmt.Sprintf(" (%s ago", s.CreatedAt.Format("Jan 02"))
	}

	msgCount := len(s.Messages)
	totalTokens := s.InputTokens + s.OutputTokens

	lines := []string{
		fmt.Sprintf("%d. %s", index, s.ID),
		fmt.Sprintf("   Model:        %s", s.Model),
		fmt.Sprintf("   Created:      %s%s", created, durationStr),
		fmt.Sprintf("   Last updated: %s", updated),
		fmt.Sprintf("   Messages:     %d", msgCount),
		fmt.Sprintf("   Tokens:       %d total (%d input, %d output)", totalTokens, s.InputTokens, s.OutputTokens),
	}

	// Show first message as preview
	if len(s.Messages) > 0 {
		preview := getMessageText(s.Messages[0])
		if len(preview) > 60 {
			preview = preview[:60] + "..."
		}
		lines = append(lines, fmt.Sprintf("   Preview:      %s", preview))
	}

	return strings.Join(lines, "\n")
}

// getMessageText extracts text content from a message.
func getMessageText(msg api.Message) string {
	for _, block := range msg.Content {
		if block.Type == "text" {
			return block.Text
		}
	}
		return ""
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}