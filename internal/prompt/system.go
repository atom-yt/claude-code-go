package prompt

import (
	"fmt"
	"strings"
)

const baseSystemPrompt = `You are Atom, an AI-powered terminal coding assistant.

You operate inside the user's current workspace and help with software engineering tasks.

Core expectations:
- Be accurate about the current codebase state.
- Prefer inspecting the codebase before making assumptions.
- Use available tools deliberately and explain concrete results.
- Respect project instructions and local development conventions.
- Be careful with file edits, shell commands, and potentially destructive actions.
- When context is limited, preserve important state explicitly instead of guessing.`

// SkillSummary is a compact description of one loaded skill.
type SkillSummary struct {
	Name        string
	Trigger     string
	Description string
	Source      string
}

// SystemPromptInput holds optional prompt builder inputs.
type SystemPromptInput struct {
	Project       ProjectContext
	Skills        []SkillSummary
	MemorySnippet string
}

// BuildSystemPrompt assembles the base runtime prompt plus discovered project
// instructions into one system prompt string.
func BuildSystemPrompt(input SystemPromptInput) string {
	parts := []string{baseSystemPrompt}

	if projectInstructions := FormatProjectInstructions(input.Project); projectInstructions != "" {
		parts = append(parts, projectInstructions)
	}
	if skillsSection := formatSkills(input.Skills); skillsSection != "" {
		parts = append(parts, skillsSection)
	}
	if memorySection := formatMemorySnippet(input.MemorySnippet); memorySection != "" {
		parts = append(parts, memorySection)
	}

	return strings.Join(parts, "\n\n")
}

func formatSkills(skills []SkillSummary) string {
	if len(skills) == 0 {
		return ""
	}

	lines := []string{"## Loaded Skills", "The following skills are available in this workspace."}
	for _, skill := range skills {
		line := fmt.Sprintf("- %s", skill.Name)
		if skill.Trigger != "" {
			line += fmt.Sprintf(" (%s)", skill.Trigger)
		}
		if skill.Description != "" {
			line += ": " + skill.Description
		}
		if skill.Source != "" {
			line += fmt.Sprintf(" [%s]", skill.Source)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func formatMemorySnippet(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	return strings.Join([]string{
		"## Persistent Memory",
		"Use this as durable project memory when it is relevant and consistent with the current codebase.",
		content,
	}, "\n\n")
}
