package commands

import (
	"context"
	"fmt"
	"strings"
)

// All returns all built-in commands.
func All() []Command {
	return []Command{
		&helpCmd{},
		&clearCmd{},
		&modelCmd{},
		&costCmd{},
		&compactCmd{},
	}
}

// Registry maps command names (and aliases) to commands.
type Registry struct {
	index map[string]Command
	list  []Command
}

// NewRegistry creates a registry pre-loaded with all built-in commands.
func NewRegistry() *Registry {
	r := &Registry{index: make(map[string]Command)}
	for _, cmd := range All() {
		r.list = append(r.list, cmd)
		r.index[cmd.Name()] = cmd
		for _, alias := range cmd.Aliases() {
			r.index[alias] = cmd
		}
	}
	return r
}

// Get returns the command for the given name/alias.
func (r *Registry) Get(name string) (Command, bool) {
	c, ok := r.index[strings.TrimPrefix(name, "/")]
	return c, ok
}

// List returns all registered commands.
func (r *Registry) List() []Command { return r.list }

// ---- /help ----

type helpCmd struct{}

func (c *helpCmd) Name() string        { return "help" }
func (c *helpCmd) Aliases() []string   { return []string{"h", "?"} }
func (c *helpCmd) Description() string { return "Show available slash commands" }

func (c *helpCmd) Execute(_ context.Context, _ []string, _ *Context) (string, error) {
	lines := []string{"Available commands:"}
	for _, cmd := range All() {
		aliases := ""
		if len(cmd.Aliases()) > 0 {
			aliases = fmt.Sprintf(" (/%s)", strings.Join(cmd.Aliases(), ", /"))
		}
		lines = append(lines, fmt.Sprintf("  /%s%s — %s", cmd.Name(), aliases, cmd.Description()))
	}
	return strings.Join(lines, "\n"), nil
}

// ---- /clear ----

type clearCmd struct{}

func (c *clearCmd) Name() string        { return "clear" }
func (c *clearCmd) Aliases() []string   { return []string{"cl"} }
func (c *clearCmd) Description() string { return "Clear conversation history" }

func (c *clearCmd) Execute(_ context.Context, _ []string, ctx *Context) (string, error) {
	if ctx.ClearMessages != nil {
		ctx.ClearMessages()
	}
	return "", nil
}

// ---- /model ----

type modelCmd struct{}

func (c *modelCmd) Name() string        { return "model" }
func (c *modelCmd) Aliases() []string   { return nil }
func (c *modelCmd) Description() string { return "Show or switch model: /model [name]" }

func (c *modelCmd) Execute(_ context.Context, args []string, ctx *Context) (string, error) {
	if len(args) == 0 {
		model := ""
		if ctx.GetModel != nil {
			model = ctx.GetModel()
		}
		return fmt.Sprintf("Current model: %s", model), nil
	}
	newModel := strings.Join(args, " ")
	if ctx.SetModel != nil {
		ctx.SetModel(newModel)
	}
	return fmt.Sprintf("Switched to model: %s", newModel), nil
}

// ---- /cost ----

type costCmd struct{}

func (c *costCmd) Name() string        { return "cost" }
func (c *costCmd) Aliases() []string   { return []string{"tokens"} }
func (c *costCmd) Description() string { return "Show cumulative token usage" }

func (c *costCmd) Execute(_ context.Context, _ []string, ctx *Context) (string, error) {
	if ctx.GetCost == nil {
		return "Token tracking not available", nil
	}
	in, out := ctx.GetCost()
	total := in + out
	return fmt.Sprintf("Tokens used — input: %d  output: %d  total: %d", in, out, total), nil
}

// ---- /compact ----

type compactCmd struct{}

func (c *compactCmd) Name() string        { return "compact" }
func (c *compactCmd) Aliases() []string   { return nil }
func (c *compactCmd) Description() string { return "Summarise conversation history to save context" }

func (c *compactCmd) Execute(ctx context.Context, _ []string, cmdCtx *Context) (string, error) {
	if cmdCtx.CompactHistory == nil {
		return "Compact not available", nil
	}
	if err := cmdCtx.CompactHistory(ctx); err != nil {
		return "", fmt.Errorf("compact failed: %w", err)
	}
	return "History compacted.", nil
}
