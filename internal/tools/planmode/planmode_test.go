package planmode

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/atom-yt/claude-code-go/internal/runtime"
	"github.com/stretchr/testify/assert"
)

func setupTestState(t *testing.T) *runtime.State {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	return runtime.NewRuntimeState(tmpDir)
}

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
	state := setupTestState(t)
	tool := &EnterPlanModeTool{State: state}
	result, err := tool.Call(context.Background(), map[string]any{
		"title": "Test Plan",
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Plan mode entered")
	// Verify plan file was created
	assert.FileExists(t, filepath.Join(state.PlanFilePath()))
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
	state := setupTestState(t)
	// First enter plan mode
	enterTool := &EnterPlanModeTool{State: state}
	enterResult, err := enterTool.Call(context.Background(), map[string]any{
		"title": "Test Plan",
	})
	assert.NoError(t, err)
	assert.False(t, enterResult.IsError)

	// Create a plan file
	planContent := "# Test Plan\n\n- Step 1\n- Step 2\n"
	assert.NoError(t, os.WriteFile(state.PlanFilePath(), []byte(planContent), 0644))

	// Now exit plan mode
	exitTool := &ExitPlanModeTool{State: state}
	result, err := exitTool.Call(context.Background(), map[string]any{})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Plan approved")
}
