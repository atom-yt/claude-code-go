# claude-code-go

A Go implementation of [Claude Code](https://claude.ai/code) — an AI-powered terminal coding assistant powered by [Claude](https://www.anthropic.com/).

## Features

- **Multi-turn conversation** with Claude via the Anthropic API (streaming SSE)
- **Agent tool loop** — Claude can read/write/edit files, run shell commands, search code
- **Built-in tools**: Read, Write, Edit, Bash, Glob, Grep, WebSearch
- **MCP support** — connect external tools via Model Context Protocol servers
- **Permission system** — configurable allow/deny/ask rules per tool and path
- **Hooks** — shell or HTTP hooks on session_start, pre/post tool calls, stop
- **Session persistence** — conversations saved to `~/.claude/sessions/`; resume with `claude resume`
- **Slash commands**: `/help`, `/clear`, `/model [provider/model]`, `/cost`, `/compact`
- **Rich TUI** — scrollable history, Markdown rendering, syntax highlighting, mouse support
- **Config file** — project-local `.claude/settings.json` or `~/.claude/settings.json`
- **Automatic retry** — exponential backoff with Retry-After support for 429/5xx errors

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
| `WebSearch` | Search the web via DuckDuckGo (no API key needed) |

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

## Multi-provider support

In addition to Anthropic, any OpenAI-compatible or Anthropic-compatible provider is supported via `--provider` / `--base-url`.

| Provider | `--provider` | Protocol | Default model |
|----------|-------------|----------|---------------|
| Anthropic | `anthropic` (default) | Anthropic | claude-sonnet-4-6 |
| Codex | `codex` | OpenAI | (user-specified) |
| Kimi (Moonshot) | `kimi` | OpenAI | moonshot-v1-8k |
| OpenAI | `openai` | OpenAI | gpt-4o |
| DeepSeek | `deepseek` | OpenAI | deepseek-chat |
| 通义千问 | `qwen` | OpenAI | qwen-plus |
| 字节 Ark (OpenAI) | `ark` | OpenAI | ark-code-latest |
| 字节 Ark (Anthropic) | `ark-anthropic` | Anthropic | ark-code-latest |

**Codex 示例：**

```bash
# Via environment variables
export CODEX_BASE_URL="https://coder.api.visioncoder.cn/v1"
export CODEX_API_KEY="your-codex-api-key"
claude --provider codex --model your-model

# Via CLI flags
claude --provider codex --api-key "your-key" --base-url "https://coder.api.visioncoder.cn/v1" --model your-model
```

**Runtime model switching:**

```
/model codex/o3          # Switch to codex provider with o3 model
/model deepseek/deepseek-chat  # Switch to deepseek
/model gpt-4o            # Switch model only (keep current provider)
/model                   # Show current provider/model
```

**字节 Ark 示例：**

```bash
# OpenAI-compatible endpoint
claude --provider ark --model ark-code-latest --api-key $ARK_API_KEY

# Anthropic-compatible endpoint
claude --provider ark-anthropic --model ark-code-latest --api-key $ARK_API_KEY
```

**settings.json 配置：**

```json
{
  "provider": "ark",
  "model": "ark-code-latest",
  "apiKey": "your-ark-api-key"
}
```

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
