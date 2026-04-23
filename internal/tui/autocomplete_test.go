package tui

import (
	"testing"

	"github.com/atom-yt/claude-code-go/internal/commands"
	"github.com/atom-yt/claude-code-go/internal/config"
)

func newTestAutocompleteModel() Model {
	m := Model{
		cfg:    config.Settings{Model: "test-model"},
		cmdRegistry: commands.NewRegistry(),
	}
	return m
}

func TestShowAutocomplete(t *testing.T) {
	m := newTestAutocompleteModel()
	m.showAutocomplete("/he")

	if !m.isAutocompleteActive() {
		t.Error("autocomplete should be active after showAutocomplete")
	}
	if m.autocomplete == nil {
		t.Error("autocomplete state should be set")
	}
}

func TestHideAutocomplete(t *testing.T) {
	m := newTestAutocompleteModel()
	m.showAutocomplete("/he")
	m.hideAutocomplete()

	if m.isAutocompleteActive() {
		t.Error("autocomplete should be inactive after hideAutocomplete")
	}
	if m.autocomplete != nil {
		t.Error("autocomplete state should be nil after hideAutocomplete")
	}
}

func TestIsAutocompleteActive(t *testing.T) {
	m := newTestAutocompleteModel()

	if m.isAutocompleteActive() {
		t.Error("autocomplete should be inactive initially")
	}

	m.showAutocomplete("/he")
	if !m.isAutocompleteActive() {
		t.Error("autocomplete should be active after showAutocomplete")
	}

	m.autocomplete.visible = false
	if m.isAutocompleteActive() {
		t.Error("autocomplete should be inactive when visible is false")
	}
}

func TestUpdateAutocompleteQuery(t *testing.T) {
	m := newTestAutocompleteModel()
	m.showAutocomplete("/he")

	m.updateAutocompleteQuery("/cl")

	if m.autocomplete.query != "cl" {
		t.Errorf("query should be 'cl', got %q", m.autocomplete.query)
	}
}

func TestSelectNextAutocomplete_Wraps(t *testing.T) {
	m := newTestAutocompleteModel()
	m.showAutocomplete("/he")
	m.autocomplete.suggestions = []string{"help", "hello"}
	m.autocomplete.selectedIndex = 1

	m.selectNextAutocomplete()

	if m.autocomplete.selectedIndex != 0 {
		t.Errorf("should wrap to 0, got %d", m.autocomplete.selectedIndex)
	}
}

func TestSelectNextAutocomplete_Increments(t *testing.T) {
	m := newTestAutocompleteModel()
	m.showAutocomplete("/he")
	m.autocomplete.suggestions = []string{"help", "hello"}
	m.autocomplete.selectedIndex = 0

	m.selectNextAutocomplete()

	if m.autocomplete.selectedIndex != 1 {
		t.Errorf("should increment to 1, got %d", m.autocomplete.selectedIndex)
	}
}

func TestSelectNextAutocomplete_NoOpWhenInactive(t *testing.T) {
	m := newTestAutocompleteModel()
	m.autocomplete = nil

	m.selectNextAutocomplete()

	// Should not panic
}

func TestSelectPrevAutocomplete_Wraps(t *testing.T) {
	m := newTestAutocompleteModel()
	m.showAutocomplete("/he")
	m.autocomplete.suggestions = []string{"help", "hello"}
	m.autocomplete.selectedIndex = 0

	m.selectPrevAutocomplete()

	if m.autocomplete.selectedIndex != 1 {
		t.Errorf("should wrap to 1, got %d", m.autocomplete.selectedIndex)
	}
}

func TestSelectPrevAutocomplete_Decrements(t *testing.T) {
	m := newTestAutocompleteModel()
	m.showAutocomplete("/he")
	m.autocomplete.suggestions = []string{"help", "hello"}
	m.autocomplete.selectedIndex = 1

	m.selectPrevAutocomplete()

	if m.autocomplete.selectedIndex != 0 {
		t.Errorf("should decrement to 0, got %d", m.autocomplete.selectedIndex)
	}
}

func TestSelectPrevAutocomplete_NoOpWhenInactive(t *testing.T) {
	m := newTestAutocompleteModel()
	m.autocomplete = nil

	m.selectPrevAutocomplete()

	// Should not panic
}

func TestAcceptAutocomplete_ReplacesInput(t *testing.T) {
	m := newTestAutocompleteModel()
	m.showAutocomplete("/he")
	m.autocomplete.suggestions = []string{"help", "hello"}
	m.autocomplete.selectedIndex = 1

	m.acceptAutocomplete()

	if m.input != "/hello" {
		t.Errorf("input should be '/hello', got %q", m.input)
	}
	if m.isAutocompleteActive() {
		t.Error("autocomplete should be hidden after accept")
	}
}

func TestAcceptAutocomplete_NoOpWhenInactive(t *testing.T) {
	m := newTestAutocompleteModel()
	m.autocomplete = nil
	m.input = "/test"

	m.acceptAutocomplete()

	if m.input != "/test" {
		t.Errorf("input should remain unchanged, got %q", m.input)
	}
}

func TestCycleAutocomplete_ReplacesInput(t *testing.T) {
	m := newTestAutocompleteModel()
	m.showAutocomplete("/he")
	m.autocomplete.suggestions = []string{"help", "hello"}
	m.autocomplete.selectedIndex = 1

	m.cycleAutocomplete()

	if m.input != "/hello" {
		t.Errorf("input should be '/hello', got %q", m.input)
	}
	if m.isAutocompleteActive() {
		t.Error("autocomplete should be hidden after cycle")
	}
}

func TestFilterCommands_EmptyQuery(t *testing.T) {
	m := newTestAutocompleteModel()
	matches := m.filterCommands("")

	if len(matches) == 0 {
		t.Error("should return all commands for empty query")
	}
}

func TestFilterCommands_CaseInsensitive(t *testing.T) {
	m := newTestAutocompleteModel()
	matches := m.filterCommands("HE")

	if len(matches) == 0 {
		t.Error("should match commands case-insensitively")
	}
}

func TestFilterCommands_PrefixMatch(t *testing.T) {
	m := newTestAutocompleteModel()
	matches := m.filterCommands("he")

	for _, cmd := range matches {
		if len(cmd) < 2 || cmd[:2] != "he" {
			t.Errorf("command %q should start with 'he'", cmd)
		}
	}
}