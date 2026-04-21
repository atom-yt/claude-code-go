package commands

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ---- /subagent ----

type subagentCmd struct{}

func (c *subagentCmd) Name() string      { return "subagent" }
func (c *subagentCmd) Aliases() []string { return []string{"agents"} }
func (c *subagentCmd) Description() string {
	return "Show background subagent status and details"
}

func (c *subagentCmd) Execute(_ context.Context, args []string, ctx *Context) (string, error) {
	runtime := ctx.GetSubagentRuntime()
	if runtime == nil {
		return "Subagent runtime not available", nil
	}

	var sb strings.Builder

	// Get all subagents
	agents := runtime.List("")

	if len(agents) == 0 {
		sb.WriteString("No background subagents running.\n")
		sb.WriteString("────────────────────────────────────────────\n")
		sb.WriteString("Subagents can be spawned by Agent tool for parallel execution.\n")
		return "<!-- raw -->\n" + sb.String(), nil
	}

	sb.WriteString(fmt.Sprintf("Background Subagents: %d running\n", len(agents)))
	sb.WriteString("────────────────────────────────────────────\n")

	// Show each subagent
	for _, agent := range agents {
		sb.WriteString(fmt.Sprintf("ID:      %s\n", agent.ID))
		sb.WriteString(fmt.Sprintf("Type:    %s\n", agent.Type))
		sb.WriteString(fmt.Sprintf("Status:   %s\n", agent.Status))
		sb.WriteString(fmt.Sprintf("Created:  %s\n", agent.CreatedAt.Format("2006-01-02 15:04:05")))
		if !agent.StartedAt.IsZero() {
			sb.WriteString(fmt.Sprintf("Started:  %s\n", agent.StartedAt.Format("2006-01-02 15:04:05")))
		}
		if !agent.EndedAt.IsZero() {
			duration := agent.EndedAt.Sub(agent.CreatedAt)
			sb.WriteString(fmt.Sprintf("Ended:    %s (duration: %v)\n", agent.EndedAt.Format("2006-01-02 15:04:05"), duration.Round(time.Second)))
		} else {
			duration := time.Since(agent.CreatedAt)
			sb.WriteString(fmt.Sprintf("Running:  %v\n", duration.Round(time.Second)))
		}
		if agent.TaskID != "" {
			sb.WriteString(fmt.Sprintf("Task:     #%s\n", agent.TaskID))
		}
		sb.WriteString("\n")
	}

	// Show summary
	runningCount := 0
	completedCount := 0
	stoppedCount := 0
	for _, agent := range agents {
		switch agent.Status {
		case "running":
			runningCount++
		case "stopped":
			stoppedCount++
		case "completed":
			completedCount++
		}
	}

	sb.WriteString("────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Summary: ● %d running  ✓ %d completed  ■ %d stopped\n",
		runningCount, completedCount, stoppedCount))

	return "<!-- raw -->\n" + sb.String(), nil
}