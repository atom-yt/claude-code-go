# claude-code-go

A Go implementation of [Claude Code](https://claude.ai/code) — an AI-powered terminal coding assistant powered by [Claude](https://www.anthropic.com/).

## Features

- **Multi-turn conversation** with Claude via the Anthropic API (streaming SSE)
- **Agent tool loop** — Claude can read/write/edit files, run shell commands, search code
- **Built-in tools**: Read, Write, Edit, Bash, Glob, Grep
- **MCP support** — connect external tools via Model Context Protocol servers
- **Permission system** — configurable allow/deny/ask rules per tool and path
- **Hooks** — shell or HTTP hooks on session_start, pre/post tool calls, stop
- **Session persistence** — conversations saved to `~/.claude/sessions/`; resume with `claude resume`
- **Slash commands**: `/help`, `/clear`, `/model`, `/cost`, `/compact`
- **Rich TUI** — scrollable history, Markdown rendering, syntax highlighting, mouse support
- **Config file** — project-local `.claude/settings.json` or `~/.claude/settings.json`

## Installation

### From source

```bash
git clone https://github.com/atom-yt/claude-code-go.git
cd claude-code-go
make install        # installs to $GOPATH/bin/claude
```

### Using go install

```bash
go install github.com/atom-yt/claude-code-go/cmd/claude@latest
```

## Usage

```bash
# Set your API key
export ANTHROPIC_API_KEY=sk-ant-...

# Start interactive session
claude

# Start with an initial prompt
claude "explain this codebase"

# Resume the most recent session
claude resume

# Resume a specific session
claude resume session-1712345678901

# Override model
claude --model claude-opus-4-6

# Show version
claude version
```

## Configuration

Claude Code reads settings from (in priority order):

1. CLI flags (`--model`, `--api-key`, `--verbose`)
2. Environment variables (`ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL`)
3. Project settings: `.claude/settings.json` in the current directory (or any parent)
4. User settings: `~/.claude/settings.json`

### Example `settings.json`

```json
{
  "model": "claude-sonnet-4-6",
  "permissions": {
    "defaultMode": "default",
    "allow": [
      { "tool": "Read" },
      { "tool": "Glob" },
      { "tool": "Grep" }
    ],
    "deny": [
      { "tool": "Bash", "command": "rm -rf" }
    ],
    "ask": [
      { "tool": "Write" },
      { "tool": "Edit" }
    ]
  },
  "hooks": {
    "pre_tool_call": [
      {
        "matcher": "Bash",
        "hooks": [
          { "type": "shell", "command": "echo \"Running: $HOOK_INPUT\" >&2" }
        ]
      }
    ]
  },
  "mcpServers": {
    "my-server": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@my/mcp-server"]
    }
  }
}
```

### Permission modes

| Mode | Description |
|------|-------------|
| `default` | Read-only tools allowed; write/execute tools require explicit allow rule or ask |
| `trust-all` | All tools allowed without prompting |
| `manual` | All tools require explicit allow rule or ask |

## Built-in Tools

| Tool | Description |
|------|-------------|
| `Read` | Read a file with line numbers (supports offset/limit) |
| `Write` | Write content to a file (creates parent dirs automatically) |
| `Edit` | Replace an exact string in a file |
| `Bash` | Execute a shell command (30s default timeout) |
| `Glob` | Find files matching a glob pattern (supports `**`) |
| `Grep` | Search file contents with regexp (with context lines) |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Submit input |
| `Ctrl+J` | Insert newline |
| `Ctrl+C` / `Ctrl+D` | Quit |
| `Ctrl+L` | Clear screen |
| `↑` / `↓` | Navigate input history |
| `PgUp` / `PgDn` | Scroll conversation |
| Mouse wheel | Scroll conversation |

## Development

```bash
# Build
make build

# Run tests
make test

# Cross-compile for all platforms
make release

# Install locally
make install
```

## License

MIT
