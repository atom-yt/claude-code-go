package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverProjectContextNearestRoot(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "pkg", "inner")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("root claude"), 0o644); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("root agents"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	ctx, err := DiscoverProjectContext(nested)
	if err != nil {
		t.Fatalf("DiscoverProjectContext: %v", err)
	}
	if ctx.Root != root {
		t.Fatalf("want root %q, got %q", root, ctx.Root)
	}
	if len(ctx.Instructions) != 2 {
		t.Fatalf("want 2 instruction files, got %d", len(ctx.Instructions))
	}
}

func TestBuildSystemPromptIncludesProjectInstructions(t *testing.T) {
	system := BuildSystemPrompt(SystemPromptInput{
		Project: ProjectContext{
			Root: "/tmp/project",
			Instructions: []InstructionFile{
				{Name: "AGENTS.md", Content: "agents instructions"},
				{Name: "CLAUDE.md", Content: "claude instructions"},
			},
		},
		Skills: []SkillSummary{
			{Name: "review", Trigger: "/review", Description: "Review code changes", Source: "project"},
		},
		MemorySnippet: "# MEMORY\nrecent facts",
	})

	for _, want := range []string{
		"You are Atom",
		"## Project Instructions",
		"Project root: /tmp/project",
		"### AGENTS.md",
		"agents instructions",
		"### CLAUDE.md",
		"claude instructions",
		"## Loaded Skills",
		"review (/review): Review code changes [project]",
		"## Persistent Memory",
		"recent facts",
	} {
		if !strings.Contains(system, want) {
			t.Fatalf("system prompt missing %q:\n%s", want, system)
		}
	}
}
