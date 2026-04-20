package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectContext contains discovered project instruction files.
type ProjectContext struct {
	Root         string
	Instructions []InstructionFile
}

// InstructionFile is one discovered instruction file and its content.
type InstructionFile struct {
	Path    string
	Name    string
	Content string
}

// DiscoverProjectContext walks upward from startDir and collects supported
// instruction files. The nearest directory containing at least one instruction
// file becomes the project root.
func DiscoverProjectContext(startDir string) (ProjectContext, error) {
	dir := startDir
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return ProjectContext{}, err
		}
	}

	for {
		ctx, found, err := loadInstructionsFromDir(dir)
		if err != nil {
			return ProjectContext{}, err
		}
		if found {
			ctx.Root = dir
			return ctx, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ProjectContext{}, nil
		}
		dir = parent
	}
}

func loadInstructionsFromDir(dir string) (ProjectContext, bool, error) {
	candidates := []string{"AGENTS.md", "CLAUDE.md"}
	ctx := ProjectContext{}

	for _, name := range candidates {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return ProjectContext{}, false, err
		}
		content := strings.TrimSpace(string(data))
		if content == "" {
			continue
		}
		ctx.Instructions = append(ctx.Instructions, InstructionFile{
			Path:    path,
			Name:    name,
			Content: content,
		})
	}

	return ctx, len(ctx.Instructions) > 0, nil
}

// FormatProjectInstructions renders the discovered project instructions as one
// system-prompt friendly text block.
func FormatProjectInstructions(ctx ProjectContext) string {
	if len(ctx.Instructions) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "## Project Instructions")
	if ctx.Root != "" {
		parts = append(parts, fmt.Sprintf("Project root: %s", ctx.Root))
	}

	for _, inst := range ctx.Instructions {
		parts = append(parts, fmt.Sprintf("### %s", inst.Name))
		parts = append(parts, inst.Content)
	}

	return strings.Join(parts, "\n\n")
}
