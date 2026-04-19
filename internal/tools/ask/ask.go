// Package ask implements the AskUserQuestion tool for interactive user decision support.
// This tool allows the agent to ask the user questions and get their input during execution.
package ask

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

// Option represents a question option.
type Option struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	Preview     string `json:"preview,omitempty"` // Optional preview content
}

// Question represents a question to ask the user.
type Question struct {
	Question    string   `json:"question"`
	Header      string   `json:"header,omitempty"`      // Short label (max 12 chars)
	Options     []Option `json:"options,omitempty"`     // Options (if multi-select is false)
	MultiSelect bool     `json:"multiSelect,omitempty"` // Allow multiple selections
}

// Answer represents the user's response.
type Answer struct {
	Question string   `json:"question"`
	Selected []string `json:"selected"` // Selected option labels
	Other    string   `json:"other"`    // Custom text (if user selects "Other")
}

// Tool implements the AskUserQuestion tool.
type Tool struct {
	// In a full implementation, this would have a callback to TUI
	// to actually interact with the user
}

var _ tools.Tool = (*Tool)(nil)

func (t *Tool) Name() string            { return "AskUserQuestion" }
func (t *Tool) IsReadOnly() bool        { return true }
func (t *Tool) IsConcurrencySafe() bool { return false } // Requires user interaction

func (t *Tool) Description() string {
	return "Use this tool when you need to ask the user questions during execution. " +
		"This allows you to: gather user preferences or requirements, clarify ambiguous instructions, " +
		"get decisions on implementation choices, offer choices between different approaches. " +
		"Users can always select 'Other' to provide custom text input."
}

func (t *Tool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"questions": map[string]any{
				"type":     "array",
				"minItems": 1,
				"maxItems": 4,
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"question": map[string]any{
							"type":        "string",
							"description": "The question to ask the user (should end with a question mark)",
						},
						"header": map[string]any{
							"type":        "string",
							"maxLength":   12,
							"description": "Short label displayed as a chip/tag (max 12 chars)",
						},
						"options": map[string]any{
							"type":     "array",
							"minItems": 2,
							"maxItems": 4,
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"label": map[string]any{
										"type":        "string",
										"description": "The display text for the option (1-5 words, should be concise)",
									},
									"description": map[string]any{
										"type":        "string",
										"description": "Explanation of what this option means",
									},
									"preview": map[string]any{
										"type":        "string",
										"description": "Optional preview content (ASCII mockups, code snippets, diagrams)",
									},
								},
								"required": []string{"label", "description"},
							},
						},
						"multiSelect": map[string]any{
							"type":        "boolean",
							"description": "Whether multiple options can be selected",
						},
					},
					"required": []string{"question"},
				},
			},
			"answers": map[string]any{
				"type":        "object",
				"description": "Pre-populated answers from previous user input (for resume)",
			},
		},
		"required": []string{"questions"},
	}
}

func (t *Tool) Call(_ context.Context, input map[string]any) (tools.ToolResult, error) {
	questionsRaw, ok := input["questions"]
	if !ok {
		return tools.ToolResult{Output: "Error: questions parameter is required", IsError: true}, nil
	}

	questions, ok := questionsRaw.([]any)
	if !ok {
		return tools.ToolResult{Output: "Error: questions must be an array", IsError: true}, nil
	}

	if len(questions) == 0 || len(questions) > 4 {
		return tools.ToolResult{Output: "Error: questions must have 1-4 items", IsError: true}, nil
	}

	// Parse questions
	var parsedQuestions []Question
	for i, qRaw := range questions {
		qMap, ok := qRaw.(map[string]any)
		if !ok {
			return tools.ToolResult{Output: fmt.Sprintf("Error: question %d is not an object", i+1), IsError: true}, nil
		}

		questionText, _ := qMap["question"].(string)
		if questionText == "" {
			return tools.ToolResult{Output: fmt.Sprintf("Error: question %d is missing 'question' field", i+1), IsError: true}, nil
		}

		header, _ := qMap["header"].(string)
		multiSelect, _ := qMap["multiSelect"].(bool)

		var options []Option
		if optsRaw, ok := qMap["options"].([]any); ok {
			if len(optsRaw) < 2 || len(optsRaw) > 4 {
				return tools.ToolResult{Output: fmt.Sprintf("Error: question %d options must have 2-4 items", i+1), IsError: true}, nil
			}

			for _, optRaw := range optsRaw {
				optMap, ok := optRaw.(map[string]any)
				if !ok {
					continue
				}

				label, _ := optMap["label"].(string)
				desc, _ := optMap["description"].(string)
				preview, _ := optMap["preview"].(string)

				if label == "" || desc == "" {
					continue
				}

				options = append(options, Option{
					Label:       label,
					Description: desc,
					Preview:     preview,
				})
			}
		}

		parsedQuestions = append(parsedQuestions, Question{
			Question:    questionText,
			Header:      header,
			Options:     options,
			MultiSelect: multiSelect,
		})
	}

	// In a full implementation, this would:
	// 1. Send the questions to the TUI layer
	// 2. Wait for user input
	// 3. Return the user's answers

	// For now, return a formatted question display
	return tools.ToolResult{
		Output: formatQuestions(parsedQuestions) + "\n\n[Interactive question mode - user response needed]",
	}, nil
}

// formatQuestions formats questions for display
func formatQuestions(questions []Question) string {
	var sb strings.Builder

	for i, q := range questions {
		if i > 0 {
			sb.WriteString("\n\n")
		}

		sb.WriteString(fmt.Sprintf("Q%d: %s\n", i+1, q.Question))

		if len(q.Options) > 0 {
			for j, opt := range q.Options {
				sb.WriteString(fmt.Sprintf("  [%d] %s\n", j+1, opt.Label))
				sb.WriteString(fmt.Sprintf("      %s\n", opt.Description))
				if opt.Preview != "" {
					sb.WriteString(fmt.Sprintf("      Preview:\n%s\n", opt.Preview))
				}
			}
			if !q.MultiSelect {
				sb.WriteString("  [O] Other (custom input)")
			}
		}
	}

	return sb.String()
}

// MarshalJSON implements json.Marshaler for Question
func (q *Question) MarshalJSON() ([]byte, error) {
	type Alias Question
	return json.Marshal((*Alias)(q))
}

// MarshalJSON implements json.Marshaler for Option
func (o *Option) MarshalJSON() ([]byte, error) {
	type Alias Option
	return json.Marshal((*Alias)(o))
}
