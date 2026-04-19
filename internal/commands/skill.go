package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/skills"
)

// skillCmd implements the /skill command.
type skillCmd struct {
	registry *skills.Registry
}

// NewSkillCmd creates a new skill command.
func NewSkillCmd(registry *skills.Registry) *skillCmd {
	return &skillCmd{registry: registry}
}

func (c *skillCmd) Name() string        { return "skill" }
func (c *skillCmd) Aliases() []string   { return nil }
func (c *skillCmd) Description() string { return "List available skills or show skill details" }

func (c *skillCmd) Execute(ctx context.Context, args []string, cmdCtx *Context) (string, error) {
	if len(args) == 0 {
		return c.listSkills(), nil
	}

	// Show details for a specific skill
	skillName := args[0]
	if skill, ok := c.registry.GetByName(skillName); ok {
		return c.showSkillDetails(skill), nil
	}
	if skill, ok := c.registry.GetByTrigger(skillName); ok {
		return c.showSkillDetails(skill), nil
	}

	return fmt.Sprintf("Skill not found: %s\n\nAvailable skills: %s", skillName, formatSkillList(c.registry)), nil
}

// listSkills lists all available skills.
func (c *skillCmd) listSkills() string {
	skillList := c.registry.List()
	if len(skillList) == 0 {
		return "No skills available. Add skills to ~/.claude/skills/ or .claude/skills/"
	}

	var sb strings.Builder
	sb.WriteString("Available skills:\n\n")

	for _, skill := range skillList {
		sb.WriteString(fmt.Sprintf("  %s — %s (%s)\n", skill.Trigger, skill.Description, skill.Source))
	}

	sb.WriteString(fmt.Sprintf("\nTotal: %d skills", len(skillList)))
	return sb.String()
}

// showSkillDetails shows detailed information about a skill.
func (c *skillCmd) showSkillDetails(skill *skills.Skill) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Skill: %s\n", skill.Name))
	sb.WriteString(fmt.Sprintf("Trigger: %s\n", skill.Trigger))
	sb.WriteString(fmt.Sprintf("Description: %s\n", skill.Description))
	sb.WriteString(fmt.Sprintf("Source: %s\n", skill.Source))
	sb.WriteString(fmt.Sprintf("Path: %s\n", skill.Path))
	sb.WriteString("\nContent:\n")
	sb.WriteString("───────────────────────────────────────\n")
	sb.WriteString(skill.Content)

	return sb.String()
}

// formatSkillList returns a comma-separated list of skill triggers.
func formatSkillList(registry *skills.Registry) string {
	triggers := registry.GetTriggers()
	if len(triggers) == 0 {
		return "none"
	}
	return strings.Join(triggers, ", ")
}

// ScanSkills scans the default directories for skills.
func ScanSkills() *skills.Registry {
	registry := skills.NewRegistry()

	// Add global skills directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalSkillsDir := filepath.Join(homeDir, ".claude", "skills")
		registry.AddDir(globalSkillsDir)
	}

	// Add project skills directory
	projectSkillsDir := filepath.Join(".", ".claude", "skills")
	if _, err := os.Stat(projectSkillsDir); err == nil {
		registry.AddDir(projectSkillsDir)
	}

	// Scan all directories
	_ = registry.Scan()

	return registry
}

// GetDefaultSkillsPath returns the default skills directory path.
func GetDefaultSkillsPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".claude", "skills")
}
