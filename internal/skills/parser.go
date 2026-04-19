package skills

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// parseSkillFile parses a SKILL.md file and returns a Skill.
func parseSkillFile(path string) (*Skill, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var (
		inFrontmatter bool
		frontmatter   strings.Builder
		content       strings.Builder
		skill         = &Skill{
			Path: path,
		}
	)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "---" && !inFrontmatter && frontmatter.Len() == 0 {
			inFrontmatter = true
			continue
		}

		if line == "---" && inFrontmatter {
			inFrontmatter = false
			continue
		}

		if inFrontmatter {
			frontmatter.WriteString(line)
			frontmatter.WriteString("\n")
		} else {
			if content.Len() > 0 {
				content.WriteString("\n")
			}
			content.WriteString(line)
		}
	}

	// Parse frontmatter
	if frontmatter.Len() > 0 {
		if err := parseFrontmatter(frontmatter.String(), skill); err != nil {
			return nil, err
		}
	}

	skill.Content = content.String()

	// Set default trigger if name is set but trigger is not
	if skill.Name != "" && skill.Trigger == "" {
		skill.Trigger = "/" + strings.ToLower(skill.Name)
	}

	return skill, nil
}

// parseFrontmatter parses the YAML-like frontmatter.
func parseFrontmatter(frontmatter string, skill *Skill) error {
	lines := strings.Split(frontmatter, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			skill.Name = value
		case "description":
			skill.Description = value
		case "trigger":
			skill.Trigger = normalizeTrigger(value)
		}
	}

	return nil
}

// LoadSkillFromContent creates a skill from a string content.
// Useful for dynamically creating skills without a file.
func LoadSkillFromContent(name, description, trigger, content string) *Skill {
	skill := &Skill{
		Name:        name,
		Description: description,
		Content:     content,
		Source:      SourceGlobal,
	}

	// Normalize trigger if provided
	if trigger != "" {
		skill.Trigger = normalizeTrigger(trigger)
	} else if skill.Name != "" {
		// Generate default trigger from name
		skill.Trigger = "/" + strings.ToLower(skill.Name)
	}

	return skill
}

// Validate checks if a skill has the required fields.
func (s *Skill) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if s.Trigger == "" {
		return fmt.Errorf("skill trigger is required")
	}
	return nil
}
