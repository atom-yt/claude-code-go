package prompt

import (
	"fmt"
	"strings"
)

// ContextAssembly represents the full assembled context for an agent request.
// This structure captures all components that make up the final prompt context.
type ContextAssembly struct {
	// System prompt components
	BasePrompt      string
	ProjectSection  string
	SkillsSection   string
	MemorySection   string
	FullSystemPrompt string

	// Runtime context
	ToolsCount      int
	HistoryLength   int
	ProviderInfo    string
	ModelInfo       string

	// Metadata
	SessionID       string
}

// Format returns a human-readable representation of the assembled context.
// This is useful for debugging and verification.
func (c *ContextAssembly) Format() string {
	var sb strings.Builder

	sb.WriteString("=== Assembled Context ===\n\n")

	// System prompt breakdown
	sb.WriteString("## System Prompt\n")
	sb.WriteString(fmt.Sprintf("  Base prompt: %d chars\n", len(c.BasePrompt)))

	if c.ProjectSection != "" {
		sb.WriteString(fmt.Sprintf("  Project instructions: %d chars\n", len(c.ProjectSection)))
	}
	if c.SkillsSection != "" {
		sb.WriteString(fmt.Sprintf("  Skills section: %d chars\n", len(c.SkillsSection)))
	}
	if c.MemorySection != "" {
		sb.WriteString(fmt.Sprintf("  Memory section: %d chars\n", len(c.MemorySection)))
	}
	sb.WriteString(fmt.Sprintf("  Full system prompt: %d chars total\n", len(c.FullSystemPrompt)))

	// Runtime context
	sb.WriteString("\n## Runtime Context\n")
	sb.WriteString(fmt.Sprintf("  Tools available: %d\n", c.ToolsCount))
	sb.WriteString(fmt.Sprintf("  History messages: %d\n", c.HistoryLength))
	if c.ProviderInfo != "" {
		sb.WriteString(fmt.Sprintf("  Provider: %s\n", c.ProviderInfo))
	}
	if c.ModelInfo != "" {
		sb.WriteString(fmt.Sprintf("  Model: %s\n", c.ModelInfo))
	}

	// Metadata
	if c.SessionID != "" {
		sb.WriteString(fmt.Sprintf("  Session: %s\n", c.SessionID))
	}

	return sb.String()
}

// FormatVerbose returns a detailed representation including content previews.
func (c *ContextAssembly) FormatVerbose() string {
	var sb strings.Builder

	sb.WriteString("=== Assembled Context (Verbose) ===\n\n")

	// System prompt breakdown
	sb.WriteString("## System Prompt Components\n\n")

	sb.WriteString("### Base Prompt\n")
	sb.WriteString(truncate(c.BasePrompt, 500))
	sb.WriteString("\n\n")

	if c.ProjectSection != "" {
		sb.WriteString("### Project Instructions\n")
		sb.WriteString(truncate(c.ProjectSection, 500))
		sb.WriteString("\n\n")
	}

	if c.SkillsSection != "" {
		sb.WriteString("### Skills\n")
		sb.WriteString(truncate(c.SkillsSection, 500))
		sb.WriteString("\n\n")
	}

	if c.MemorySection != "" {
		sb.WriteString("### Memory\n")
		sb.WriteString(truncate(c.MemorySection, 500))
		sb.WriteString("\n\n")
	}

	// Runtime context
	sb.WriteString("## Runtime Context\n")
	sb.WriteString(fmt.Sprintf("  Tools: %d\n", c.ToolsCount))
	sb.WriteString(fmt.Sprintf("  History: %d messages\n", c.HistoryLength))
	if c.ProviderInfo != "" {
		sb.WriteString(fmt.Sprintf("  Provider: %s\n", c.ProviderInfo))
	}
	if c.ModelInfo != "" {
		sb.WriteString(fmt.Sprintf("  Model: %s\n", c.ModelInfo))
	}

	return sb.String()
}

// BuildContextAssembly creates a ContextAssembly from the prompt builder input.
func BuildContextAssembly(input SystemPromptInput, fullPrompt string) *ContextAssembly {
	assembly := &ContextAssembly{
		BasePrompt:       baseSystemPrompt,
		FullSystemPrompt: fullPrompt,
	}

	// Extract sections
	assembly.ProjectSection = FormatProjectInstructions(input.Project)
	assembly.SkillsSection = formatSkills(input.Skills)
	assembly.MemorySection = formatMemorySnippet(input.MemorySnippet)

	return assembly
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (truncated)"
}