package prompt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextAssembly_Format(t *testing.T) {
	assembly := &ContextAssembly{
		BasePrompt:      "You are helpful",
		ProjectSection:  "Project rules",
		SkillsSection:   "Skills: test",
		MemorySection:   "Memory facts",
		FullSystemPrompt: "You are helpful\n\nProject rules\n\nSkills: test",
		ToolsCount:      5,
		HistoryLength:   10,
		ProviderInfo:    "anthropic",
		ModelInfo:       "claude-sonnet-4-6",
		SessionID:       "session-123",
	}

	result := assembly.Format()

	assert.Contains(t, result, "=== Assembled Context ===")
	assert.Contains(t, result, "## System Prompt")
	assert.Contains(t, result, "Base prompt: 15 chars")
	assert.Contains(t, result, "Project instructions: 13 chars")
	assert.Contains(t, result, "Skills section: 12 chars")
	assert.Contains(t, result, "Memory section: 12 chars")
	assert.Contains(t, result, "Full system prompt: 44 chars total")
	assert.Contains(t, result, "## Runtime Context")
	assert.Contains(t, result, "Tools available: 5")
	assert.Contains(t, result, "History messages: 10")
	assert.Contains(t, result, "Provider: anthropic")
	assert.Contains(t, result, "Model: claude-sonnet-4-6")
	assert.Contains(t, result, "Session: session-123")
}

func TestContextAssembly_Format_OptionalSections(t *testing.T) {
	assembly := &ContextAssembly{
		BasePrompt:       "You are helpful",
		FullSystemPrompt: "You are helpful",
		ToolsCount:       3,
		HistoryLength:    5,
	}

	result := assembly.Format()

	assert.Contains(t, result, "Base prompt: 15 chars")
	assert.NotContains(t, result, "Project instructions:")
	assert.NotContains(t, result, "Skills section:")
	assert.NotContains(t, result, "Memory section:")
	assert.NotContains(t, result, "Provider:")
	assert.NotContains(t, result, "Model:")
	assert.NotContains(t, result, "Session:")
}

func TestContextAssembly_Format_EmptyStrings(t *testing.T) {
	assembly := &ContextAssembly{
		BasePrompt:       "",
		FullSystemPrompt: "",
	}

	result := assembly.Format()

	assert.Contains(t, result, "Base prompt: 0 chars")
	assert.Contains(t, result, "Full system prompt: 0 chars total")
}

func TestContextAssembly_FormatVerbose(t *testing.T) {
	assembly := &ContextAssembly{
		BasePrompt:      "You are Atom, an AI coding assistant",
		ProjectSection:  "# Project\nRules here",
		SkillsSection:   "- Skill 1\n- Skill 2",
		MemorySection:   "Memory content",
		FullSystemPrompt: "Full prompt",
		ToolsCount:      10,
		HistoryLength:   20,
		ProviderInfo:    "openai",
		ModelInfo:       "gpt-4o",
		SessionID:       "session-456",
	}

	result := assembly.FormatVerbose()

	assert.Contains(t, result, "=== Assembled Context (Verbose) ===")
	assert.Contains(t, result, "## System Prompt Components")
	assert.Contains(t, result, "### Base Prompt")
	assert.Contains(t, result, "You are Atom")
	assert.Contains(t, result, "### Project Instructions")
	assert.Contains(t, result, "### Skills")
	assert.Contains(t, result, "### Memory")
	assert.Contains(t, result, "## Runtime Context")
	assert.Contains(t, result, "Tools: 10")
	assert.Contains(t, result, "History: 20 messages")
	assert.Contains(t, result, "Provider: openai")
	assert.Contains(t, result, "Model: gpt-4o")
}

func TestContextAssembly_FormatVerbose_WithTruncation(t *testing.T) {
	longText := strings.Repeat("x ", 1000) // 2000 chars

	assembly := &ContextAssembly{
		BasePrompt:      longText,
		ProjectSection:  longText,
		SkillsSection:   longText,
		MemorySection:   longText,
		FullSystemPrompt: longText,
	}

	result := assembly.FormatVerbose()

	assert.Contains(t, result, "(truncated)")
}

func TestContextAssembly_FormatVerbose_OptionalSections(t *testing.T) {
	assembly := &ContextAssembly{
		BasePrompt:       "You are helpful",
		FullSystemPrompt: "You are helpful",
		ToolsCount:       2,
		HistoryLength:    3,
	}

	result := assembly.FormatVerbose()

	assert.Contains(t, result, "### Base Prompt")
	assert.NotContains(t, result, "### Project Instructions")
	assert.NotContains(t, result, "### Skills")
	assert.NotContains(t, result, "### Memory")
	assert.Contains(t, result, "Tools: 2")
	assert.Contains(t, result, "History: 3 messages")
}

func TestBuildContextAssembly(t *testing.T) {
	input := SystemPromptInput{
		Project: ProjectContext{
			Root: "/project",
			Instructions: []InstructionFile{
				{Name: "README.md", Content: "rules"},
			},
		},
		Skills: []SkillSummary{
			{Name: "skill1", Description: "First skill"},
			{Name: "skill2", Description: "Second skill"},
		},
		MemorySnippet: "Memory content",
	}

	assembly := BuildContextAssembly(input, "full system prompt")

	assert.Equal(t, baseSystemPrompt, assembly.BasePrompt)
	assert.Equal(t, "full system prompt", assembly.FullSystemPrompt)
	assert.NotEmpty(t, assembly.ProjectSection)
	assert.NotEmpty(t, assembly.SkillsSection)
	assert.NotEmpty(t, assembly.MemorySection)
}

func TestTruncate_Short(t *testing.T) {
	result := truncate("hello", 10)
	assert.Equal(t, "hello", result)
}

func TestTruncate_Exact(t *testing.T) {
	result := truncate("hello", 5)
	assert.Equal(t, "hello", result)
}

func TestTruncate_Long(t *testing.T) {
	result := truncate("hello world", 5)
	assert.Equal(t, "hello\n... (truncated)", result)
}

func TestTruncate_Empty(t *testing.T) {
	result := truncate("", 10)
	assert.Empty(t, result)
}

func TestTruncate_Unicode(t *testing.T) {
	// Test with unicode content
	result := truncate("你好世界", 3)
	// Chinese characters take 3 bytes in UTF-8: "你好" = 6 bytes
	// truncate(3) will give 1 byte, which may be invalid UTF-8
	// The important thing is it doesn't panic and returns something
	assert.NotEmpty(t, result)
}