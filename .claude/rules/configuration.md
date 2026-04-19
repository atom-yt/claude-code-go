# 配置管理

## 配置优先级

配置按以下优先级加载（高到低）：

1. **CLI flags**：`--model`, `--api-key`, `--provider`, etc.
2. **环境变量**：`ANTHROPIC_API_KEY`, `CLAUDE_PROVIDER`, etc.
3. **项目设置**：`.claude/settings.json`（当前目录或父目录）
4. **用户设置**：`~/.claude/settings.json`

## 配置文件格式

```json
{
  "model": "claude-sonnet-4-6",
  "provider": "anthropic",
  "apiKey": "sk-ant-...",
  "baseURL": "",
  "permissions": {
    "defaultMode": "default",
    "allow": [
      { "tool": "Read" },
      { "tool": "Glob" }
    ],
    "deny": [
      { "tool": "Bash", "command": "rm -rf" }
    ],
    "ask": [
      { "tool": "Write" }
    ]
  },
  "hooks": {
    "pre_tool_call": [
      {
        "matcher": "Write",
        "hooks": [
          { "type": "shell", "command": "echo \"Writing file\" >&2" }
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

## 环境变量

| 变量名 | 说明 |
|--------|------|
| `ANTHROPIC_API_KEY` | Anthropic API Key |
| `CLAUDE_PROVIDER` | 默认 provider |
| `CLAUDE_MODEL` | 默认模型 |
| `CLAUDE_BASE_URL` | API Base URL |
| `CODEX_API_KEY` | Codex API Key |
| `CODEX_BASE_URL` | Codex Base URL |

## CLI 参数

| 参数 | 说明 |
|------|------|
| `--provider` | Provider 名称 |
| `--model` | 模型名称 |
| `--api-key` | API Key |
| `--base-url` | API Base URL |
| `--verbose` | 详细输出 |

## 配置加载

配置加载逻辑在 `internal/config/settings.go` 中实现：

1. 从命令行 flags 读取
2. 从环境变量读取（覆盖 CLI flags）
3. 从项目 `.claude/settings.json` 读取（覆盖环境变量）
4. 从用户 `~/.claude/settings.json` 读取（覆盖项目设置）

最终配置是所有来源的合并结果，优先级高的覆盖优先级低的。
