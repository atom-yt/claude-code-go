package skills

import (
	"context"
	"fmt"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

// SkillTool implements the Skill tool for executing skills.
type SkillTool struct {
	registry *Registry
}

var _ tools.Tool = (*SkillTool)(nil)

// NewSkillTool creates a new Skill tool with the given registry.
func NewSkillTool(registry *Registry) *SkillTool {
	return &SkillTool{registry: registry}
}

func (t *SkillTool) Name() string            { return "Skill" }
func (t *SkillTool) IsReadOnly() bool        { return true }
func (t *SkillTool) IsConcurrencySafe() bool { return true }

func (t *SkillTool) Description() string {
	return "Execute a skill by name or trigger. Skills are user-defined capabilities that extend the agent's functionality. " +
		"Use this tool when the user references a slash command or skill name that matches a registered skill. " +
		"When the user types '/skillname', invoke this tool with skill: 'skillname' before doing anything else."
}

func (t *SkillTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"skill": map[string]any{
				"type":        "string",
				"description": "The skill name or trigger (without the leading /) to invoke",
			},
			"args": map[string]any{
				"type":        "string",
				"description": "Optional arguments to pass to the skill",
			},
		},
		"required": []string{"skill"},
	}
}

func (t *SkillTool) Call(ctx context.Context, input map[string]any) (tools.ToolResult, error) {
	skillName, _ := input["skill"].(string)
	if skillName == "" {
		return tools.ToolResult{Output: "Error: skill parameter is required", IsError: true}, nil
	}

	args, _ := input["args"].(string)

	// Try to find the skill by trigger first
	skill, ok := t.registry.GetByTrigger("/" + skillName)
	if !ok {
		// Try by name
		skill, ok = t.registry.GetByName(skillName)
		if !ok {
			return tools.ToolResult{
				Output:  fmt.Sprintf("Skill not found: %s\n\nAvailable skills: %s", skillName, formatAvailableSkills(t.registry)),
				IsError: true,
			}, nil
		}
	}

	// Return the skill content
	output := formatSkillOutput(skill, args)
	return tools.ToolResult{Output: output}, nil
}

// formatSkillOutput formats the skill output for the agent.
func formatSkillOutput(skill *Skill, args string) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("---\n"))
	output.WriteString(fmt.Sprintf("Skill: %s\n", skill.Name))
	output.WriteString(fmt.Sprintf("Trigger: %s\n", skill.Trigger))
	output.WriteString(fmt.Sprintf("Source: %s\n", skill.Source))
	if args != "" {
		output.WriteString(fmt.Sprintf("Arguments: %s\n", args))
	}
	output.WriteString(fmt.Sprintf("---\n\n"))
	output.WriteString(skill.Content)

	return output.String()
}

// formatAvailableSkills returns a comma-separated list of available skills.
func formatAvailableSkills(registry *Registry) string {
	skills := registry.List()
	if len(skills) == 0 {
		return "none"
	}

	var names []string
	for _, s := range skills {
		names = append(names, s.Trigger)
	}
	return strings.Join(names, ", ")
}
