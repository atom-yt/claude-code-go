# Architecture Documentation

## Project: claude-code-go

Claude Code Go is a Go implementation of Claude Code - an AI-driven terminal CLI tool.

## Layer Architecture

This project follows a strict layered architecture to maintain separation of concerns and prevent circular dependencies.

### Layer Mapping

```
Layer 0 (Foundation):  pkg/
Layer 1 (Core):        internal/messages/, internal/pathutil/, internal/urlutil/
Layer 2 (Services):    internal/config/, internal/permissions/, internal/hooks/, internal/memory/, internal/providers/, internal/compact/
Layer 3 (Tools):      internal/tools/, internal/api/, internal/commands/, internal/skills/, internal/mcp/, internal/prompt/
Layer 4 (Orchestration): internal/agent/, internal/tui/, internal/session/, internal/subagent/, internal/runtime/, internal/taskstore/
```

### Dependency Rules

- **Higher layers can import lower layers**
- **Lower layers CANNOT import higher layers**
- **Same layer imports allowed where appropriate**

**Examples:**
- ✓ `internal/tools/` → `internal/api/` (Layer 3 → Layer 3, same layer)
- ✓ `internal/agent/` → `internal/tools/` (Layer 4 → Layer 3, valid)
- ✓ `internal/api/` → `internal/config/` (Layer 3 → Layer 2, valid)
- ✗ `internal/config/` → `internal/agent/` (Layer 2 → Layer 4, **VIOLATION**)
- ✗ `internal/agent/` → `pkg/` (Layer 4 → Layer 0, allowed - foundational)
- ✗ `pkg/` → `internal/config/` (Layer 0 → Layer 2, **VIOLATION**)

## Package Structure

### pkg/
Foundation packages with no internal dependencies.
- `pkg/anthropic/` - Anthropic API client

### internal/messages/
Message types and data structures used throughout the application.

### internal/pathutil/
Path validation and utility functions.

### internal/urlutil/
URL validation with SSRF protection.

### internal/config/
Configuration management and settings.

### internal/permissions/
Permission system for tool execution.

### internal/hooks/
Hook system for pre/post tool execution.

### internal/memory/
Session memory and cross-session persistence.

### internal/providers/
Provider registry and configuration.

### internal/compact/
Message compaction services.

### internal/tools/
Tool implementations (Read, Write, Bash, Glob, Grep, etc.).

### internal/api/
API clients (Anthropic, OpenAI) for LLM communication.

### internal/commands/
Slash command implementations.

### internal/skills/
Skill system and execution.

### internal/mcp/
Model Context Protocol client implementation.

### internal/prompt/
Prompt building and context management.

### internal/agent/
Agent main loop and orchestration.

### internal/tui/
Terminal User Interface using bubbletea.

### internal/session/
Session management and persistence.

### internal/subagent/
Subagent runtime for parallel execution.

### internal/runtime/
Runtime state management.

### internal/taskstore/
Durable task storage for plan mode.

## Tool System

All tools implement the `Tool` interface defined in `internal/tools/tool.go`:

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]any
    Call(ctx context.Context, input map[string]any) (ToolResult, error)
    IsReadOnly() bool
    IsConcurrencySafe() bool
}
```

## Provider System

Providers are registered in `internal/providers/registry.go` and support both Anthropic and OpenAI protocols.

## Session Lifecycle

1. Configuration load (env vars → project settings → user settings)
2. Session start (TUI initialization)
3. Agent loop (user input → API → tool execution → repeat)
4. Session save (to memory)
5. Session exit

## Import Guidelines

1. **No circular dependencies** - Use dependency injection when needed
2. **Interface-based design** - Define interfaces at layer boundaries
3. **Context propagation** - Always pass context.Context through the stack
4. **Error handling** - Wrap errors with context, don't ignore errors