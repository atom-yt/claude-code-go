# Atom AI Platform 开发路线图

> 最后更新: 2026-04-25

## 项目概述

Atom AI Platform 是基于 claude-code-go 的全栈 AI Agent 平台，支持多种交互方式：
- **CLI**: 原生终端交互
- **Gateway**: 接口提供者网关（Feishu、WeChat 等）
- **Backend**: 独立后端服务（JWT 认证、PostgreSQL）
- **Frontend**: Next.js Web 应用

---

## Phase 1: 核心基础设施 ✅ 已完成

### 1.1 CLI Agent 核心
- [x] Agent 主循环实现
- [x] 工具系统 (Read, Write, Bash, Glob, Grep, WebSearch, WebFetch 等)
- [x] 多 Provider 支持 (Anthropic, OpenAI, Kimi, DeepSeek, Qwen, Ark)
- [x] MCP 协议支持
- [x] 权限控制系统
- [x] Hooks 系统
- [x] 会话持久化
- [x] TUI 界面

### 1.2 配置与记忆
- [x] 配置优先级系统 (CLI > 环境变量 > 项目配置 > 用户配置)
- [x] 自动记忆管理 (claude-mem)
- [x] 历史压缩 (SummaryStore + MicroCompactor)

### 1.3 测试覆盖
- [x] 核心包单元测试
- [x] 工具测试
- [x] API 客户端测试

---

## Phase 2: Gateway & 接口系统 ✅ 已完成

### 2.1 Gateway CLI
- [x] `gateway list` - 列出已注册的接口提供者
- [x] `gateway start` - 启动接口提供者
- [x] `gateway api` - 启动 HTTP API 服务器

### 2.2 API Server
- [x] REST API 端点
  - [x] `GET /health` - 健康检查
  - [x] `POST /api/v1/sessions` - 创建会话
  - [x] `GET /api/v1/sessions/{id}` - 获取会话
  - [x] `DELETE /api/v1/sessions/{id}` - 删除会话
  - [x] `POST /api/v1/chat/completions` - REST 聊天
  - [x] `POST /api/v1/chat/stream` - SSE 流式聊天
- [x] 多部署模式 (single, per-session, pool)
- [x] CORS 支持
- [x] API 认证

### 2.3 Feishu Interface Provider
- [x] WebSocket 连接管理
- [x] Webhook 接收器
- [x] 消息格式化 (Text, Markdown, Card)
- [x] 会话管理
- [x] 速率限制
- [x] 消息队列
- [x] 媒体处理 (图片、文件)
- [x] 签名验证

---

## Phase 3: Monorepo 后端服务 🔄 进行中

### 3.1 项目结构 ✅
```
backend/
├── cmd/server/main.go      # 服务入口
├── internal/
│   ├── auth/               # JWT 认证
│   ├── db/                 # 数据库连接
│   ├── handlers/           # HTTP 处理器
│   ├── models/             # 数据模型
│   ├── repository/         # 数据访问层
│   └── services/           # 业务逻辑层
├── migrations/             # 数据库迁移
└── go.mod
```

### 3.2 认证系统 ✅
- [x] JWT 生成和验证
- [x] 密码哈希 (bcrypt)
- [x] 用户注册/登录
- [x] Token 刷新
- [x] 认证中间件

### 3.3 数据库层 ✅
- [x] PostgreSQL 连接 (pgx)
- [x] 数据库迁移脚本
- [x] Repository 模式实现

### 3.4 API 端点 ✅
- [x] `POST /api/v1/auth/register` - 注册
- [x] `POST /api/v1/auth/login` - 登录
- [x] `POST /api/v1/auth/refresh` - 刷新 Token
- [x] `GET /api/v1/sessions` - 用户会话列表
- [x] `POST /api/v1/sessions` - 创建会话
- [x] `GET /api/v1/agents` - Agent 列表
- [x] `POST /api/v1/agents` - 创建 Agent

### 3.5 API Endpoints ✅
- [x] `POST /api/v1/auth/register` - 注册
- [x] `POST /api/v1/auth/login` - 登录
- [x] `POST /api/v1/auth/refresh` - 刷新 Token
- [x] `GET /api/v1/sessions` - 用户会话列表
- [x] `POST /api/v1/sessions` - 创建会话
- [x] `GET /api/v1/agents` - Agent 列表
- [x] `POST /api/v1/agents` - 创建 Agent

### 3.6 Agent Core Integration ✅
- [x] `pkg/agent` - 公共 API 包导出
  - [x] `ChatAgent` - 流式聊天代理
  - [x] `ConfigFactory` - 代理实例工厂
  - [x] `EventType` 枚举 - Delta, ToolCall, ToolResult, Error, Done
  - [x] `Config` - 配置结构体
- [x] `POST /api/v1/chat` - SSE 流式聊天
- [x] `WS /ws/chat` - WebSocket 聊天
- [x] 会话代理缓存

### 3.7 待完成 🔲
- [ ] 单元测试补充
- [ ] 集成测试
- [ ] 数据库连接池优化
- [ ] Docker 容器化

---

## Phase 4: Frontend Web 应用 🔄 进行中

### 4.1 项目结构 ✅
```
frontend/
├── src/
│   ├── app/                # Next.js App Router
│   │   ├── (auth)/         # 认证相关页面
│   │   ├── (dashboard)/    # 主应用页面
│   │   └── api/            # API 路由
│   ├── components/         # React 组件
│   │   ├── ui/             # 基础 UI 组件
│   │   ├── chat/           # 聊天组件
│   │   ├── auth/           # 认证组件
│   │   ├── agent/          # Agent 配置组件
│   │   └── session/        # 会话管理组件
│   ├── stores/             # Zustand 状态管理
│   └── lib/                # 工具函数
├── package.json
└── tailwind.config.ts
```

### 4.2 基础设施 ✅
- [x] Next.js 14 + TypeScript
- [x] TailwindCSS 样式
- [x] Zustand 状态管理
- [x] 基础 UI 组件 (Button, Input, Textarea, Tabs, Label)

### 4.3 页面结构 ✅
- [x] 登录页面 `/login`
- [x] 注册页面 `/register`
- [x] 仪表盘 `/dashboard`
- [x] 会话列表 `/sessions`
- [x] 会话详情 `/sessions/[id]`
- [x] Agent 列表 `/agents`
- [x] Agent 创建 `/agents/new`
- [x] Agent 详情 `/agents/[id]`

### 4.4 待完成 🔲
- [ ] Chat 组件实现（流式响应）
- [ ] Agent 配置表单
- [ ] 与 Backend API 集成
- [ ] 认证状态持久化
- [ ] 响应式布局优化
- [ ] 单元测试

---

## Phase 5: 集成与部署 🔲 待开始

### 5.1 服务集成
- [ ] Frontend ↔ Backend API 对接
- [ ] Backend ↔ Agent Core 集成
- [ ] Gateway ↔ Backend 状态同步

### 5.2 部署配置
- [ ] Docker Compose 多服务编排
- [ ] 环境变量配置规范
- [ ] 健康检查与监控
- [ ] 日志聚合

### 5.3 文档完善
- [ ] API 文档 (OpenAPI/Swagger)
- [ ] 部署指南
- [ ] 用户手册

---

## Phase 6: 高级特性 🔲 规划中

### 6.1 多租户支持
- [ ] 租户隔离
- [ ] 资源配额
- [ ] 使用计费

### 6.2 扩展接口
- [ ] WeChat 企业微信
- [ ] Telegram Bot
- [ ] Slack App

### 6.3 高可用
- [ ] 集群部署
- [ ] 负载均衡
- [ ] 故障恢复

---

## 当前工作焦点

### 优先级 1: Backend 与 Agent 集成
1. Backend 调用 `internal/agent` 执行对话
2. 实现 WebSocket 实时通信
3. 补充单元测试

### 优先级 2: Frontend 完善
1. Chat 组件实现流式响应
2. 与 Backend API 完整对接
3. 认证流程完整测试

### 优先级 3: 文档与测试
1. 编写集成测试
2. 完善 API 文档
3. 用户使用指南

---

## 技术栈总览

| 层级 | 技术 | 用途 |
|------|------|------|
| Frontend | Next.js 14, TypeScript, TailwindCSS, Zustand | Web UI |
| Backend | Go, Gorilla Mux, pgx, JWT | HTTP API |
| Database | PostgreSQL 14+ | 持久化存储 |
| CLI | Go, Cobra, Bubble Tea | 终端交互 |
| Gateway | Go, WebSocket, HTTP | 接口网关 |
| AI | Anthropic API / OpenAI API | 模型调用 |

---

## 目录结构

```
claude-code-go/
├── cmd/
│   ├── claude/           # CLI 入口
│   └── gateway/          # Gateway CLI
├── internal/
│   ├── agent/            # Agent 核心
│   ├── api/              # API 客户端
│   ├── apiserver/        # API Server
│   ├── interfaces/       # 接口提供者
│   │   └── feishu/       # 飞书实现
│   ├── tools/            # 工具实现
│   ├── config/           # 配置管理
│   └── ...               # 其他核心包
├── backend/              # 独立后端服务
├── frontend/             # Next.js 应用
├── shared/               # 共享类型
└── Makefile              # 构建命令
```

---

## 快速命令

```bash
# CLI 开发
make build          # 构建 CLI
make test           # 运行测试

# Gateway
go build -o gateway ./cmd/gateway
./gateway list      # 列出接口
./gateway api       # 启动 API 服务

# Backend
cd backend && go run cmd/server/main.go

# Frontend
cd frontend && npm run dev

# 全部构建
make monorepo-build
```