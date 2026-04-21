// Package plugin provides plugin contract and loading mechanism.
package plugin

import (
	"fmt"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

// Version represents a semantic version.
type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Info describes plugin metadata.
type Info struct {
	Name        string
	Version     Version
	Description string
	Author      string
}

// Plugin defines the interface for external plugins.
type Plugin interface {
	// Info returns plugin metadata.
	Info() Info

	// Tools returns all tools provided by this plugin.
	Tools() []tools.Tool

	// Register is called when the plugin is being registered.
	// Plugins can perform initialization here.
	Register(registry *tools.Registry) error

	// Unregister is called when the plugin is being unloaded.
	// Plugins should clean up resources here.
	Unregister() error
}

// Manager handles plugin loading and lifecycle.
type Manager struct {
	plugins map[string]Plugin
}

// NewManager creates a new plugin manager.
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
	}
}

// Load adds a plugin to the manager and registers its tools.
func (m *Manager) Load(p Plugin, registry *tools.Registry) error {
	info := p.Info()

	// Check for duplicate plugin names
	if _, exists := m.plugins[info.Name]; exists {
		return fmt.Errorf("plugin %q is already loaded", info.Name)
	}

	// Register the plugin
	if err := p.Register(registry); err != nil {
		return fmt.Errorf("failed to register plugin %q: %w", info.Name, err)
	}

	m.plugins[info.Name] = p
	return nil
}

// Unload removes a plugin from the manager.
func (m *Manager) Unload(name string) error {
	p, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %q is not loaded", name)
	}

	// Call the plugin's unregister function
	if err := p.Unregister(); err != nil {
		return fmt.Errorf("failed to unregister plugin %q: %w", name, err)
	}

	delete(m.plugins, name)
	return nil
}

// List returns all loaded plugins.
func (m *Manager) List() []Info {
	var infos []Info
	for _, p := range m.plugins {
		infos = append(infos, p.Info())
	}
	return infos
}

// Get returns a plugin by name, or nil if not found.
func (m *Manager) Get(name string) Plugin {
	return m.plugins[name]
}

// IsLoaded returns true if a plugin with the given name is loaded.
func (m *Manager) IsLoaded(name string) bool {
	_, exists := m.plugins[name]
	return exists
}

// Close unloads all plugins.
func (m *Manager) Close() error {
	var errors []error
	for name, p := range m.plugins {
		if err := p.Unregister(); err != nil {
			errors = append(errors, fmt.Errorf("failed to unregister plugin %q: %w", name, err))
		}
	}
	m.plugins = make(map[string]Plugin)

	if len(errors) > 0 {
		return fmt.Errorf("%d plugin(s) failed to unregister", len(errors))
	}
	return nil
}