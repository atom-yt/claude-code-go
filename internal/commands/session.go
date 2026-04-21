package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/session"
)

// ---- /session ----

type sessionCmd struct{}

func (c *sessionCmd) Name() string      { return "session" }
func (c *sessionCmd) Aliases() []string { return []string{"sessions"} }
func (c *sessionCmd) Description() string {
	return "List, search, and restore saved sessions. Usage: /session [search|restore] [args]"
}

func (c *sessionCmd) Execute(_ context.Context, args []string, cmdCtx *Context) (string, error) {
	if len(args) == 0 {
		return c.listSessions(cmdCtx, "", "recent")
	}

	switch args[0] {
	case "search":
		query := ""
		if len(args) > 1 {
			query = strings.Join(args[1:], " ")
		}
		return c.listSessions(cmdCtx, query, "recent")
	case "restore":
		if len(args) < 2 {
			return "Usage: /session restore <session-id>", nil
		}
		sessionID := args[1]
		info, err := cmdCtx.RestoreSession(sessionID)
		if err != nil {
			return fmt.Sprintf("Failed to restore session: %v", err), nil
		}
		return fmt.Sprintf("Restored session: %s\n\n%s", sessionID, info), nil
	case "recent":
		return c.listSessions(cmdCtx, "", "recent")
	case "tokens":
		return c.listSessions(cmdCtx, "", "tokens")
	case "messages":
		return c.listSessions(cmdCtx, "", "messages")
	case "model":
		if len(args) < 2 {
			return c.listSessions(cmdCtx, "", "model")
		}
		model := args[1]
		return c.listSessions(cmdCtx, model, "recent")
	default:
		// Treat as search query
		query := strings.Join(args, " ")
		return c.listSessions(cmdCtx, query, "recent")
	}
}

func (c *sessionCmd) listSessions(cmdCtx *Context, query, sortBy string) (string, error) {
	sessions, err := session.List()
	if err != nil {
		return fmt.Sprintf("Failed to list sessions: %v", err), nil
	}

	if len(sessions) == 0 {
		return "No saved sessions found.", nil
	}

	// Build and query the session index
	index := session.NewIndex()
	_ = index.Build(sessions)

	// Apply search filter if query provided
	var filteredSessions []session.Record
	if query != "" {
		entries := index.Search(query)
		filteredSessions = make([]session.Record, 0, len(entries))
		for _, entry := range entries {
			rec, err := session.Load(entry.SessionID)
			if err == nil {
				filteredSessions = append(filteredSessions, rec)
			}
		}
	} else {
		filteredSessions = sessions
	}

	if len(filteredSessions) == 0 {
		return fmt.Sprintf("No sessions found matching query: %q", query), nil
	}

	// Sort sessions based on sortBy option
	switch sortBy {
	case "tokens":
		sort.Slice(filteredSessions, func(i, j int) bool {
			return filteredSessions[i].InputTokens+filteredSessions[i].OutputTokens >
				filteredSessions[j].InputTokens+filteredSessions[j].OutputTokens
		})
	case "messages":
		sort.Slice(filteredSessions, func(i, j int) bool {
			return len(filteredSessions[i].Messages) > len(filteredSessions[j].Messages)
		})
	case "model":
		sort.Slice(filteredSessions, func(i, j int) bool {
			return filteredSessions[i].Model < filteredSessions[j].Model
		})
	default:
		// Already sorted by recency (UpdatedAt descending)
	}

	// Format output
	lines := []string{
		fmt.Sprintf("Saved sessions: %d found", len(filteredSessions)),
	}

	if query != "" {
		lines[0] += fmt.Sprintf(" (query: %q)", query)
	}
	if sortBy != "" && sortBy != "recent" {
		lines[0] += fmt.Sprintf(" (sorted by: %s)", sortBy)
	}

	lines = append(lines, "")

	for i, s := range filteredSessions {
		lines = append(lines, c.formatSession(i+1, s))
	}

	lines = append(lines, "")
	lines = append(lines, "Commands:")
	lines = append(lines, "  /session [query]          - Search sessions")
	lines = append(lines, "  /session recent         - List recent sessions")
	lines = append(lines, "  /session tokens         - List by token usage")
	lines = append(lines, "  /session messages        - List by message count")
	lines = append(lines, "  /session model <name>    - Filter by model")
	lines = append(lines, "  /session restore <id>    - Restore a session")

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