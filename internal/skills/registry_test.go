package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeTrigger(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"graphify", "/graphify"},
		{"/graphify", "/graphify"},
		{"  graphify  ", "/graphify"},
		{"test", "/test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeTrigger(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadSkillFromContent(t *testing.T) {
	skill := LoadSkillFromContent(
		"test",
		"Test skill",
		"/test",
		"Test content",
	)

	assert.Equal(t, "test", skill.Name)
	assert.Equal(t, "Test skill", skill.Description)
	assert.Equal(t, "/test", skill.Trigger)
	assert.Equal(t, "Test content", skill.Content)
	assert.Equal(t, SourceGlobal, skill.Source)
}

func TestLoadSkillFromContent_DefaultTrigger(t *testing.T) {
	skill := LoadSkillFromContent(
		"test",
		"Test skill",
		"", // No trigger
		"Test content",
	)

	assert.Equal(t, "/test", skill.Trigger)
}

func TestSkill_Validate(t *testing.T) {
	t.Run("valid skill", func(t *testing.T) {
		skill := &Skill{Name: "test", Trigger: "/test"}
		assert.NoError(t, skill.Validate())
	})

	t.Run("missing name", func(t *testing.T) {
		skill := &Skill{Name: "", Trigger: "/test"}
		assert.Error(t, skill.Validate())
	})

	t.Run("missing trigger", func(t *testing.T) {
		skill := &Skill{Name: "test", Trigger: ""}
		assert.Error(t, skill.Validate())
	})
}

func TestRegistry_NewRegistry(t *testing.T) {
	registry := NewRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.skills)
	assert.NotNil(t, registry.byName)
	assert.NotNil(t, registry.triggers)
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	skill := &Skill{
		Name:    "test",
		Trigger: "/test",
		Content: "test content",
		Path:    "/test/path",
	}

	registry.register(skill)

	// Check by trigger
	s, ok := registry.GetByTrigger("/test")
	assert.True(t, ok)
	assert.Equal(t, "test", s.Name)

	// Check by name
	s, ok = registry.GetByName("test")
	assert.True(t, ok)
	assert.Equal(t, "/test", s.Trigger)
}

func TestRegistry_HasTrigger(t *testing.T) {
	registry := NewRegistry()
	skill := &Skill{
		Name:    "test",
		Trigger: "/test",
	}

	registry.register(skill)

	assert.True(t, registry.HasTrigger("/test"))
	assert.True(t, registry.HasTrigger("test")) // Normalized
	assert.False(t, registry.HasTrigger("/nonexistent"))
}

func TestRegistry_MatchUserInput(t *testing.T) {
	registry := NewRegistry()
	skill1 := &Skill{Name: "test1", Trigger: "/test1"}
	skill2 := &Skill{Name: "test2", Trigger: "/test2"}
	registry.register(skill1)
	registry.register(skill2)

	// Exact match
	s, remaining := registry.MatchUserInput("/test1")
	assert.NotNil(t, s)
	assert.Equal(t, "test1", s.Name)
	assert.Equal(t, "", remaining)

	// Prefix match
	s, remaining = registry.MatchUserInput("/test1 some args")
	assert.NotNil(t, s)
	assert.Equal(t, "test1", s.Name)
	assert.Equal(t, "some args", remaining)

	// No match
	s, remaining = registry.MatchUserInput("/nonexistent")
	assert.Nil(t, s)
	assert.Equal(t, "", remaining)
}

func TestFormatSkill(t *testing.T) {
	skill := &Skill{
		Name:        "test",
		Description: "Test skill",
		Trigger:     "/test",
	}

	result := skill.FormatSkill()
	assert.Contains(t, result, "/test")
	assert.Contains(t, result, "Test skill")
}
