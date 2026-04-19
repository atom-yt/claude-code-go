package memory

import (
	"fmt"
	"os"
	"strings"
)

// BuildConsolidationPrompt creates the prompt for memory consolidation.
func BuildConsolidationPrompt() (string, error) {
	var parts []string

	// System prompt
	parts = append(parts, `You are conducting an auto-dream memory consolidation. Review the session summaries and existing memory to extract and update knowledge.

**Your task:**
1. Read existing memory files (including MEMORY.md index)
2. Review recent session summaries
3. Identify new insights, patterns, or knowledge to add
4. Update memory files with new information
5. Update MEMORY.md index as needed

**Constraints:**
- Use Read, Glob, Grep tools only (no Write, Edit, Bash)
- Provide the exact content for new/updated files in your response
- Keep memory entries focused and actionable
- Update MEMORY.md to reference new files
- Do not duplicate information already in memory
- Organize by topic/area of knowledge`)

	// Add session summaries
	lockState, err := GetLockState()
	if err == nil && !lockState.LastConsolidation.IsZero() {
		sessions, err := GetSessionsSince(lockState.LastConsolidation)
		if err == nil && len(sessions) > 0 {
			parts = append(parts, "\n## Recent Session Summaries\n")
			for i, s := range sessions {
				if i >= 50 { // Limit to 50 most recent sessions
					parts = append(parts, fmt.Sprintf("... and %d more sessions\n", len(sessions)-50))
					break
				}
				parts = append(parts, fmt.Sprintf("- Session %s (model: %s, updated: %s)\n",
					s.ID, s.Model, s.UpdatedAt.Format("2006-01-02 15:04")))
			}
		}
	}

	// Add existing memory overview
	memDir, err := MemoryRootDir()
	if err == nil {
		parts = append(parts, "\n## Memory Directory Structure\n")
		entries, err := os.ReadDir(memDir)
		if err == nil {
			if len(entries) == 0 {
				parts = append(parts, "(Memory directory is empty - this is the first consolidation)\n")
			} else {
				for _, e := range entries {
					if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
						continue
					}
					parts = append(parts, fmt.Sprintf("- %s\n", e.Name()))
				}
			}
		}
	}

	return strings.Join(parts, "\n"), nil
}

// BuildStatusPrompt creates a prompt for displaying consolidation status.
func BuildStatusPrompt() (string, error) {
	var parts []string

	lockState, err := GetLockState()
	if err != nil {
		return "", err
	}

	parts = append(parts, "# Auto-Dream Status\n")
	parts = append(parts, fmt.Sprintf("Running: %v\n", lockState.IsRunning))

	if !lockState.LastConsolidation.IsZero() {
		parts = append(parts, fmt.Sprintf("Last consolidation: %s\n", lockState.LastConsolidation.Format("2006-01-02 15:04:05")))

		sessions, err := CountSessionsSince(lockState.LastConsolidation)
		if err == nil {
			parts = append(parts, fmt.Sprintf("Sessions since last consolidation: %d\n", sessions))
		}
	} else {
		parts = append(parts, "Last consolidation: never\n")
	}

	// Memory file count
	memDir, err := MemoryRootDir()
	if err == nil {
		entries, err := os.ReadDir(memDir)
		if err == nil {
			count := 0
			for _, e := range entries {
				if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
					count++
				}
			}
			parts = append(parts, fmt.Sprintf("Memory files: %d\n", count))
		}
	}

	memIndexPath, err := MemoryIndexPath()
	if err == nil {
		if _, err := os.Stat(memIndexPath); err == nil {
			parts = append(parts, "MEMORY.md: exists\n")
		} else {
			parts = append(parts, "MEMORY.md: not found\n")
		}
	}

	return strings.Join(parts, "\n"), nil
}
