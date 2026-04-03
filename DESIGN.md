# Claude Code Go — 设计方案

> 基于原 TypeScript 版本 [claude-code](../claude-code/) 的 Golang 重实现。

---

## 一、项目概述

Claude Code 是一个 AI 驱动的终端 CLI 工具，核心能力：

- 与 Claude API 进行流式对话
- 通过工具（Tool）调用执行本地操作（读写文件、运行命令、搜索等）
- 支持多轮 Agent 循环（工具调用 → 结果反馈 → 继续推理）
- 支持 MCP（Model Context Protocol）服务器扩展
- Hooks 系统（pre/post tool call 等生命周期拦截）
- 权限控制（每个工具调用都需权限检查）
- 会话持久化与恢复
- 斜杠命令（/help, /clear, /compact 等）

---

## 二、原版架构分析

### 2.1 整体架构

```
User Input
    │
    ▼
CLI Parser (Commander.js)
    │
    ▼
┌─────────────────────────────────┐
│          Agent Loop             │
│  query.ts (1729 lines)          │
│                                 │
│  1. Build messages              │
│  2. Stream API call             │
│  3. Parse stream events         │
│  4. Execute tool calls          │
│  5. Collect tool results        │
│  6. Loop until done             │
└─────────────────────────────────┘
    │               │
    ▼               ▼
Claude API      Tool Executor
(Anthropic)     (45+ tools)
    │               │
    ▼               ▼
Stream Events   Tool Results
(SSE/HTTP)      (file/bash/web...)
    │
    ▼
Ink React TUI (Terminal UI)
```

### 2.2 核心模块

| 模块 | TS 实现 | 职责 |
|------|---------|------|
| CLI 入口 | `entrypoints/cli.tsx` | 启动、参数解析 |
| Agent 主循环 | `query.ts` | 流式 API 调用、工具调度 |
| 工具系统 | `tools/`, `Tool.ts` | 45+ 内置工具定义与注册 |
| 权限系统 | `utils/permissions/` | Allow/Deny/Ask 规则匹配 |
| Hooks 系统 | `utils/hooks.ts` | 生命周期拦截执行 |
| MCP 集成 | `services/mcp/` | MCP 协议客户端 |
| 配置管理 | `utils/settings/` | JSON 配置、MDM、合并 |
| 会话管理 | `state/AppStateStore.ts` | 全局状态、会话持久化 |
| TUI 渲染 | `ink/`, `components/` | React + Ink 终端渲染 |
| 命令系统 | `commands/` | 100+ 斜杠命令 |

### 2.3 Agent 循环核心流程

```
┌──────────────────────────────────────────┐
│              query() generator            │
│                                          │
│  messages = prepareMessages()            │
│                                          │
│  loop:                                   │
│    response = streamAPI(messages)        │
│    for event in response.stream:         │
│      if text_delta: yield text event     │
│      if tool_use:                        │
│        result = executeTool(tool, args)  │
│        messages.append(tool_result)      │
│    if stop_reason != 'tool_use': break   │
│                                          │
│  yield stop event                        │
└──────────────────────────────────────────┘
```

---

## 三、Go 版本设计

### 3.1 设计原则

1. **忠实核心逻辑**：忠实还原 Agent 循环、工具系统、权限系统、Hooks 等核心逻辑
2. **Go 惯用模式**：用 channel + goroutine 替代 AsyncGenerator，用 interface 替代 TypeScript 的 type union
3. **模块化边界清晰**：每个包只暴露必要接口，内部实现可替换
4. **可测试性优先**：核心业务逻辑与 I/O 分离，便于单元测试
5. **渐进式实现**：先实现 MVP（基础对话 + 核心工具），再扩展高级特性

### 3.2 目录结构

```
claude-code-go/
├── cmd/
│   └── claude/
│       └── main.go              # CLI 入口
├── internal/
│   ├── agent/
│   │   ├── agent.go             # Agent 主循环
│   │   ├── stream.go            # 流式处理
│   │   └── context.go           # ToolUseContext 定义
│   ├── api/
│   │   ├── client.go            # Anthropic API 客户端
│   │   ├── types.go             # API 消息类型
│   │   └── stream.go            # SSE 流式解析
│   ├── tools/
│   │   ├── registry.go          # 工具注册表
│   │   ├── tool.go              # Tool 接口定义
│   │   ├── bash/                # Bash 工具
│   │   ├── read/                # 文件读取
│   │   ├── write/               # 文件写入
│   │   ├── edit/                # 文件编辑
│   │   ├── glob/                # 文件搜索
│   │   ├── grep/                # 内容搜索
│   │   ├── webfetch/            # HTTP 请求
│   │   └── ...                  # 其他工具
│   ├── permissions/
│   │   ├── checker.go           # 权限检查
│   │   ├── rules.go             # 规则匹配
│   │   └── types.go             # 权限类型
│   ├── hooks/
│   │   ├── executor.go          # Hook 执行器
│   │   ├── types.go             # Hook 类型
│   │   └── runner.go            # 各类 Hook 运行
│   ├── mcp/
│   │   ├── client.go            # MCP 客户端
│   │   ├── stdio.go             # stdio transport
│   │   ├── sse.go               # SSE transport
│   │   └── types.go             # MCP 协议类型
│   ├── config/
│   │   ├── settings.go          # 配置加载
│   │   ├── types.go             # 配置类型
│   │   └── merge.go             # 配置合并
│   ├── session/
│   │   ├── store.go             # 会话存储
│   │   ├── persist.go           # 持久化
│   │   └── types.go             # 会话类型
│   ├── messages/
│   │   ├── types.go             # 消息类型
│   │   ├── normalize.go         # 消息规范化
│   │   └── compact.go           # 历史压缩
│   ├── commands/
│   │   ├── registry.go          # 命令注册
│   │   ├── command.go           # Command 接口
│   │   ├── help.go              # /help
│   │   ├── clear.go             # /clear
│   │   ├── compact.go           # /compact
│   │   └── ...                  # 其他命令
│   └── tui/
│       ├── app.go               # TUI 主界面
│       ├── input.go             # 输入处理
│       ├── render.go            # 输出渲染
│       └── style.go             # 样式定义
├── pkg/
│   └── anthropic/               # 可复用的 Anthropic SDK
│       ├── client.go
│       ├── messages.go
│       └── stream.go
├── go.mod
├── go.sum
├── DESIGN.md                    # 本文件
└── PLAN.md                      # 开发计划
```

### 3.3 核心接口设计

#### Tool 接口

```go
// internal/tools/tool.go

type ToolResult struct {
    Data     any
    Messages []Message
    Error    error
}

type ToolUseContext struct {
    Messages    []Message
    AppState    *AppState
    Tools       []Tool
    Commands    []Command
    Permissions *PermissionContext
    MCPClients  []MCPClient
    CWD         string
    // OnProgress callback
    OnProgress  func(data any)
}

type Tool interface {
    Name() string
    Description(input map[string]any) string
    InputSchema() map[string]any   // JSON Schema
    Call(ctx context.Context, input map[string]any, tctx *ToolUseContext) (ToolResult, error)
    IsReadOnly() bool
    IsConcurrencySafe() bool
}
```

#### Agent 主循环

```go
// internal/agent/agent.go

type StreamEvent struct {
    Type    string // "text", "tool_use", "tool_result", "stop", "error"
    Text    string
    Tool    *ToolUse
    Result  *ToolResult
    Error   error
}

type Agent struct {
    client    *api.Client
    tools     []tools.Tool
    commands  []commands.Command
    config    *config.Settings
    hooks     *hooks.Executor
    perms     *permissions.Checker
}

// Query 执行一次 Agent 查询，通过 channel 流式返回事件
func (a *Agent) Query(ctx context.Context, messages []Message) <-chan StreamEvent {
    ch := make(chan StreamEvent)
    go a.run(ctx, messages, ch)
    return ch
}

func (a *Agent) run(ctx context.Context, messages []Message, ch chan<- StreamEvent) {
    defer close(ch)
    for {
        // 1. 调用 API
        stream := a.client.StreamMessages(ctx, messages, a.buildToolSpecs())
        
        // 2. 处理流式事件
        var toolUses []ToolUse
        for event := range stream {
            switch event.Type {
            case "text_delta":
                ch <- StreamEvent{Type: "text", Text: event.Text}
            case "tool_use":
                toolUses = append(toolUses, event.ToolUse)
            }
        }
        
        // 3. 如果没有工具调用则退出
        if len(toolUses) == 0 {
            ch <- StreamEvent{Type: "stop"}
            return
        }
        
        // 4. 执行工具（可并发）
        results := a.executeTools(ctx, toolUses)
        
        // 5. 追加工具结果到消息
        messages = append(messages, buildToolResultMessage(results))
    }
}
```

#### 权限系统

```go
// internal/permissions/checker.go

type PermissionMode string
const (
    ModeDefault  PermissionMode = "default"
    ModeTrustAll PermissionMode = "trust-all"
    ModeManual   PermissionMode = "manual"
)

type Rule struct {
    Tool    string // e.g. "Bash", "Write"
    Path    string // file path pattern
    Command string // shell command pattern
}

type Checker struct {
    Mode      PermissionMode
    AllowRules []Rule
    DenyRules  []Rule
    AskRules   []Rule
    // AskFn 由 TUI 实现，用于向用户询问
    AskFn     func(ctx context.Context, req AskRequest) (bool, error)
}

type Decision struct {
    Allowed bool
    Reason  string
}

func (c *Checker) Check(ctx context.Context, toolName string, input map[string]any) (Decision, error)
```

#### Hooks 系统

```go
// internal/hooks/types.go

type HookEvent string
const (
    HookPreToolCall  HookEvent = "pre_tool_call"
    HookPostToolCall HookEvent = "post_tool_call"
    HookSessionStart HookEvent = "session_start"
    HookStop         HookEvent = "stop"
    // ...
)

type HookCommand struct {
    Type    string // "command", "http", "prompt"
    Command string // shell command
    URL     string // for http type
    Prompt  string // for prompt type
    Timeout int    // seconds
}

type HookMatcher struct {
    Matcher string       // e.g. "Write", "Read(*.ts)"
    Hooks   []HookCommand
}

type HookInput struct {
    Event     HookEvent
    ToolName  string
    ToolInput map[string]any
    SessionID string
}

type HookResult struct {
    Decision string // "allow", "deny", "continue"
    Reason   string
    Output   string
}
```

#### 配置类型

```go
// internal/config/types.go

type Settings struct {
    Model              string
    MaxTokens          int
    APIKey             string
    APIBaseURL         string
    Permissions        PermissionsConfig
    Hooks              map[string][]HookMatcher
    MCPServers         map[string]MCPServerConfig
    EnvironmentVars    map[string]string
    // ...
}

type MCPServerConfig struct {
    Type    string            // "stdio", "sse", "http"
    Command string            // for stdio
    Args    []string          // for stdio
    URL     string            // for sse/http
    Env     map[string]string
    Headers map[string]string
}
```

### 3.4 流式处理设计

Go 使用 channel 替代 TypeScript 的 AsyncGenerator：

```go
// internal/api/stream.go

type APIEvent struct {
    Type    string
    // text_delta
    Text    string
    // tool_use
    ToolUse *ToolUse
    // message_stop
    StopReason string
    // error
    Error  error
    Usage  *Usage
}

// StreamMessages 返回事件 channel
func (c *Client) StreamMessages(
    ctx context.Context,
    req MessagesRequest,
) <-chan APIEvent {
    ch := make(chan APIEvent, 32)
    go c.streamSSE(ctx, req, ch)
    return ch
}
```

### 3.5 TUI 设计

使用 [bubbletea](https://github.com/charmbracelet/bubbletea) 构建终端 UI：

```
┌─────────────────────────────────┐
│  Claude Code (Go)               │
│─────────────────────────────────│
│  [conversation history]         │
│  User: explain this file        │
│  Claude: [streaming response]   │
│  > Reading file.go...           │
│  > Running bash...              │
│─────────────────────────────────│
│  Input: █                       │
│  [Ctrl+C: exit] [Enter: send]   │
└─────────────────────────────────┘
```

- `bubbletea` 处理事件循环和键盘输入
- `lipgloss` 处理样式和颜色
- `glamour` 渲染 Markdown 输出

### 3.6 并发模型

```
Main Goroutine (TUI)
    │
    ├── Agent Goroutine
    │       │
    │       ├── API Stream Goroutine
    │       │
    │       └── Tool Executor Pool
    │               ├── Tool 1 Goroutine
    │               ├── Tool 2 Goroutine
    │               └── ...
    │
    └── Hook Goroutine Pool
```

- Agent 循环运行在独立 goroutine
- 并发安全工具（`isConcurrencySafe`）并行执行
- Hooks 异步执行，不阻塞主流程
- 通过 `context.Context` 传播取消信号

---

## 四、技术选型

| 功能 | 选型 | 说明 |
|------|------|------|
| CLI 参数解析 | [cobra](https://github.com/spf13/cobra) | 成熟的 Go CLI 框架 |
| TUI 框架 | [bubbletea](https://github.com/charmbracelet/bubbletea) | Elm 架构的终端 UI |
| 终端样式 | [lipgloss](https://github.com/charmbracelet/lipgloss) | 终端颜色和布局 |
| Markdown 渲染 | [glamour](https://github.com/charmbracelet/glamour) | 终端 Markdown 渲染 |
| JSON Schema 验证 | [jsonschema](https://github.com/santhosh-tekuri/jsonschema) | JSON Schema 校验 |
| HTTP 客户端 | 标准库 `net/http` | 支持 SSE 流式 |
| 配置文件 | [viper](https://github.com/spf13/viper) | 多层配置合并 |
| 日志 | [slog](https://pkg.go.dev/log/slog) | Go 1.21 标准库 |
| 测试 | `testing` + [testify](https://github.com/stretchr/testify) | 标准测试框架 |
| MCP 协议 | 自实现 | 基于官方 spec |

---

## 五、MVP 范围

第一版（MVP）聚焦核心能力，不追求功能完整性：

### 包含
- [x] Claude API 流式调用（SSE）
- [x] 基础消息对话循环
- [x] Agent 工具调用循环
- [x] 核心工具：Bash、Read、Write、Edit、Glob、Grep
- [x] 基础权限系统（allow/deny/ask）
- [x] JSON 配置加载（`~/.claude/settings.json`）
- [x] 基础 TUI（输入框 + 输出区域）
- [x] 会话历史显示

### 不包含（后续版本）
- MCP 服务器集成
- Hooks 系统
- 斜杠命令系统
- 会话持久化与恢复
- 自动历史压缩（compact）
- WebSearch/WebFetch 工具
- 多 Agent 协调

---

## 六、与原版主要差异

| 方面 | TypeScript 版 | Go 版 |
|------|--------------|-------|
| 异步模型 | AsyncGenerator | channel + goroutine |
| UI 框架 | React + Ink | bubbletea |
| Schema 校验 | Zod | jsonschema |
| 状态管理 | Zustand-like store | 结构体 + mutex |
| 类型系统 | 丰富的 Union Type | interface |
| 构建工具 | Bun | go build |
| 包管理 | npm | go modules |

---

## 七、关键技术挑战

1. **SSE 流式解析**：Anthropic API 返回 SSE 格式，需正确解析 `data:` 前缀、事件类型、JSON body
2. **工具并发执行**：标记为 `ConcurrencySafe` 的工具应并发执行，需协调结果收集
3. **权限 DSL 匹配**：`"Bash(git *)"` 这类规则需要解析和匹配
4. **流式 TUI 更新**：API 流式输出需实时更新终端，不能等全部完成
5. **上下文取消**：用户 Ctrl+C 需要中断正在执行的工具和 API 请求
6. **错误边界**：工具执行失败不应崩溃整个 agent 循环
