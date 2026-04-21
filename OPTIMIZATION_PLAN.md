# claude-code-go 对标分析与优化计划

> 对比对象：
> - `../claude-code`
> - `../hermes-agent`
> - `../OpenHarness`
> - 当前仓库 `claude-code-go`

## 1. 结论先行

`claude-code-go` 现在更像一个“可用的单机版 agent loop + TUI 原型”，而不是一个成熟的 agent harness。

它已经具备这些核心基础：

- 流式多轮 agent loop
- 基础工具系统
- 权限检查与 hooks
- session 持久化
- MCP stdio 工具接入
- skills 扫描
- slash commands
- 基础 auto-compact / auto-dream 雏形

但和 `claude-code`、`hermes-agent`、`OpenHarness` 对比，当前短板不是“工具数量少”这么简单，而是下面几个更关键的问题：

1. 很多能力已经有“tool 壳子”或“目录雏形”，但没有打通成完整产品闭环。
2. 系统提示词、上下文注入、计划模式、任务管理、记忆系统都还偏浅，缺少稳定的 runtime contract。
3. 安全与执行边界仍较粗糙，尤其是 bash/path/web/mcp 侧。
4. 缺少 coordinator / subagent / background task 这一层，导致复杂任务扩展性明显弱于另外三个项目。
5. 目前最应该优先补的是“架构完成度”，不是继续横向加更多工具。

建议把未来 1-2 个版本的目标定义为：

- 先把 `claude-code-go` 做成一个扎实的本地单 agent harness
- 再引入受控的 plan/task/subagent 能力
- 最后再考虑插件生态、消息网关、多终端平台等外围扩张

---

## 2. 四个项目的定位差异

### 2.1 claude-code

定位是成熟的生产级 coding agent CLI。

强项：

- 工具体系完整，且不是简单堆工具，而是围绕工作流组织
- 有 plan mode、worktree、agent tool、task、LSP、MCP resource、plugin、remote/CCR 等完整运行时
- 安全与权限规则做得很细，尤其是 bash/path/sandbox
- TUI、命令、服务层、状态层边界清晰

对 `claude-code-go` 的启发：

- 不是补“更多命令”，而是补“运行时层”
- 先补 prompt/context orchestration、task lifecycle、tool execution governance

### 2.2 hermes-agent

定位是长期运行、跨平台、多环境、多技能、多记忆的 personal agent。

强项：

- messaging gateway 与长期运行能力强
- memory / skill / profile / model routing / environment backend 很完整
- 明显偏“agent platform”，不只是本地 coding CLI

对 `claude-code-go` 的启发：

- 记忆、技能、环境抽象值得借鉴
- 但 messaging/gateway/voice 不应成为当前阶段优先级

### 2.3 OpenHarness

定位是开源 agent harness 基础设施，强调可组合架构和 swarm。

强项：

- 模块边界清晰：engine / prompts / permissions / swarm / plugins / memory / channels
- 对 coordinator、task、plugin、auth/provider、channel 支撑比较系统
- 很适合作为“中间态架构”参考

对 `claude-code-go` 的启发：

- 最值得借鉴的是架构分层方式
- 先把 runtime 与 service 层拆清楚，再继续长功能

### 2.4 claude-code-go

定位目前是：

- Go 重写版本地 coding agent CLI
- 已有可运行体验
- 但系统层还偏薄，许多高级能力处在“半实现”状态

---

## 3. 当前 claude-code-go 的真实状态

下面这个判断基于代码而不是 README。

### 3.1 已经比较扎实的部分

- `internal/agent/agent.go`
  - agent loop 已能完成文本流式输出、tool_use 收集、权限检查、hooks、串并行工具执行、tool_result 回注
- `internal/api/`
  - 同时支持 Anthropic 风格和 OpenAI-compatible 风格的流式客户端
- `internal/tui/`
  - 基础 TUI 可用，包含 markdown 渲染、滚动、输入历史、autocomplete、状态栏
- `internal/permissions/`
  - allow / deny / ask 三层规则已成型
- `internal/hooks/`
  - shell / HTTP hooks 能工作
- `internal/session/`
  - session 落盘和恢复已经够用
- `internal/mcp/`
  - stdio MCP tools 已能注册进本地 registry

### 3.2 已实现但明显偏“原型”的部分

- `EnterPlanMode` / `ExitPlanMode`
  - 现在只是返回提示文本，没有真实 plan-mode 状态机
- `AskUserQuestion`
  - 现在只会格式化问题，不会真的驱动交互问答闭环
- `TodoWrite` / `Task*`
  - 都是进程内全局内存结构，不是持久任务系统，也没有 subagent/runtime 集成
- auto-compact
  - 现在本质是裁剪历史，不是高质量摘要压缩
- auto-dream
  - 触发条件、锁、prompt 有了，但记忆抽取/写回仍比较脆
- skills
  - 已支持扫描和注入，但还没有 skill workflow、依赖装载、插件化技能组合

### 3.3 明显缺失或薄弱的部分

- 没有稳定的 prompt builder / context builder
- 没有 `CLAUDE.md` / `AGENTS.md` / 项目上下文注入主链路
- 没有真正的 coordinator / subagent / background worker runtime
- 没有 sandbox abstraction
- MCP 只做到 stdio tool bridge，缺 resource/auth/http/reconnect
- provider 体系仍是硬编码 map，扩展性弱
- 安全策略没有细化到 destructive command、敏感路径、URL guard、path normalization 等层次
- session / memory 没有搜索、索引、摘要、回放能力
- plugin 目前没有真正的 runtime/plugin contract

---

## 4. 对标后的核心优化方向

## 4.1 优化方向一：先把“上下文系统”做完整

这是当前最值得优先补的设计层。

### 当前问题

`claude-code-go` 现在把大部分能力放在：

- model config
- tools
- slash commands
- TUI

但真正决定 agent 上限的是上下文装配能力，当前这里明显不够：

- 没有统一 system prompt builder
- 没有环境上下文、项目上下文、用户规则、技能上下文的分层注入
- 没有像 `claude-code` / `OpenHarness` 那样的 prompt/context service

### 对标参考

- `claude-code/services/*Prompt*`, `services/compact`, `services/toolUseSummary`
- `OpenHarness/src/openharness/prompts/*`
- `hermes-agent/agent/prompt_builder.py`

### 建议设计

引入独立的上下文装配层：

- `internal/prompt/system.go`
- `internal/prompt/context.go`
- `internal/prompt/project.go`
- `internal/prompt/environment.go`
- `internal/prompt/skills.go`

明确每次请求的上下文组成：

1. base system prompt
2. project instructions
3. runtime mode instructions
4. tool policy / permission hints
5. loaded skills
6. memory snippets
7. conversation history

### 直接收益

- 后续 plan mode、compact、dream、task、subagent 都有统一挂载点
- 可以把“功能”变成“模式”，而不是继续堆 command/tool

---

## 4.2 优化方向二：把 plan / task / subagent 从原型做成运行时

这是和另外三个项目差距最大的地方。

### 当前问题

当前仓库已经有：

- `EnterPlanMode`
- `ExitPlanMode`
- `TodoWrite`
- `TaskCreate/Update/List/...`

但它们还没有形成完整执行链：

- plan mode 没有独立状态
- task 没有 durable storage
- task 和 tool execution 没有调度关系
- 没有真正的 subagent spawn / wait / stop 机制

### 对标参考

- `claude-code/tasks/*`
- `claude-code/tools/AgentTool/*`
- `OpenHarness/src/openharness/swarm/*`
- `OpenHarness/src/openharness/tasks/*`
- `hermes-agent/tools/delegate_tool.py`

### 建议设计

优先做“轻量版”，不要一步学 Hermes 那种全平台代理系统。

建议路线：

1. 先做 plan session
   - 进入 plan mode 后，切换 prompt profile
   - 限制允许工具
   - 产出 plan artifact
   - 用户批准后再进入 implement mode

2. 再做 durable task store
   - 任务落盘到 session scope 或 workspace scope
   - task 状态与会话关联
   - TUI 能显示 current task / background task

3. 最后做本地 subagent runtime
   - 先支持 subprocess / goroutine worker 即可
   - 明确只允许只读 explorer 和有限写 worker 两类

### 直接收益

- 复杂任务不再全塞进单轮上下文
- 可以开始追平 `OpenHarness`/`claude-code` 的 coordinator 设计

---

## 4.3 优化方向三：重做 compact / memory，使其从“裁剪”升级为“压缩系统”

### 当前问题

当前 compact 主要是：

- 保留最近若干消息
- 删除旧消息

这不是严格意义上的 compact，更接近“截断”。

当前 dream 也存在类似问题：

- 有锁和阈值
- 有 consolidation prompt
- 但缺少稳定的 memory schema、索引、冲突合并策略

### 对标参考

- `claude-code/services/compact/*`
- `claude-code/services/autoDream/*`
- `claude-code/services/SessionMemory/*`
- `OpenHarness/src/openharness/memory/*`
- `hermes-agent/agent/memory_manager.py`

### 建议设计

把 compact 分为三层：

- micro-compact: 针对本轮或短段历史
- session-compact: 提取 session summary + open loops + pending tasks
- memory extract: 从 session summary 里抽 durable memory

memory 需要至少有三类对象：

- session summary
- project memory
- user/workspace memory

### 直接收益

- 长对话能力会明显提升
- dream 才真正有价值，不会只是“让模型再读一遍自己说过的话”

---

## 4.4 优化方向四：把权限与安全从“规则匹配”升级为“执行治理”

### 当前问题

当前权限系统有基础规则，但还远不够稳：

- `isReadOnlyTool()` 只硬编码了 `Read/Glob/Grep`
- `Bash` 没有 destructive command 语义层
- 文件路径没有更完整的 normalize / sensitive-path / workspace boundary 校验
- `WebFetch` / `WebSearch` 安全策略偏弱
- MCP 没有 server-level trust / capability gating

### 对标参考

- `claude-code/tools/BashTool/*`
- `claude-code/utils/sandbox/*`
- `OpenHarness/src/openharness/utils/network_guard.py`
- `OpenHarness/src/openharness/sandbox/*`
- `hermes-agent/tools/path_security.py`

### 建议设计

把安全层拆成三类：

- policy layer
  - allow/deny/ask
- validation layer
  - path normalization
  - command semantics
  - URL validation
  - sensitive path guard
- execution layer
  - sandbox backend
  - timeout
  - output truncation
  - provenance logging

优先补这些点：

- Bash command classifier
- 路径边界检查
- Web URL allow/deny rules
- MCP server trust level

### 直接收益

- 这是把 `claude-code-go` 从 demo 提升到“能放心跑”的关键

---

## 4.5 优化方向五：把 MCP 从“可调用工具桥”升级为“完整扩展协议层”

### 当前问题

现在 MCP 主要能力是：

- 连接 stdio server
- 列工具
- 注册为本地 tool

这离成熟实现还有明显差距：

- 没有 HTTP/SSE transport
- 没有 resource 读取
- 没有 auth / reconnect / health
- 没有按 server 分类显示与权限控制

### 对标参考

- `claude-code/services/mcp/*`
- `claude-code/tools/ListMcpResourcesTool/*`
- `claude-code/tools/ReadMcpResourceTool/*`
- `OpenHarness/src/openharness/mcp/*`

### 建议设计

先补最有价值的 3 件事：

1. `list_mcp_resources`
2. `read_mcp_resource`
3. HTTP transport + reconnect

之后再补：

- MCP auth
- tool/resource capability filtering
- TUI 中 server 状态展示

### 直接收益

- MCP 才会从“可用”变成“平台能力”

---

## 4.6 优化方向六：重构 provider 层，避免继续硬编码

### 当前问题

当前 provider 体系虽然支持多 provider，但总体还是：

- hardcoded provider map
- API key / base URL 映射零散
- 模型能力没有统一元数据

这会导致：

- provider 越多，分支越多
- 模型能力差异无法被 runtime 感知

### 对标参考

- `OpenHarness/src/openharness/api/provider.py`
- `OpenHarness/src/openharness/api/registry.py`
- `hermes-agent/agent/model_metadata.py`
- `hermes-agent/agent/smart_model_routing.py`

### 建议设计

引入 provider registry：

- provider id
- protocol
- auth env keys
- default base URL
- default model
- capability flags
  - tools
  - reasoning
  - image input
  - prompt caching
  - max context

### 直接收益

- `/model`、config、prompt builder、token estimation 都能复用同一份元数据

---

## 4.7 优化方向七：补齐 project instructions / memory files / skills 的主链路

### 当前问题

这个点被三个对标项目都证明非常重要，但 `claude-code-go` 现在还没真正打通。

缺口包括：

- 没有稳定发现并注入 `CLAUDE.md` / `AGENTS.md`
- skill 只是“扫描和返回内容”
- 记忆文件没有和 prompt builder 绑定

### 对标参考

- `OpenHarness/src/openharness/prompts/claudemd.py`
- `claude-code/skills/*`
- `hermes-agent/agent/skill_utils.py`

### 建议设计

在 prompt build 阶段明确注入顺序：

1. workspace instructions
2. active skill instructions
3. memory snippets
4. user prompt

同时给 skills 增加最少的运行时约束：

- source
- trigger
- dependencies
- load mode
- visibility

### 直接收益

- 用户会立刻感受到“项目懂我了”，这是体验跃迁点

---

## 4.8 优化方向八：把 TUI 从“能用”做成“适合长任务”

### 当前问题

当前 TUI 基础够用，但长任务体验仍弱于另外几个项目：

- tool output 展示较轻
- ask / background event / auto-compact / auto-dream 与 UI 状态没有完全同步设计
- 任务、计划、session、MCP、permissions 没有统一的侧边信息视图

### 对标参考

- `claude-code/components/*`
- `OpenHarness` React/Ink TUI
- `hermes-agent` 的 curses/TUI 与 gateway status 组织方式

### 建议设计

不用一开始就做大 UI 重构，先补下面这些：

- tool result collapse/expand
- current mode indicator
- pending task indicator
- current session / plan file / memory status
- MCP server status
- permission prompt modal 化

### 直接收益

- 更适合长时间 coding session
- 也为 subagent/background task 提前铺路

---

## 5. 从三个对标项目各自应该学什么

## 5.1 应该向 claude-code 学

- plan mode / worktree / task / agent tool 的运行时整合方式
- bash/path/sandbox 的安全治理深度
- compact / dream / session memory 的服务层拆分
- MCP resource 与 plugin 的系统化设计

## 5.2 应该向 OpenHarness 学

- engine / prompts / permissions / swarm / plugins 的模块边界
- provider/auth/channel/plugin 的注册式架构
- 作为开源 harness 的“中台层”组织方式

## 5.3 应该向 hermes-agent 学

- 长期记忆、技能沉淀、环境抽象、模型元数据
- 多平台运行的能力边界设计

## 5.4 当前阶段不建议优先学什么

下面这些很强，但不适合作为当前 1-2 个版本的优先项：

- Hermes 的 messaging gateway / voice / social platform
- OpenHarness 的 channels 全家桶
- Claude Code 的远程容器、云端规划、团队同步

原因很简单：

- `claude-code-go` 现在最缺的不是外沿能力，而是本地 runtime 的厚度

---

## 6. 推荐的目标架构

建议把当前结构往下面这个方向演进：

```text
cmd/
  claude/

internal/
  app/            # 顶层装配
  agent/          # 单 agent loop
  prompt/         # system/context/project/skills/memory builder
  runtime/        # mode/task/subagent lifecycle
  tools/          # builtin tools
  permissions/    # policy + validation
  sandbox/        # local sandbox backends
  hooks/          # lifecycle hooks
  mcp/            # tools/resources/auth/transports
  memory/         # session summary + durable memory
  sessions/       # persistence/search/index
  providers/      # provider registry + model metadata
  plugins/        # plugin loader / contracts
  tui/            # presentation layer only
```

关键原则：

- `tui` 只负责交互，不承载核心业务决策
- `agent` 只负责 loop，不承担所有 runtime 状态
- `prompt`、`runtime`、`memory`、`providers` 必须从 `tui` 中分离出来

---

## 7. 分阶段实施建议

## Phase 1：补架构地基 ✅

目标：

- 先解决”功能有了，但没有统一主链路”的问题

完成项：

- ✅ 引入 prompt/context builder (`internal/prompt/`)
- ✅ 注入 `CLAUDE.md` / `AGENTS.md` / skills / memory
- ✅ provider registry 化 (`internal/providers/registry.go`)
- ✅ 将 auto-compact、auto-dream 从 TUI 状态里拆出 service
- ✅ 安全校验补强 (`internal/pathutil/`, `internal/cmdutil`, `internal/urlutil`)

验收标准：

- ✅ 每次请求都能稳定打印一份可解释的 assembled context 结构

## Phase 2：补执行运行时 🚧 (进行中)

目标：

- 把 plan/task 变成真正可依赖的能力

完成项：

- ✅ real plan mode (`internal/runtime/state.go`, `internal/tools/planmode/`)
  - Plan/PlanStep 结构定义
  - Plan/Implement 模式切换
  - 计划文件持久化 (.claude/plan.md)
  - Markdown 步骤解析
- ✅ durable task store (`internal/taskstore/`)
  - 文件持久化 (.claude/tasks.json)
  - CRUD 操作与自动保存
  - 任务依赖追踪 (Blocks/BlockedBy)
  - Session 作用域查询
- ✅ limited local subagent runtime (`internal/subagent/`)
  - Explorer/Writer Worker 类型
  - 并发 spawn 支持
  - Stop/Cleanup 机制
  - 竞态条件修复
- ✅ task-aware TUI (`internal/tui/`)
  - 状态栏显示 plan mode 指示器
  - 状态栏显示待处理任务数
  - tooltask.Initialize() 与 TUI 集成
  - runtimeState 与 TUI 视图集成

待完成：

- ⏳ plan mode 与工具权限关联（限制可用工具）
- ⏳ TUI 任务面板视图
- ⏳ /tasks 斜杠命令显示任务列表

验收标准：

- 一个复杂任务能拆解、批准、执行、跟踪，而不是全靠单轮上下文硬撑

## Phase 3：补安全与扩展

目标：

- 让系统更稳、更可扩展

建议项：

- bash/path/url guard
- sandbox abstraction
- MCP resources + HTTP transport
- plugin contract

验收标准：

- 对外扩展和安全控制都有清晰边界

## Phase 4：补产品化体验

目标：

- 把“工程上能跑”升级为“长期使用舒服”

建议项：

- tool/result 视图增强
- session search
- compact/memory 可视化
- plan/task/subagent 状态视图

验收标准：

- 连续工作数小时仍然可控

---

## 8. 建议优先级

### P0：必须先做

- prompt/context builder
- project instructions 注入
- compact/memory 服务化
- provider registry
- 安全校验补强

### P1：很值得紧接着做

- real plan mode
- durable task system
- 轻量 subagent runtime
- MCP resources / HTTP transport

### P2：做完会明显加分

- plugin contract
- session search / indexing
- richer TUI panels
- 模型能力感知与智能路由

### P3：先不要急

- messaging gateway
- voice
- 多平台社交集成
- 远程容器和云端 orchestration

---

## 9. 一个更务实的版本目标

如果只给 `claude-code-go` 规划两个里程碑，我建议是：

### v0.2

目标是“成为一个扎实的单机 coding harness”

范围：

- prompt builder
- `CLAUDE.md`/`AGENTS.md` 注入
- provider registry
- 真 compact
- memory 提取最小闭环
- 安全治理第一版

### v0.3

目标是“成为一个可规划、可拆解的本地 agent”

范围：

- real plan mode
- durable task store
- local subagent runtime
- MCP resources
- TUI 任务/模式/权限增强

---

## 10. 最后判断

如果把三个对标项目当作参考系：

- 对 `claude-code`，`claude-code-go` 当前差在“运行时深度”和“安全治理深度”
- 对 `OpenHarness`，差在“架构完整度”和“扩展层设计”
- 对 `hermes-agent`，差在“长期记忆/技能/环境抽象”

但从投入产出比看，`claude-code-go` 现在最有价值的方向不是追求大而全，而是：

- 先把本地 coding agent 的 runtime 厚起来
- 把半实现能力打通
- 把上下文、安全、记忆、计划四个核心层补齐

只要这四层补起来，`claude-code-go` 就会从“Go 版复刻项目”变成“有自己工程优势的 agent harness”。
