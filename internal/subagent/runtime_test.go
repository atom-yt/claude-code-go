package subagent

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuntime(t *testing.T) {
	runtime := NewRuntime()
	assert.NotNil(t, runtime)
	assert.Equal(t, 0, runtime.GetSubagentCount())
}

func TestRuntime_Spawn(t *testing.T) {
	runtime := NewRuntime()
	ctx := context.Background()

	// Create a simple executor
	execFn := func(ctx context.Context) Result {
		return Result{
			Type:    "stream",
			Content: "test output",
		}
	}

	subagent, err := runtime.Spawn(ctx, "task-1", "test prompt", WorkerTypeExplorer, execFn)
	require.NoError(t, err)

	assert.Equal(t, "subagent-1", subagent.ID)
	assert.Equal(t, "task-1", subagent.TaskID)
	assert.Equal(t, WorkerTypeExplorer, subagent.Type)
}

func TestRuntime_SpawnWithOutput(t *testing.T) {
	runtime := NewRuntime()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	executed := false
	execFn := func(ctx context.Context) Result {
		executed = true
		return Result{
			Type:    "stream",
			Content: "execution complete",
		}
	}

	subagent, err := runtime.Spawn(ctx, "task-1", "test prompt", WorkerTypeExplorer, execFn)
	require.NoError(t, err)

	// Wait for subagent to complete by checking its status
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if subagent.Status == "completed" {
			break
		}
	}

	assert.True(t, executed)
	assert.Equal(t, "completed", subagent.Status)
}

func TestRuntime_Get(t *testing.T) {
	runtime := NewRuntime()
	ctx := context.Background()

	execFn := func(ctx context.Context) Result {
		return Result{Type: "stream", Content: "output"}
	}

	subagent, err := runtime.Spawn(ctx, "task-1", "test", WorkerTypeExplorer, execFn)
	require.NoError(t, err)

	// Get by ID
	found, ok := runtime.Get(subagent.ID)
	require.True(t, ok)
	assert.Equal(t, subagent.ID, found.ID)

	// Get non-existent
	_, ok = runtime.Get("non-existent")
	assert.False(t, ok)
}

func TestRuntime_List(t *testing.T) {
	runtime := NewRuntime()
	ctx := context.Background()

	execFn := func(ctx context.Context) Result {
		return Result{Type: "stream", Content: "output"}
	}

	// Spawn multiple subagents
	s1, _ := runtime.Spawn(ctx, "task-1", "test1", WorkerTypeExplorer, execFn)
	_, _ = runtime.Spawn(ctx, "task-2", "test2", WorkerTypeWriter, execFn)
	s3, _ := runtime.Spawn(ctx, "task-1", "test3", WorkerTypeExplorer, execFn)

	// List all
	all := runtime.List("")
	assert.Len(t, all, 3)

	// List by task ID
	task1Subagents := runtime.List("task-1")
	assert.Len(t, task1Subagents, 2)
	ids := []string{task1Subagents[0].ID, task1Subagents[1].ID}
	assert.Contains(t, ids, s1.ID)
	assert.Contains(t, ids, s3.ID)
}

func TestRuntime_Stop(t *testing.T) {
	runtime := NewRuntime()
	ctx := context.Background()

	// Create a long-running executor
	longRunning := make(chan bool)
	execFn := func(ctx context.Context) Result {
		<-ctx.Done() // Wait for cancellation
		longRunning <- true
		return Result{Type: "stream", Content: "cancelled"}
	}

	subagent, err := runtime.Spawn(ctx, "task-1", "test", WorkerTypeExplorer, execFn)
	require.NoError(t, err)

	// Wait for goroutine to start
	time.Sleep(20 * time.Millisecond)

	// Stop it
	err = runtime.Stop(subagent.ID)
	require.NoError(t, err)

	// Wait for cancellation
	<-longRunning

	// Check status
	found, ok := runtime.Get(subagent.ID)
	require.True(t, ok)
	assert.Equal(t, "stopped", found.Status)

	// Try to stop again - should fail
	err = runtime.Stop(subagent.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already stopped")
}

func TestRuntime_StopNonExistent(t *testing.T) {
	runtime := NewRuntime()

	err := runtime.Stop("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestIsToolAllowed(t *testing.T) {
	// Explorer tools
	assert.True(t, IsToolAllowed("Read", WorkerTypeExplorer))
	assert.True(t, IsToolAllowed("Glob", WorkerTypeExplorer))
	assert.True(t, IsToolAllowed("Grep", WorkerTypeExplorer))
	assert.False(t, IsToolAllowed("Write", WorkerTypeExplorer))
	assert.False(t, IsToolAllowed("Bash", WorkerTypeExplorer))

	// Writer tools
	assert.True(t, IsToolAllowed("Read", WorkerTypeWriter))
	assert.True(t, IsToolAllowed("Write", WorkerTypeWriter))
	assert.True(t, IsToolAllowed("Edit", WorkerTypeWriter))
	assert.False(t, IsToolAllowed("Bash", WorkerTypeWriter))
	assert.True(t, IsToolAllowed("WebFetch", WorkerTypeWriter))

	// Unknown worker type
	assert.False(t, IsToolAllowed("Read", WorkerType("unknown")))
}

func TestRuntime_Cleanup(t *testing.T) {
	runtime := NewRuntime()
	ctx := context.Background()

	execFn := func(ctx context.Context) Result {
		return Result{Type: "stream", Content: "output"}
	}

	// Spawn some subagents
	s1, _ := runtime.Spawn(ctx, "task-1", "test1", WorkerTypeExplorer, execFn)
	s2, _ := runtime.Spawn(ctx, "task-2", "test2", WorkerTypeWriter, execFn)

	// Stop one
	runtime.Stop(s1.ID)

	// Wait for both to have ended
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if s1.Status == "stopped" {
			break
		}
	}

	// Cleanup subagents older than 0 time (all of them)
	count := runtime.Cleanup(0)
	assert.Greater(t, count, 0)

	// Both subagents should be removed
	_, ok := runtime.Get(s1.ID)
	assert.False(t, ok)

	_, ok = runtime.Get(s2.ID)
	assert.False(t, ok)
}

func TestRuntime_ConcurrentSpawn(t *testing.T) {
	runtime := NewRuntime()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			execFn := func(ctx context.Context) Result {
				return Result{Type: "stream", Content: fmt.Sprintf("task %d", i)}
			}
			runtime.Spawn(ctx, "task-1", fmt.Sprintf("test%d", i), WorkerTypeExplorer, execFn)
		}(i)
	}

	wg.Wait()
	// Small delay to let goroutines set their status
	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, 10, runtime.GetSubagentCount())
}

func TestSubagent_Status(t *testing.T) {
	runtime := NewRuntime()
	ctx := context.Background()

	execFn := func(ctx context.Context) Result {
		return Result{Type: "stream", Content: "output"}
	}

	subagent, _ := runtime.Spawn(ctx, "task-1", "test", WorkerTypeExplorer, execFn)

	// Wait for goroutine to start
	time.Sleep(20 * time.Millisecond)

	// Status should be running after sleep
	assert.Equal(t, "completed", subagent.Status)

	// Wait for completion
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if subagent.Status == "completed" {
			break
		}
	}

	// Status should be completed after result
	assert.Equal(t, "completed", subagent.Status)
	assert.False(t, subagent.StartedAt.IsZero())
}