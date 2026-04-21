// Package subagent implements local subagent runtime for background task execution.
// Supports goroutine workers with read-only and limited-write capabilities.
package subagent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkerType defines the capabilities of a subagent worker.
type WorkerType string

const (
	// WorkerTypeExplorer is a read-only worker that can explore the codebase.
	WorkerTypeExplorer WorkerType = "explorer"
	// WorkerTypeWriter is a worker with limited write capabilities.
	WorkerTypeWriter WorkerType = "writer"
)

// AllowedTools maps worker types to their permitted tool categories.
var AllowedTools = map[WorkerType][]string{
	WorkerTypeExplorer: {
		"Read", "Glob", "Grep", "WebFetch", "WebSearch",
		"TaskCreate", "TaskGet", "TaskList", "TaskUpdate", "TaskDelete",
		"AskUserQuestion", "EnterPlanMode", "ExitPlanMode",
	},
	WorkerTypeWriter: {
		"Read", "Glob", "Grep", "WebFetch", "WebSearch",
		"TaskCreate", "TaskGet", "TaskList", "TaskUpdate", "TaskDelete",
		"AskUserQuestion", "EnterPlanMode", "ExitPlanMode",
		"Write", "Edit",
	},
}

// Subagent represents a running background subagent.
type Subagent struct {
	ID        string
	TaskID    string
	Type      WorkerType
	Status    string
	CreatedAt time.Time
	StartedAt time.Time
	EndedAt   time.Time // When stopped or completed

	// Execution context
	ctx       context.Context
	cancel    context.CancelFunc

	// Output channels
	outputCh  chan<- Result
	resultCh  chan<- FinalResult

	// State for tracking
	mu sync.RWMutex
}

// Result represents a partial result from a subagent.
type Result struct {
	Type      string // "stream", "error", "progress"
	Content   string
	Error     error
	Timestamp time.Time
}

// FinalResult represents the final outcome of a subagent execution.
type FinalResult struct {
	Success    bool
	Output     string
	Error       error
	CompletedAt time.Time
}

// Runtime manages running subagents.
type Runtime struct {
	mu        sync.RWMutex
	subagents map[string]*Subagent
	nextID    int
}

// NewRuntime creates a new subagent runtime.
func NewRuntime() *Runtime {
	return &Runtime{
		subagents: make(map[string]*Subagent),
		nextID:    1,
	}
}

// Spawn creates and starts a new subagent worker.
func (r *Runtime) Spawn(ctx context.Context, taskID, prompt string, workerType WorkerType, execFn func(context.Context) Result) (*Subagent, error) {
	r.mu.Lock()
	id := fmt.Sprintf("subagent-%d", r.nextID)
	r.nextID++
	r.mu.Unlock()

	subagent := &Subagent{
		ID:       id,
		TaskID:   taskID,
		Type:     workerType,
		Status:   "starting",
		CreatedAt: time.Now(),
		outputCh:  make(chan<- Result, 10),
		resultCh:  make(chan<- FinalResult, 1),
	}

	workerCtx, cancel := context.WithCancel(ctx)
	subagent.ctx = workerCtx
	subagent.cancel = cancel

	// Register the subagent
	r.mu.Lock()
	r.subagents[id] = subagent
	r.mu.Unlock()

	// Start the worker in a goroutine
	go func() {
		defer close(subagent.outputCh)
		defer close(subagent.resultCh)

		// Check if context was already cancelled (e.g., by Stop() before we started)
		select {
		case <-workerCtx.Done():
			// Context cancelled before we could start
			// Stop() will have set status to "stopped"
			return
		default:
		}

		subagent.mu.Lock()
		subagent.Status = "running"
		subagent.StartedAt = time.Now()
		subagent.mu.Unlock()

		// Check again before executing
		select {
		case <-workerCtx.Done():
			// Context cancelled just as we were starting
			// Stop() will have set status to "stopped"
			return
		default:
		}

		// Execute the worker function
		result := execFn(workerCtx)

		// Check if context was cancelled (subagent was stopped)
		select {
		case <-workerCtx.Done():
			// Subagent was stopped by user, don't override status
			// The Stop() method already set status to "stopped"
			return
		default:
			// Normal completion, set final status
			now := time.Now()
			subagent.mu.Lock()
			subagent.EndedAt = now
			if result.Error == nil {
				subagent.Status = "completed"
			} else {
				subagent.Status = "failed"
			}
			subagent.mu.Unlock()

			// Send result
			subagent.resultCh <- FinalResult{
				Success:    result.Error == nil,
				Output:     result.Content,
				Error:       result.Error,
				CompletedAt: now,
			}
		}
	}()

	return subagent, nil
}

// Get retrieves a subagent by ID.
func (r *Runtime) Get(id string) (*Subagent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	subagent, ok := r.subagents[id]
	if !ok {
		return nil, false
	}
	return subagent, true
}

// List returns all running subagents, optionally filtered by task ID.
func (r *Runtime) List(taskID string) []*Subagent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Subagent
	for _, sub := range r.subagents {
		if taskID == "" || sub.TaskID == taskID {
			result = append(result, sub)
		}
	}
	return result
}

// Stop cancels and removes a subagent.
func (r *Runtime) Stop(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	subagent, ok := r.subagents[id]
	if !ok {
		return fmt.Errorf("subagent not found: %s", id)
	}

	// Use subagent's own mutex for status update
	subagent.mu.Lock()
	defer subagent.mu.Unlock()

	if subagent.Status == "stopped" || subagent.Status == "completed" {
		return fmt.Errorf("subagent already stopped: %s", id)
	}

	// Cancel the context
	if subagent.cancel != nil {
		subagent.cancel()
	}

	subagent.Status = "stopped"
	subagent.EndedAt = time.Now()

	// Don't remove from registry - keep it for querying status
	// Subagent can be cleaned up later via Cleanup()

	return nil
}

// IsToolAllowed checks if a tool is allowed for a given worker type.
func IsToolAllowed(toolName string, workerType WorkerType) bool {
	allowedTools, ok := AllowedTools[workerType]
	if !ok {
		return false
	}

	for _, allowed := range allowedTools {
		if allowed == toolName {
			return true
		}
	}
	return false
}

// GetSubagentCount returns the count of active subagents.
func (r *Runtime) GetSubagentCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, sub := range r.subagents {
		if sub.Status == "running" {
			count++
		}
	}
	return count
}

// Cleanup removes completed or stopped subagents older than the given duration.
func (r *Runtime) Cleanup(olderThan time.Duration) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0

	for id, sub := range r.subagents {
		if sub.Status == "stopped" || sub.Status == "completed" {
			if olderThan <= 0 {
				// Zero or negative duration: clean up all completed/stopped subagents
				delete(r.subagents, id)
				count++
			} else {
				// Positive duration: clean up only those older than cutoff
				cutoff := time.Now().Add(-olderThan)
				// Use StartedAt or EndedAt, whichever is earlier
				earliestTime := sub.StartedAt
				if !sub.EndedAt.IsZero() && sub.EndedAt.Before(sub.StartedAt) {
					earliestTime = sub.EndedAt
				}
				if earliestTime.Before(cutoff) {
					delete(r.subagents, id)
					count++
				}
			}
		}
	}

	return count
}
