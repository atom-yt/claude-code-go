// Package runtime manages agent execution state including plan mode and tasks.
package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Mode represents the current execution mode of the agent.
type Mode int

const (
	// ModeImplement is the default mode where the agent executes tasks.
	ModeImplement Mode = iota
	// ModePlan is a read-only exploration mode for planning.
	ModePlan
)

func (m Mode) String() string {
	switch m {
	case ModePlan:
		return "plan"
	default:
		return "implement"
	}
}

// Plan represents a plan artifact created during plan mode.
type Plan struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Steps       []PlanStep `json:"steps"`
	CreatedAt   time.Time `json:"createdAt"`
	Approved    bool      `json:"approved"`
	ApprovedAt  *time.Time `json:"approvedAt,omitempty"`
}

// PlanStep represents a single step in a plan.
type PlanStep struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"` // "pending", "in_progress", "completed", "blocked"
	Tool        string `json:"tool,omitempty"`
	Command     string `json:"command,omitempty"`
}

// State manages the runtime state of the agent.
type State struct {
	mu sync.RWMutex

	// Current mode
	mode Mode

	// Current plan (if in plan mode)
	currentPlan *Plan

	// Plan file path
	planFilePath string

	// Workspace root for plan storage
	workspaceRoot string
}

// NewState creates a new runtime state.
func NewState(workspaceRoot string) *State {
	s := &State{
		mode:          ModeImplement,
		workspaceRoot: workspaceRoot,
	}

	// Set plan file path
	if workspaceRoot != "" {
		s.planFilePath = filepath.Join(workspaceRoot, ".claude", "plan.md")
	}

	return s
}

// Mode returns the current execution mode.
func (s *State) Mode() Mode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode
}

// SetMode changes the execution mode.
func (s *State) SetMode(mode Mode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
}

// IsPlanMode returns true if the agent is in plan mode.
func (s *State) IsPlanMode() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode == ModePlan
}

// EnterPlanMode transitions to plan mode.
// Returns the plan file path for the agent to write to.
func (s *State) EnterPlanMode(ctx context.Context, title string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mode = ModePlan

	// Create new plan
	now := time.Now()
	s.currentPlan = &Plan{
		ID:          fmt.Sprintf("plan-%d", now.Unix()),
		Title:       title,
		CreatedAt:   now,
		Steps:       []PlanStep{},
	}

	// Ensure plan directory exists
	if s.planFilePath != "" {
		dir := filepath.Dir(s.planFilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create plan directory: %w", err)
		}

		// Write initial plan file
		initialContent := fmt.Sprintf("# Plan: %s\n\nCreated: %s\n\n## Steps\n\n(Plan steps will be added here)\n",
			title, now.Format(time.RFC3339))
		if err := os.WriteFile(s.planFilePath, []byte(initialContent), 0644); err != nil {
			return "", fmt.Errorf("failed to write plan file: %w", err)
		}

		return s.planFilePath, nil
	}

	return "", nil
}

// ExitPlanMode transitions out of plan mode and marks the plan for approval.
func (s *State) ExitPlanMode(ctx context.Context) (*Plan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentPlan == nil {
		return nil, fmt.Errorf("no active plan")
	}

	// Read plan file content
	if s.planFilePath != "" {
		content, err := os.ReadFile(s.planFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read plan file: %w", err)
		}

		// Parse steps from markdown
		s.currentPlan.Steps = parsePlanSteps(string(content))
		s.currentPlan.Description = string(content)
	}

	s.mode = ModeImplement
	return s.currentPlan, nil
}

// CurrentPlan returns the current plan.
func (s *State) CurrentPlan() *Plan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentPlan
}

// ApprovePlan marks the current plan as approved.
func (s *State) ApprovePlan() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentPlan == nil {
		return fmt.Errorf("no active plan")
	}

	now := time.Now()
	s.currentPlan.Approved = true
	s.currentPlan.ApprovedAt = &now

	return nil
}

// PlanFilePath returns the path to the plan file.
func (s *State) PlanFilePath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.planFilePath
}

// parsePlanSteps extracts plan steps from markdown content.
func parsePlanSteps(content string) []PlanStep {
	var steps []PlanStep
	lines := strings.Split(content, "\n")

	stepID := 1
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for list items that start with - or * or numbered
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			steps = append(steps, PlanStep{
				ID:          fmt.Sprintf("step-%d", stepID),
				Title:       strings.TrimPrefix(line, "- "),
				Description: strings.TrimPrefix(line, "- "),
				Status:      "pending",
			})
			stepID++
		} else if matched, _ := fmt.Sscanf(line, "%d. %s", new(int), new(string)); matched > 0 {
			// Handle numbered list items
			steps = append(steps, PlanStep{
				ID:          fmt.Sprintf("step-%d", stepID),
				Title:       strings.TrimSpace(line[strings.Index(line, ". ")+2:]),
				Description: strings.TrimSpace(line[strings.Index(line, ". ")+2:]),
				Status:      "pending",
			})
			stepID++
		}
	}

	return steps
}

// NewRuntimeState creates a new runtime state instance.
func NewRuntimeState(workspaceRoot string) *State {
	return NewState(workspaceRoot)
}
