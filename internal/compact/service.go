package compact

import (
	"context"
	"fmt"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/api"
)

const summaryPrompt = `Summarize the earlier conversation history for future continuation.

Requirements:
- Preserve the user's goals, constraints, and explicit preferences.
- Preserve important code changes, files discussed, and technical decisions.
- Preserve unresolved issues, follow-up work, and open questions.
- Mention important tool results only when they materially affect next steps.
- Be concise but specific. Prefer bullets.
- Do not invent facts that are not in the conversation.
- This summary will replace the earlier history, so keep enough detail for the agent to continue accurately.`

const summaryHeader = "## Conversation Summary"

// Service compacts conversation history using the current model.
type Service struct {
	client       api.Streamer
	model        string
	systemPrompt string
}

// Result is the output of one compaction operation.
type Result struct {
	Summary string
	History []api.Message
	Noop    bool
}

// NewService creates a compaction service.
func NewService(client api.Streamer, model, systemPrompt string) *Service {
	return &Service{
		client:       client,
		model:        model,
		systemPrompt: systemPrompt,
	}
}

// Compact summarizes older history and returns rewritten history containing a
// compact summary plus the most recent tail.
func (s *Service) Compact(ctx context.Context, history []api.Message, keepRecent int) (Result, error) {
	if s.client == nil {
		return Result{}, fmt.Errorf("compact service requires an API client")
	}
	if len(history) == 0 {
		return Result{History: nil, Noop: true}, nil
	}
	if keepRecent < 0 {
		keepRecent = 0
	}
	if len(history) <= keepRecent {
		copied := append([]api.Message(nil), history...)
		return Result{History: copied, Noop: true}, nil
	}

	split := len(history) - keepRecent
	older := append([]api.Message(nil), history[:split]...)
	recent := append([]api.Message(nil), history[split:]...)

	summary, err := s.generateSummary(ctx, older)
	if err != nil {
		return Result{}, err
	}
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return Result{}, fmt.Errorf("compact summary is empty")
	}

	summaryMessage := api.Message{
		Role: api.RoleAssistant,
		Content: []api.ContentBlock{{
			Type: "text",
			Text: formatSummary(summary),
		}},
	}

	newHistory := make([]api.Message, 0, 1+len(recent))
	newHistory = append(newHistory, summaryMessage)
	newHistory = append(newHistory, recent...)

	return Result{
		Summary: formatSummary(summary),
		History: newHistory,
	}, nil
}

func (s *Service) generateSummary(ctx context.Context, older []api.Message) (string, error) {
	req := api.MessagesRequest{
		Model:    s.model,
		Messages: append(append([]api.Message(nil), older...), api.TextMessage(api.RoleUser, summaryPrompt)),
	}
	if s.systemPrompt != "" {
		req.SetSystemWithCaching(buildCompactSystemPrompt(s.systemPrompt))
	}

	var out strings.Builder
	for event := range s.client.StreamMessages(ctx, req) {
		switch event.Type {
		case api.EventTextDelta:
			out.WriteString(event.Text)
		case api.EventError:
			return "", event.Error
		}
	}

	return out.String(), nil
}

func buildCompactSystemPrompt(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		return "You are compacting conversation history for later continuation."
	}
	return base + "\n\n## Compaction Mode\nYou are compressing earlier conversation history into a durable continuation summary."
}

func formatSummary(summary string) string {
	return summaryHeader + "\n\n" + strings.TrimSpace(summary)
}
