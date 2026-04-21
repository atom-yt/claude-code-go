package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/atom-yt/claude-code-go/internal/tools"
)

// mockPlugin is a simple plugin implementation for testing.
type mockPlugin struct {
	info        Info
	tools       []tools.Tool
	initialized bool
}

func (m *mockPlugin) Info() Info {
	return m.info
}

func (m *mockPlugin) Tools() []tools.Tool {
	return m.tools
}

func (m *mockPlugin) Register(registry *tools.Registry) error {
	m.initialized = true
	return nil
}

func (m *mockPlugin) Unregister() error {
	m.initialized = false
	return nil
}

func TestVersionString(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	assert.Equal(t, "1.2.3", v.String())
}

func TestNewManager(t *testing.T) {
	mgr := NewManager()
	assert.NotNil(t, mgr)
	assert.Empty(t, mgr.List())
}

func TestManagerLoad(t *testing.T) {
	mgr := NewManager()
	reg := tools.NewRegistry()

	p := &mockPlugin{
		info: Info{
			Name:    "test-plugin",
			Version: Version{Major: 1, Minor: 0, Patch: 0},
		},
	}

	err := mgr.Load(p, reg)
	assert.NoError(t, err)
	assert.True(t, mgr.IsLoaded("test-plugin"))
	assert.Equal(t, p, mgr.Get("test-plugin"))
	assert.True(t, p.initialized)
}

func TestManagerDuplicateLoad(t *testing.T) {
	mgr := NewManager()
	reg := tools.NewRegistry()

	p := &mockPlugin{
		info: Info{
			Name: "test-plugin",
			Version: Version{Major: 1, Minor: 0, Patch: 0},
		},
	}

	// First load should succeed
	err := mgr.Load(p, reg)
	assert.NoError(t, err)

	// Second load should fail
	err = mgr.Load(p, reg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already loaded")
}

func TestManagerUnload(t *testing.T) {
	mgr := NewManager()
	reg := tools.NewRegistry()

	p := &mockPlugin{
		info: Info{
			Name: "test-plugin",
			Version: Version{Major: 1, Minor: 0, Patch: 0},
		},
	}

	_ = mgr.Load(p, reg)

	err := mgr.Unload("test-plugin")
	assert.NoError(t, err)
	assert.False(t, mgr.IsLoaded("test-plugin"))
	assert.Nil(t, mgr.Get("test-plugin"))
	assert.False(t, p.initialized)
}

func TestManagerUnloadNotFound(t *testing.T) {
	mgr := NewManager()

	err := mgr.Unload("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not loaded")
}

func TestManagerList(t *testing.T) {
	mgr := NewManager()
	reg := tools.NewRegistry()

	p1 := &mockPlugin{info: Info{Name: "plugin-1"}}
	p2 := &mockPlugin{info: Info{Name: "plugin-2"}}

	_ = mgr.Load(p1, reg)
	_ = mgr.Load(p2, reg)

	plugins := mgr.List()
	assert.Len(t, plugins, 2)

	names := make([]string, 2)
	for _, p := range plugins {
		names = append(names, p.Name)
	}
	assert.Contains(t, names, "plugin-1")
	assert.Contains(t, names, "plugin-2")
}

func TestManagerClose(t *testing.T) {
	mgr := NewManager()
	reg := tools.NewRegistry()

	p1 := &mockPlugin{info: Info{Name: "plugin-1"}}
	p2 := &mockPlugin{info: Info{Name: "plugin-2"}}

	_ = mgr.Load(p1, reg)
	_ = mgr.Load(p2, reg)

	err := mgr.Close()
	assert.NoError(t, err)
	assert.Empty(t, mgr.List())

	// Plugins should be uninitialized
	assert.False(t, p1.initialized)
	assert.False(t, p2.initialized)
}