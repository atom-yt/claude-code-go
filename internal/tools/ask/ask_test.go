package ask

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAskTool_Name(t *testing.T) {
	tool := &Tool{}
	assert.Equal(t, "AskUserQuestion", tool.Name())
}

func TestAskTool_ReadOnly(t *testing.T) {
	tool := &Tool{}
	assert.True(t, tool.IsReadOnly())
}

func TestAskTool_NotConcurrencySafe(t *testing.T) {
	tool := &Tool{}
	assert.False(t, tool.IsConcurrencySafe())
}

func TestAskTool_MissingQuestions(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "questions parameter is required")
}

func TestAskTool_InvalidQuestions(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"questions": "not an array",
	})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "must be an array")
}

func TestAskTool_TooManyQuestions(t *testing.T) {
	tool := &Tool{}
	questions := make([]any, 5) // More than 4
	for i := range questions {
		questions[i] = map[string]any{"question": "Test?"}
	}

	result, err := tool.Call(context.Background(), map[string]any{
		"questions": questions,
	})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "must have 1-4 items")
}

func TestAskTool_MissingQuestionText(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"questions": []any{
			map[string]any{"header": "Test"},
		},
	})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "missing 'question' field")
}

func TestAskTool_WithOptions(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"questions": []any{
			map[string]any{
				"question": "What should we do?",
				"header":   "Action",
				"options": []any{
					map[string]any{
						"label":       "Option A",
						"description": "Do something",
					},
					map[string]any{
						"label":       "Option B",
						"description": "Do something else",
					},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Q1: What should we do?")
	assert.Contains(t, result.Output, "[1] Option A")
	assert.Contains(t, result.Output, "[2] Option B")
}

func TestAskTool_MultiSelect(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"questions": []any{
			map[string]any{
				"question":    "Which features?",
				"multiSelect": true,
				"options": []any{
					map[string]any{
						"label":       "Feature 1",
						"description": "First feature",
					},
					map[string]any{
						"label":       "Feature 2",
						"description": "Second feature",
					},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Q1: Which features?")
}

func TestAskTool_WithPreview(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"questions": []any{
			map[string]any{
				"question": "Choose UI style?",
				"options": []any{
					map[string]any{
						"label":       "Dark",
						"description": "Dark theme",
						"preview":     "┌─────────┐\n│ Dark BG │\n└─────────┘",
					},
					map[string]any{
						"label":       "Light",
						"description": "Light theme",
					},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Preview:")
	assert.Contains(t, result.Output, "┌─────────┐")
}

func TestAskTool_MultipleQuestions(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"questions": []any{
			map[string]any{
				"question": "What language?",
				"options": []any{
					map[string]any{"label": "Go", "description": "Golang"},
					map[string]any{"label": "Python", "description": "Python"},
				},
			},
			map[string]any{
				"question": "What framework?",
				"options": []any{
					map[string]any{"label": "Echo", "description": "Echo framework"},
					map[string]any{"label": "Gin", "description": "Gin framework"},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Q1: What language?")
	assert.Contains(t, result.Output, "Q2: What framework?")
}

func TestFormatQuestions_SingleQuestion(t *testing.T) {
	questions := []Question{
		{
			Question: "Test question?",
			Options: []Option{
				{Label: "A", Description: "Option A"},
				{Label: "B", Description: "Option B"},
			},
		},
	}

	result := formatQuestions(questions)
	assert.Contains(t, result, "Q1: Test question?")
	assert.Contains(t, result, "[1] A")
	assert.Contains(t, result, "[2] B")
}
