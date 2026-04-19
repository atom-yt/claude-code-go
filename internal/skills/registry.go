// Package skills implements the skill system for extensible capabilities.
// Skills are user-defined capabilities that can be triggered by specific commands.
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Skill represents a user-defined skill.
type Skill struct {
	Name        string
	Description string
	Trigger     string      // Command that triggers this skill (e.g., "/graphify")
	Content     string      // The skill content (after the YAML frontmatter)
	Path        string      // Path to the SKILL.md file
	Source      SkillSource // Where this skill came from
}

// SkillSource indicates where a skill was loaded from.
type SkillSource string

const (
	SourceGlobal  SkillSource = "global"  // From ~/.claude/skills/
	SourceProject SkillSource = "project" // From .claude/skills/
)

// Registry manages all loaded skills.
type Registry struct {
	skills   map[string]*Skill // key: trigger
	byName   map[string]*Skill // key: name
	byPath   map[string]*Skill // key: file path
	dirs     []string          // directories to scan
	triggers map[string]string // mapping from normalized trigger to skill name
}

// NewRegistry creates a new skill registry.
func NewRegistry() *Registry {
	return &Registry{
		skills:   make(map[string]*Skill),
		byName:   make(map[string]*Skill),
		byPath:   make(map[string]*Skill),
		triggers: make(map[string]string),
	}
}

// AddDir adds a directory to scan for skills.
func (r *Registry) AddDir(dir string) {
	r.dirs = append(r.dirs, dir)
}

// Scan scans all registered directories for skills.
func (r *Registry) Scan() error {
	for _, dir := range r.dirs {
		if err := r.scanDir(dir); err != nil {
			// Log but continue scanning other directories
			continue
		}
	}
	return nil
}

// scanDir scans a single directory for skill files.
func (r *Registry) scanDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(dir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil {
			skill, err := parseSkillFile(skillPath)
			if err != nil {
				continue
			}

			// Determine source
			if strings.HasPrefix(dir, filepath.Join(os.Getenv("HOME"), ".claude")) {
				skill.Source = SourceGlobal
			} else {
				skill.Source = SourceProject
			}

			r.register(skill)
		}
	}
	return nil
}

// register adds a skill to the registry.
func (r *Registry) register(skill *Skill) {
	// By trigger
	if skill.Trigger != "" {
		normalizedTrigger := normalizeTrigger(skill.Trigger)
		r.skills[normalizedTrigger] = skill
		r.triggers[normalizedTrigger] = skill.Name
	}
	// By name
	if skill.Name != "" {
		r.byName[skill.Name] = skill
	}
	// By path
	r.byPath[skill.Path] = skill
}

// GetByTrigger returns the skill for a given trigger.
func (r *Registry) GetByTrigger(trigger string) (*Skill, bool) {
	normalized := normalizeTrigger(trigger)
	skill, ok := r.skills[normalized]
	return skill, ok
}

// GetByName returns the skill by name.
func (r *Registry) GetByName(name string) (*Skill, bool) {
	skill, ok := r.byName[name]
	return skill, ok
}

// List returns all skills.
func (r *Registry) List() []*Skill {
	skills := make([]*Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}
	return skills
}

// GetTriggers returns all triggers as a slice.
func (r *Registry) GetTriggers() []string {
	triggers := make([]string, 0, len(r.triggers))
	for trigger := range r.triggers {
		triggers = append(triggers, trigger)
	}
	return triggers
}

// HasTrigger checks if a trigger is registered.
func (r *Registry) HasTrigger(trigger string) bool {
	normalized := normalizeTrigger(trigger)
	_, ok := r.skills[normalized]
	return ok
}

// MatchUserInput checks if user input matches any skill trigger.
// Returns the matching skill and the remaining input.
func (r *Registry) MatchUserInput(input string) (*Skill, string) {
	// Check for exact trigger match first
	if skill, ok := r.GetByTrigger(input); ok {
		return skill, ""
	}

	// Check for trigger prefix match
	for _, skill := range r.skills {
		trigger := normalizeTrigger(skill.Trigger)
		if strings.HasPrefix(input, trigger) {
			// Get the remaining input after the trigger
			remaining := strings.TrimSpace(strings.TrimPrefix(input, trigger))
			return skill, remaining
		}
	}

	return nil, ""
}

// normalizeTrigger normalizes a trigger string (adds leading slash if missing).
func normalizeTrigger(trigger string) string {
	trigger = strings.TrimSpace(trigger)
	if !strings.HasPrefix(trigger, "/") {
		trigger = "/" + trigger
	}
	return trigger
}

// FormatSkill formats a skill for display.
func (s *Skill) FormatSkill() string {
	return fmt.Sprintf("%s — %s", s.Trigger, s.Description)
}
