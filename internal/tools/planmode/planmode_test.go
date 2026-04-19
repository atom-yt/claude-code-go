package planmode

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnterPlanModeTool_Name(t *testing.T) {
	tool := &EnterPlanModeTool{}
	assert.Equal(t, "EnterPlanMode", tool.Name())
}

func TestEnterPlanModeTool_ReadOnly(t *testing.T) {
	tool := &EnterPlanModeTool{}
	assert.True(t, tool.IsReadOnly())
}

func TestEnterPlanModeTool_ConcurrencySafe(t *testing.T) {
	tool := &EnterPlanModeTool{}
	assert.True(t, tool.IsConcurrencySafe())
}

func TestEnterPlanModeTool_Call(t *testing.T) {
	tool := &EnterPlanModeTool{}
	result, err := tool.Call(context.Background(), map[string]any{})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Plan mode entered")
}

func TestExitPlanModeTool_Name(t *testing.T) {
	tool := &ExitPlanModeTool{}
	assert.Equal(t, "ExitPlanMode", tool.Name())
}

func TestExitPlanModeTool_ReadOnly(t *testing.T) {
	tool := &ExitPlanModeTool{}
	assert.True(t, tool.IsReadOnly())
}

func TestExitPlanModeTool_ConcurrencySafe(t *testing.T) {
	tool := &ExitPlanModeTool{}
	assert.True(t, tool.IsConcurrencySafe())
}

func TestExitPlanModeTool_Call(t *testing.T) {
	tool := &ExitPlanModeTool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"allowedPrompts": []map[string]any{
			{
				"prompt": "run tests",
				"tool":   "Bash",
			},
		},
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Plan mode exited")
}

func TestExitPlanModeTool_CallWithoutPrompts(t *testing.T) {
	tool := &ExitPlanModeTool{}
	result, err := tool.Call(context.Background(), map[string]any{})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Plan mode exited")
}
