package tools

import "fmt"

// APIToolSpec is the format Anthropic expects for each tool in a request.
type APIToolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// Registry holds the available tools indexed by name.
type Registry struct {
	tools []Tool
	index map[string]Tool
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{index: make(map[string]Tool)}
}

// Register adds a tool. Panics if a tool with the same name is already registered.
func (r *Registry) Register(t Tool) {
	if _, exists := r.index[t.Name()]; exists {
		panic(fmt.Sprintf("tool %q already registered", t.Name()))
	}
	r.tools = append(r.tools, t)
	r.index[t.Name()] = t
}

// GetAll returns all registered tools in registration order.
func (r *Registry) GetAll() []Tool {
	return r.tools
}

// GetByName returns the tool with the given name and whether it was found.
func (r *Registry) GetByName(name string) (Tool, bool) {
	t, ok := r.index[name]
	return t, ok
}

// ToAPISpecs converts all tools to the format expected by the Anthropic API.
func (r *Registry) ToAPISpecs() []APIToolSpec {
	specs := make([]APIToolSpec, 0, len(r.tools))
	for _, t := range r.tools {
		specs = append(specs, APIToolSpec{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.InputSchema(),
		})
	}
	return specs
}
