# 项目概览

## 项目定位

这是 [Claude Code](https://claude.ai/code) 的 Go 语言实现版本，是一个 AI 驱动的终端 CLI 工具。

## 核心功能

- 与 Claude API 进行流式对话
- Agent 工具调用循环（读写文件、运行命令、搜索等）
- MCP（Model Context Protocol）服务器支持
- 权限控制与 Hooks 系统
- 会话持久化与恢复

## 技术栈

| 组件 | 选型 | 用途 |
|------|------|------|
| CLI 框架 | cobra | 命令行参数解析 |
| TUI 框架 | bubbletea | 终端用户界面 |
| 终端样式 | lipgloss | 颜色和布局 |
| Markdown | glamour | Markdown 渲染 |
| Go 版本 | 1.22.10 | 最低版本要求 |

## 目录结构

```
claude-code-go/
├── cmd/claude/          # CLI 入口
├── internal/
│   ├── agent/           # Agent 主循环
│   ├── api/             # API 客户端（Anthropic/OpenAI）
│   ├── tools/           # 工具实现
│   ├── permissions/     # 权限系统
│   ├── hooks/           # Hooks 系统
│   ├── mcp/             # MCP 协议
│   ├── config/          # 配置管理
│   ├── session/         # 会话管理
│   ├── messages/        # 消息类型
│   ├── commands/        # 斜杠命令
│   └── tui/             # 终端 UI
├── pkg/                 # 可复用包（如预留）
├── rules/               # 开发规范
└── CLAUDE.md           # 规则目录索引
```
