# Claude Code Go — 开发计划

> 分阶段实现 Go 版本的 Claude Code CLI 工具。

---

## 阶段划分总览

```
Phase 1: 基础骨架 (1-2周)
    ↓
Phase 2: API 客户端 + Agent 循环 (1-2周)
    ↓
Phase 3: 核心工具集 (2周)
    ↓
Phase 4: 权限系统 + 配置 (1周)
    ↓
Phase 5: TUI 完善 (1周)
    ↓
Phase 6: Hooks + MCP (2-3周)
    ↓
Phase 7: 命令系统 + 会话持久化 (1-2周)
    ↓
Phase 8: 测试 + 打磨 (持续)
```

---

## Phase 1：项目骨架

**目标**：建立项目结构，能编译运行，打印 hello world。

### 任务

- [ ] **1.1** 初始化 Go module
  ```bash
  go mod init github.com/atom-yt/claude-code-go
  ```

- [ ] **1.2** 引入基础依赖
  - `github.com/spf13/cobra` — CLI 框架
  - `github.com/charmbracelet/bubbletea` — TUI
  - `github.com/charmbracelet/lipgloss` — 样式
  - `github.com/charmbracelet/glamour` — Markdown 渲染

- [ ] **1.3** 实现 `cmd/claude/main.go`
  - `cobra` 根命令：`claude [flags] [prompt]`
  - 子命令：`claude resume`、`claude version`
  - 全局 flag：`--model`、`--api-key`、`--verbose`

- [ ] **1.4** 实现空壳 TUI（`internal/tui/`）
  - `bubbletea` 基础 Model/View/Update
  - 可输入文本并回显
  - Ctrl+C 退出

- [ ] **1.5** 项目目录结构按 DESIGN.md 创建

**验收标准**：`go build ./cmd/claude && ./claude` 可以启动 TUI，输入文字可以回显，Ctrl+C 退出。

---

## Phase 2：Anthropic API 客户端 + Agent 基础循环

**目标**：能和 Claude API 进行基础流式对话，看到实时输出。

### 任务

- [ ] **2.1** 实现 API 消息类型（`internal/api/types.go`）
  - `Message`、`ContentBlock`、`ToolUse`、`ToolResult`
  - `MessagesRequest`、`MessagesResponse`
  - `Usage`、`StopReason`

- [ ] **2.2** 实现 SSE 流式客户端（`internal/api/stream.go`）
  - HTTP POST 到 `https://api.anthropic.com/v1/messages`
  - 解析 SSE：`data: {...}` 格式
  - 处理事件类型：
    - `message_start`
    - `content_block_start` / `content_block_delta` / `content_block_stop`
    - `message_delta` / `message_stop`
    - `error`
  - 返回 `<-chan APIEvent` channel

- [ ] **2.3** 实现基础 Agent 循环（`internal/agent/agent.go`）
  - 接收用户 message
  - 调用 API，收集流式输出
  - 返回 `<-chan StreamEvent`（目前只支持文本，无工具调用）

- [ ] **2.4** 将 Agent 接入 TUI
  - 用户按 Enter 发送消息
  - Agent 流式响应实时显示
  - 显示 "thinking..." 状态

- [ ] **2.5** 配置基础加载
  - 从环境变量读取 `ANTHROPIC_API_KEY`
  - 从 `~/.claude/settings.json` 读取 `model` 配置

**验收标准**：`./claude "hello"` 或交互模式输入消息，能看到 Claude 的流式回复。

---

## Phase 3：核心工具系统

**目标**：Claude 能通过工具调用执行文件和命令操作。

### 任务

- [ ] **3.1** 定义 Tool 接口（`internal/tools/tool.go`）
  ```go
  type Tool interface {
      Name() string
      Description(input map[string]any) string
      InputSchema() map[string]any
      Call(ctx context.Context, input map[string]any, tctx *ToolUseContext) (ToolResult, error)
      IsReadOnly() bool
      IsConcurrencySafe() bool
  }
  ```

- [ ] **3.2** 实现工具注册表（`internal/tools/registry.go`）
  - `Register(tool Tool)`
  - `GetAll() []Tool`
  - `GetByName(name string) (Tool, bool)`
  - `ToAPISpecs() []ToolSpec`（供 API 调用时传递）

- [ ] **3.3** 实现 `Read` 工具（`internal/tools/read/`）
  - 读取文件内容
  - 支持行范围（offset/limit）
  - 大文件截断提示
  - 返回带行号的内容

- [ ] **3.4** 实现 `Write` 工具（`internal/tools/write/`）
  - 写入/创建文件
  - 支持创建父目录
  - 权限检查（需要 write 权限）

- [ ] **3.5** 实现 `Edit` 工具（`internal/tools/edit/`）
  - 精确字符串替换（`old_string` → `new_string`）
  - `replace_all` 选项
  - 唯一性检查（防止错误替换）

- [ ] **3.6** 实现 `Bash` 工具（`internal/tools/bash/`）
  - 执行 shell 命令
  - 超时控制（默认 120s）
  - 捕获 stdout/stderr
  - 工作目录跟随

- [ ] **3.7** 实现 `Glob` 工具（`internal/tools/glob/`）
  - 文件名模式匹配（`**/*.go` 等）
  - 按修改时间排序
  - 结果数量限制

- [ ] **3.8** 实现 `Grep` 工具（`internal/tools/grep/`）
  - 基于 ripgrep 或标准库
  - 支持正则表达式
  - 文件类型过滤（`--type`）
  - 上下文行（`-C`）
  - 输出模式：files/content/count

- [ ] **3.9** 将工具调用接入 Agent 循环
  - 解析 API 响应中的 `tool_use` 事件
  - 根据 tool name 找到 tool 并执行
  - 将 `tool_result` 追加到消息历史
  - 继续调用 API（循环）

- [ ] **3.10** TUI 中显示工具调用状态
  - "Running bash: git status..."
  - 工具结果折叠显示

**验收标准**：`./claude "list all .go files and show me main.go"` 能自动调用 Glob + Read 工具完成任务。

---

## Phase 4：权限系统 + 配置管理

**目标**：完善权限控制，防止未授权操作。

### 任务

- [ ] **4.1** 实现权限类型（`internal/permissions/types.go`）
  - `PermissionMode`：default / trust-all / manual
  - `Rule`：tool、path、command 匹配
  - `Decision`：allowed/denied + reason

- [ ] **4.2** 实现规则匹配器（`internal/permissions/rules.go`）
  - 解析 `"Bash(git *)"` 格式的规则
  - 路径 glob 匹配（`*.ts`、`src/**`）
  - 命令前缀匹配

- [ ] **4.3** 实现权限检查器（`internal/permissions/checker.go`）
  - 按优先级：deny > allow > ask
  - `trust-all` 模式跳过检查
  - 向 TUI 发起询问（Ask 模式）

- [ ] **4.4** 完善配置加载（`internal/config/settings.go`）
  - 加载 `~/.claude/settings.json`
  - 加载项目级 `.claude/settings.json`（当前目录向上查找）
  - 合并策略（项目覆盖用户）
  - 环境变量覆盖
  - CLI flag 最高优先级

- [ ] **4.5** 配置文件 JSON Schema 定义
  - 权限配置：`permissions.allow/deny/ask`
  - 模型配置：`model`
  - 环境变量注入：`env`

- [ ] **4.6** 在 Agent 循环中集成权限检查
  - 每次工具调用前调用 `checker.Check()`
  - 拒绝时返回错误给 Claude（不执行工具）
  - Ask 时暂停等待用户确认

**验收标准**：配置 `deny: [{tool: "Bash"}]` 后，Bash 工具调用被拒绝；trust-all 模式下所有工具自动通过。

---

## Phase 5：TUI 完善

**目标**：提供完整的终端交互体验。

### 任务

- [ ] **5.1** 多行输入支持
  - `Enter` 发送，`Shift+Enter` 换行
  - 或可配置的提交快捷键

- [ ] **5.2** 对话历史滚动
  - 鼠标滚轮 / 键盘滚动
  - 自动滚动到最新消息
  - 历史消息折叠

- [ ] **5.3** 工具调用可视化
  - 工具名称 + 参数摘要
  - 执行状态（进行中/完成/失败）
  - 结果折叠展开

- [ ] **5.4** Markdown 渲染
  - 使用 `glamour` 渲染 Claude 的 Markdown 回复
  - 代码块语法高亮
  - 链接处理

- [ ] **5.5** 状态栏
  - 当前模型
  - token 用量
  - 当前工作目录

- [ ] **5.6** 键盘快捷键
  - `Ctrl+C`：取消当前请求或退出
  - `Ctrl+L`：清屏
  - `↑/↓`：历史输入导航

- [ ] **5.7** 流式输出动画
  - 打字效果（光标闪烁）
  - 思考中的旋转动画

**验收标准**：TUI 体验流畅，支持长对话滚动、Markdown 渲染、工具状态显示。

---

## Phase 6：Hooks 系统

**目标**：支持生命周期拦截，允许用户自定义行为。

### 任务

- [ ] **6.1** 实现 Hook 类型（`internal/hooks/types.go`）
  - 事件：`session_start`、`pre_tool_call`、`post_tool_call`、`stop`
  - 命令类型：`command`（shell）、`http`（HTTP POST）

- [ ] **6.2** 实现 Shell Hook 执行器
  - 执行 shell 命令，传入 JSON 环境变量
  - 超时控制
  - 捕获 stdout 作为 hook 输出

- [ ] **6.3** 实现 HTTP Hook 执行器
  - POST JSON 到指定 URL
  - 支持自定义 headers
  - 解析响应决策（allow/deny）

- [ ] **6.4** 实现 Hook 匹配器
  - 按 tool name 匹配 `matcher` 字段
  - 支持通配符

- [ ] **6.5** 在 Agent 循环中触发 Hooks
  - `pre_tool_call`：工具执行前，可拦截（deny）
  - `post_tool_call`：工具执行后，可修改结果
  - `session_start`：会话开始时
  - `stop`：Agent 停止时

- [ ] **6.6** 从配置加载 Hooks
  ```json
  {
    "hooks": {
      "pre_tool_call": [
        { "matcher": "Bash", "hooks": [{ "type": "command", "command": "echo $TOOL_INPUT" }] }
      ]
    }
  }
  ```

**验收标准**：配置 pre_tool_call hook 后，每次工具调用前 shell 命令被执行。

---

## Phase 7：MCP 客户端集成

**目标**：支持 MCP 服务器，扩展工具能力。

### 任务

- [ ] **7.1** 实现 MCP 协议类型（`internal/mcp/types.go`）
  - `initialize` / `tools/list` / `tools/call`
  - `resources/list` / `resources/read`
  - JSON-RPC 2.0 消息格式

- [ ] **7.2** 实现 stdio transport（`internal/mcp/stdio.go`）
  - 启动子进程
  - 通过 stdin/stdout 进行 JSON-RPC 通信
  - 进程生命周期管理

- [ ] **7.3** 实现 SSE transport（`internal/mcp/sse.go`）
  - HTTP SSE 连接
  - 支持 HTTP headers 认证

- [ ] **7.4** 实现 MCP 客户端（`internal/mcp/client.go`）
  - `Connect()` 建立连接
  - `ListTools()` 获取工具列表
  - `CallTool(name, input)` 执行工具
  - 将 MCP tools 转换为 `Tool` 接口

- [ ] **7.5** 在启动时加载 MCP 服务器
  - 从配置 `mcpServers` 读取
  - 并发连接所有服务器
  - 注册 MCP 工具到 registry（前缀 `mcp__<server>__`）

- [ ] **7.6** 处理 MCP 工具调用
  - 通过对应 MCP 客户端执行
  - 错误处理和超时

**验收标准**：配置 MCP server 后，MCP 工具出现在 Claude 可用工具列表中，可被调用。

---

## Phase 8：斜杠命令 + 会话持久化

**目标**：支持内置命令和会话恢复。

### 任务

- [ ] **8.1** 实现命令接口（`internal/commands/command.go`）
  ```go
  type Command interface {
      Name() string
      Description() string
      Execute(ctx context.Context, args []string, tctx *ToolUseContext) error
  }
  ```

- [ ] **8.2** 实现内置命令
  - `/help` — 显示帮助
  - `/clear` — 清空对话历史
  - `/compact` — 压缩历史（摘要）
  - `/model <name>` — 切换模型
  - `/verbose` — 切换详细模式
  - `/cost` — 显示 token 用量和费用

- [ ] **8.3** TUI 中解析斜杠命令
  - 输入以 `/` 开头时识别为命令
  - 命令补全提示

- [ ] **8.4** 会话持久化（`internal/session/persist.go`）
  - 保存对话历史到 `~/.claude/sessions/<id>.json`
  - 记录 session ID、时间戳、模型、cost
  - 保存 AppState

- [ ] **8.5** 会话恢复（`claude resume`）
  - 列出最近会话
  - 加载选定会话的消息历史
  - 继续对话

- [ ] **8.6** 历史压缩（`/compact`）
  - 调用 Claude API 生成历史摘要
  - 用摘要替换旧消息
  - 保留最近 N 条消息

**验收标准**：`/clear` 清空历史；`claude resume` 恢复上次对话。

---

## Phase 9：测试与打磨

**目标**：确保稳定性，完善文档。

### 任务

- [ ] **9.1** API 客户端单元测试
  - SSE 解析测试
  - 错误处理测试
  - Mock HTTP server

- [ ] **9.2** 工具单元测试
  - 各工具 happy path 测试
  - 边界条件（大文件、权限拒绝等）

- [ ] **9.3** 权限系统测试
  - 规则匹配测试
  - 各种模式下的决策测试

- [ ] **9.4** Agent 循环集成测试
  - Mock API 返回工具调用
  - 验证工具被正确执行和结果被正确传递

- [ ] **9.5** 性能优化
  - 启动时间优化
  - 大消息历史处理
  - 并发工具执行

- [ ] **9.6** 文档完善
  - README.md
  - 安装说明
  - 配置参考
  - 工具列表

- [ ] **9.7** 构建和发布
  - Makefile
  - 多平台交叉编译（darwin/linux/windows）
  - GitHub Actions CI

---

## 里程碑

| 里程碑 | 目标 | 预计完成 |
|--------|------|---------|
| M1 | Phase 1+2 完成：基础对话可用 | Week 3 |
| M2 | Phase 3+4 完成：工具调用 + 权限控制 | Week 6 |
| M3 | Phase 5 完成：TUI 体验完善 | Week 7 |
| M4 | Phase 6+7 完成：Hooks + MCP | Week 10 |
| M5 | Phase 8+9 完成：完整功能 + 稳定 | Week 12 |

---

## 技术债务和风险

| 风险 | 说明 | 缓解方案 |
|------|------|---------|
| SSE 解析兼容性 | Anthropic API SSE 格式可能变化 | 充分测试，参考官方 SDK |
| bubbletea 学习曲线 | Elm 架构对 Go 开发者较陌生 | 从简单示例开始 |
| 并发工具执行 | 并发执行可能引发竞态 | 仔细设计锁和 channel |
| MCP 协议合规性 | MCP spec 还在演进 | 参考官方 TypeScript SDK |
| 权限 DSL 复杂性 | 原版规则语法复杂 | 先实现简化版，逐步完善 |

---

## 开发规范

### 代码规范
- `gofmt` 格式化
- `golint` + `staticcheck` 静态检查
- 错误必须处理（不得 `_` 忽略关键错误）
- 公开接口必须有文档注释

### Git 规范
- feature 分支开发
- commit message：`feat:`, `fix:`, `refactor:`, `test:`, `docs:`
- PR 前通过所有测试

### 测试规范
- 核心逻辑覆盖率 > 80%
- 集成测试使用 mock API server
- 工具测试使用临时目录
