# Provider 支持

## 支持的 Provider

| Provider | 协议 | 环境变量 |
|----------|------|----------|
| anthropic | Anthropic | `ANTHROPIC_API_KEY` |
| codex | OpenAI | `CODEX_API_KEY`, `CODEX_BASE_URL` |
| kimi | OpenAI | `KIMI_API_KEY` |
| openai | OpenAI | `OPENAI_API_KEY` |
| deepseek | OpenAI | `DEEPSEEK_API_KEY` |
| qwen | OpenAI | `QWEN_API_KEY` |
| ark | OpenAI | `ARK_API_KEY` |
| ark-anthropic | Anthropic | `ARK_API_KEY` |

## 新增 Provider

### 步骤

1. 在 `internal/tui/model.go` 的 `knownProviders` 中添加条目
2. （可选）在 `internal/api/capabilities.go` 中声明能力
3. 更新 `README.md` 文档

### 示例

在 `internal/tui/model.go` 中：

```go
var knownProviders = map[string]providerInfo{
    "anthropic": {"https://api.anthropic.com", "anthropic"},
    "codex":     {"https://coder.api.visioncoder.cn/v1", "openai"},
    "myprovider": {"https://api.myprovider.com/v1", "openai"},
}
```

在 `internal/api/capabilities.go` 中声明能力：

```go
var DefaultCapabilities = map[string]Capabilities{
    "anthropic": {ToolUse: true, ParallelToolCalls: true, Vision: true, Streaming: true},
    "codex":     {ToolUse: true, ParallelToolCalls: true, Vision: false, Streaming: true},
    "myprovider": {ToolUse: true, ParallelToolCalls: false, Vision: false, Streaming: true},
}
```

## 使用方式

### 环境变量

```bash
export CODEX_API_KEY="your-key"
export CODEX_BASE_URL="https://coder.api.visioncoder.cn/v1"
claude --provider codex
```

### CLI 参数

```bash
claude --provider codex --api-key "your-key" --base-url "https://..."
```

### 配置文件

```json
{
  "provider": "codex",
  "apiKey": "your-key",
  "baseURL": "https://..."
}
```

## 运行时切换

使用 `/model` 命令：

```
/model codex/o3          # 切换到 codex provider 的 o3 模型
/model openai/gpt-4o     # 切换到 openai provider 的 gpt-4o 模型
/model gpt-4o            # 仅切换模型（保持当前 provider）
/model                   # 显示当前 provider/model
```

## Provider Capabilities

不同 provider 支持的特性不同：

| 特性 | 说明 |
|------|------|
| ToolUse | 支持工具调用 |
| ParallelToolCalls | 支持并行工具调用 |
| Vision | 支持图像输入 |
| Streaming | 支持流式输出 |

Agent 会根据 provider 能力自动降级功能。
