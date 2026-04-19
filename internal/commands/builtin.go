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
		&buddyCmd{},
		&shortcutsCmd{},
		&commitCmd{},
		&prCmd{},
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
	// Use raw marker to skip markdown processing and preserve newlines
	return "<!-- raw -->\n" + strings.Join(lines, "\n"), nil
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
func (c *modelCmd) Description() string { return "Show or switch model: /model [provider/model] or /model [model]" }

func (c *modelCmd) Execute(_ context.Context, args []string, ctx *Context) (string, error) {
	if len(args) == 0 {
		model := ""
		provider := ""
		if ctx.GetModel != nil {
			model = ctx.GetModel()
		}
		if ctx.GetProvider != nil {
			provider = ctx.GetProvider()
		}
		if provider == "" {
			provider = "anthropic"
		}
		return fmt.Sprintf("Current model: %s/%s", provider, model), nil
	}

	input := strings.Join(args, " ")

	// Parse provider/model format.
	var newProvider, newModel string
	if idx := strings.Index(input, "/"); idx > 0 {
		newProvider = input[:idx]
		newModel = input[idx+1:]
	} else {
		newModel = input
	}

	if newProvider != "" && ctx.SetProvider != nil {
		ctx.SetProvider(newProvider)
	}
	if ctx.SetModel != nil {
		ctx.SetModel(newModel)
	}

	provider := ""
	if ctx.GetProvider != nil {
		provider = ctx.GetProvider()
	}
	if provider == "" {
		provider = "anthropic"
	}
	return fmt.Sprintf("Switched to model: %s/%s", provider, ctx.GetModel()), nil
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

// ---- /buddy ----

type buddyCmd struct{}

func (c *buddyCmd) Name() string        { return "buddy" }
func (c *buddyCmd) Aliases() []string   { return []string{"pet"} }
func (c *buddyCmd) Description() string { return "Show your AI buddy! Try /buddy [happy|sad|thinking|sleeping|eating|play]" }

func (c *buddyCmd) Execute(_ context.Context, args []string, _ *Context) (string, error) {
	mood := "happy"
	if len(args) > 0 {
		mood = args[0]
	}

	// Buddy art collection - use raw text prefix to skip markdown rendering
	buddies := map[string]string{
		"happy": "<!-- raw -->\n" + `      /\__      *  *  *
     (   @\___        /
    /        O      /
   /   (_____/
  /_____/   U

Your buddy is happy! 🐕`,
		"sad": "<!-- raw -->\n" + `      /\__
     (   @\___
    /        O
   /   (_____/  | |
  /_____/   |   |

Your buddy is sad... give them a pat! 🐕`,
		"thinking": "<!-- raw -->\n" + `      /\__
     (   @\___
    /        ?
   /   (_____)
  /_____/   |

Your buddy is thinking... 🤔`,
		"sleeping": "<!-- raw -->\n" + `      /\__
     (   @\___
    /        Z
   /   (_____) z
  /_____/   ZZZ

Your buddy is sleeping... 💤`,
		"eating": "<!-- raw -->\n" + `      /\__
     (   @\___
    /       ( )
   /   (_____)/
  /_____/   |

Your buddy is eating a treat! 🦴`,
		"play": "<!-- raw -->\n" + `      /\__
     (   @\___
    /        !
   /   (_____)/
  /_____/   |

Your buddy wants to play! 🎾`,
	}

	art, ok := buddies[mood]
	if !ok {
		art = buddies["happy"]
	}

	return art, nil
}

// ---- /shortcuts ----

type shortcutsCmd struct{}

func (c *shortcutsCmd) Name() string        { return "shortcuts" }
func (c *shortcutsCmd) Aliases() []string   { return []string{"keys"} }
func (c *shortcutsCmd) Description() string { return "Show all keyboard shortcuts" }

func (c *shortcutsCmd) Execute(_ context.Context, _ []string, _ *Context) (string, error) {
	lines := []string{"Keyboard shortcuts:"}

	shortcuts := []struct {
		key         string
		description string
	}{
		{"Ctrl+C / Ctrl+D", "Quit"},
		{"Ctrl+L", "Clear screen"},
		{"Enter", "Submit input"},
		{"Ctrl+J", "Insert newline"},
		{"Backspace/Delete", "Delete character"},
		{"Up / Down", "Navigate input history"},
		{"PgUp / PgDn", "Scroll history"},
		{"Mouse wheel", "Scroll history"},
		{"Tab (when typing /)", "Cycle through command suggestions"},
		{"Esc", "Dismiss autocomplete menu"},
	}

	maxKeyLen := 20
	for _, s := range shortcuts {
		if len(s.key) > maxKeyLen {
			maxKeyLen = len(s.key)
		}
	}

	for _, s := range shortcuts {
		padding := strings.Repeat(" ", maxKeyLen-len(s.key))
		lines = append(lines, fmt.Sprintf("  %s%s  —  %s", s.key, padding, s.description))
	}

	return "<!-- raw -->\n" + strings.Join(lines, "\n"), nil
}
