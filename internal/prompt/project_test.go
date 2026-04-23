package prompt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscoverProjectContext_EmptyDir(t *testing.T) {
	root := t.TempDir()

	ctx, err := DiscoverProjectContext(root)
	assert.NoError(t, err)
	assert.Empty(t, ctx.Root)
	assert.Empty(t, ctx.Instructions)
}

func TestDiscoverProjectContext_NoInstructionsFound(t *testing.T) {
	root := t.TempDir()

	ctx, err := DiscoverProjectContext(root)
	assert.NoError(t, err)
	assert.Empty(t, ctx.Root)
	assert.Empty(t, ctx.Instructions)
}

func TestDiscoverProjectContext_OnlyAGENTSmd(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("agent rules"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, err := DiscoverProjectContext(root)
	assert.NoError(t, err)
	assert.Equal(t, root, ctx.Root)
	assert.Len(t, ctx.Instructions, 1)
	assert.Equal(t, "AGENTS.md", ctx.Instructions[0].Name)
}

func TestDiscoverProjectContext_OnlyCLAUDEmd(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("claude rules"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, err := DiscoverProjectContext(root)
	assert.NoError(t, err)
	assert.Equal(t, root, ctx.Root)
	assert.Len(t, ctx.Instructions, 1)
	assert.Equal(t, "CLAUDE.md", ctx.Instructions[0].Name)
}

func TestDiscoverProjectContext_EmptyInstructionFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("   \n   "), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, err := DiscoverProjectContext(root)
	assert.NoError(t, err)
	assert.Empty(t, ctx.Root) // No actual content, so no root found
}

func TestFormatProjectInstructions_None(t *testing.T) {
	result := FormatProjectInstructions(ProjectContext{})
	assert.Empty(t, result)
}

func TestFormatProjectInstructions_WithRoot(t *testing.T) {
	ctx := ProjectContext{
		Root: "/my/project",
		Instructions: []InstructionFile{
			{Name: "README.md", Content: "instructions"},
		},
	}

	result := FormatProjectInstructions(ctx)
	assert.Contains(t, result, "Project root: /my/project")
	assert.Contains(t, result, "### README.md")
	assert.Contains(t, result, "instructions")
}

func TestFormatProjectInstructions_WithoutRoot(t *testing.T) {
	ctx := ProjectContext{
		Instructions: []InstructionFile{
			{Name: "README.md", Content: "instructions"},
		},
	}

	result := FormatProjectInstructions(ctx)
	assert.NotContains(t, result, "Project root:")
	assert.Contains(t, result, "### README.md")
}

func TestLoadInstructionsFromDir_BothFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("claude"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("agents"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, found, err := loadInstructionsFromDir(root)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, ctx.Instructions, 2)

	names := make([]string, 2)
	for i, inst := range ctx.Instructions {
		names[i] = inst.Name
	}
	assert.Contains(t, names, "CLAUDE.md")
	assert.Contains(t, names, "AGENTS.md")
}

func TestLoadInstructionsFromDir_NonexistentFile(t *testing.T) {
	root := t.TempDir()

	ctx, found, err := loadInstructionsFromDir(root)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Empty(t, ctx.Instructions)
}

func TestBuildSystemPrompt_WithAllSections(t *testing.T) {
	input := SystemPromptInput{
		Project: ProjectContext{
			Root: "/project",
			Instructions: []InstructionFile{
				{Name: "README.md", Content: "project rules"},
			},
		},
		Skills: []SkillSummary{
			{Name: "test", Trigger: "/test", Description: "Test skill", Source: "global"},
		},
		MemorySnippet: "# Memory\nImportant facts",
	}

	result := BuildSystemPrompt(input)

	assert.Contains(t, result, "You are Atom")
	assert.Contains(t, result, "## Project Instructions")
	assert.Contains(t, result, "## Loaded Skills")
	assert.Contains(t, result, "test (/test): Test skill [global]")
	assert.Contains(t, result, "## Persistent Memory")
	assert.Contains(t, result, "Important facts")
}

func TestBuildSystemPrompt_BaseOnly(t *testing.T) {
	input := SystemPromptInput{}
	result := BuildSystemPrompt(input)

	assert.Contains(t, result, "You are Atom")
	assert.NotContains(t, result, "## Project Instructions")
	assert.NotContains(t, result, "## Loaded Skills")
	assert.NotContains(t, result, "## Persistent Memory")
}

func TestFormatSkills_Empty(t *testing.T) {
	result := formatSkills([]SkillSummary{})
	assert.Empty(t, result)
}

func TestFormatSkills_Variants(t *testing.T) {
	skills := []SkillSummary{
		{Name: "minimal"},
		{Name: "with-trigger", Trigger: "/cmd"},
		{Name: "with-desc", Description: "A skill"},
		{Name: "with-source", Source: "project"},
		{Name: "full", Trigger: "/full", Description: "Full skill", Source: "global"},
	}

	result := formatSkills(skills)

	assert.Contains(t, result, "- minimal")
	assert.Contains(t, result, "- with-trigger (/cmd)")
	assert.Contains(t, result, "- with-desc: A skill")
	assert.Contains(t, result, "- with-source [project]")
	assert.Contains(t, result, "- full (/full): Full skill [global]")
}

func TestFormatMemorySnippet_Empty(t *testing.T) {
	result := formatMemorySnippet("")
	assert.Empty(t, result)
}

func TestFormatMemorySnippet_Whitespace(t *testing.T) {
	result := formatMemorySnippet("   \n\n   ")
	assert.Empty(t, result)
}

func TestFormatMemorySnippet_WithContent(t *testing.T) {
	result := formatMemorySnippet("# Important\nFacts to remember")

	assert.Contains(t, result, "## Persistent Memory")
	assert.Contains(t, result, "Important")
	assert.Contains(t, result, "Facts to remember")
}

func TestDiscoverProjectContext_StartFromNested(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("root rules"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, err := DiscoverProjectContext(nested)
	assert.NoError(t, err)
	assert.Equal(t, root, ctx.Root)
}