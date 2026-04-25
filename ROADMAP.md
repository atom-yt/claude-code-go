# Atom AI Platform 开发路线图

> 最后更新: 2026-04-25

## 项目概述

Atom AI Platform 是基于 claude-code-go 的全栈 AI Agent 平台，支持多种交互方式：
- **CLI**: 原生终端交互
- **Gateway**: 接口提供者网关（Feishu、WeChat 等）
- **Backend**: 独立后端服务（JWT 认证、PostgreSQL）
- **Frontend**: Next.js 14 Web 应用

## 技术栈总览

| 层级 | 技术 | 用途 |
|------|------|------|
| Frontend | Next.js 14, TypeScript, TailwindCSS, Zustand | Web UI |
| Backend | Go, Gorilla Mux, pgx, JWT | HTTP API |
| CLI | Go, Cobra, Bubble Tea | 终端交互 |
| Gateway | Go, WebSocket | 接口网关 |
| Core | Go, Claude/OpenAI API, Tools | Agent 核心 |

---

## Phase 1: 核心 CLI Agent ✅ 100%

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

## Phase 2: Gateway & 接口系统 ✅ 100%

### 2.1 Gateway CLI
- [x] `gateway list` - 列出已注册的接口提供者
- [x] `gateway start` - 启动接口提供者
- [x] `gateway api` - 启动 HTTP API 服务器

### 2.2 API Server
- [x] `GET /health` - 健康检查
- [x] `POST /api/v1/sessions` - 创建会话
- [x] `GET /api/v1/sessions/{id}` - 获取会话
- [x] `DELETE /api/v1/sessions/{id}` - 删除会话
- [x] `POST /api/v1/chat/completions` - REST 聊天
- [x] `POST /api/v1/chat/stream` - SSE 流式聊天
- [x] WebSocket 支持
- [x] 多部署模式 (single, per-session, pool)
- [x] CORS 和认证中间件

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

## Phase 3: Monorepo Backend 服务 ✅ 100%

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
├── go.mod
└── Makefile
```

### 3.2 认证系统 ✅
- [x] JWT 生成和验证
- [x] 密码哈希 (bcrypt)
- [x] 用户注册/登录
- [x] Token 刷新
- [x] 认证中间件

### 3.3 数据库层 ✅
- [x] PostgreSQL 连接 (pgx/v4)
- [x] Repository 模式实现
- [x] 数据库迁移脚本 (up/down)

### 3.4 API 端点 ✅
- [x] `POST /api/v1/auth/register` - 注册
- [x] `POST /api/v1/auth/login` - 登录
- [x] `POST /api/v1/auth/refresh` - 刷新 Token
- [x] `GET /api/v1/sessions` - 用户会话列表
- [x] `POST /api/v1/sessions` - 创建会话
- [x] 会话和消息管理

### 3.5 Agent 核心 ✅
- [x] `pkg/agent` - 公共 API 包导出
  - [x] `ChatAgent` - 流式聊天代理
  - [x] `ConfigFactory` - 代理实例工厂
  - [x] 支持 Anthropic、OpenAI、Kimi 等 Provider

---

## Phase 4: Frontend Web 应用 ✅ 100%

### 4.1 基础设施 ✅
- [x] Next.js 14 + TypeScript
- [x] TailwindCSS 样式
- [x] Zustand 状态管理
- [x] 基础 UI 组件

### 4.2 页面结构 ✅
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
│   │   ├── session/        # 会话管理组件
│   │   └── stores/             # Zustand 状态管理
│   │   └── lib/                # 工具函数
│   ├── package.json
│   ├── tailwind.config.ts
│   └── tsconfig.json
└── public/              # 静态资源
```

### 4.3 Chat 组件 ✅
- [x] `Chat.tsx` - 主聊天组件
  - [x] `MessageList.tsx` - 消息列表
- [x] `MessageItem.tsx` - 单条消息
  - [x] `ChatInput.tsx` - 输入框
- [x] SSE/WebSocket 流式响应
- [x] 加载状态
- [x] 错误处理

### 4.4 Agent 配置组件 ✅
```
frontend/src/components/agent/
├── AgentConfigForm.tsx       # 主配置表单
├── ModelSelect.tsx            # 模型选择
├── ProviderSelect.tsx        # 提供商选择
├── badge.tsx                 # Badge 组件
├── select.tsx                 # Select 组件
├── constants.ts              # 配置常量
├── types.ts                 # 类型定义
└── index.ts                  # 导出
```

**配置支持**：
- 7 个 AI Provider（Anthropic、OpenAI、Kimi、DeepSeek、Qwen、Ark）
- 7+16 个 AI Model
- API Key 和 Base URL 配置
- System Prompt 自定义
- 快速配置预设

### 4.5 测试覆盖 ✅ 94.3%
- [x] `pkg/agent/agent_test.go` - 公共 API 包单元测试
  - 测试覆盖：94.3%
  - ConfigFactory 测试（4 个子测，18 个断言）
  - ChatAgent 测试（3 个子测，29 个断言）
  - 事件流测试（2 个子测，13 个断言）
  - 配置测试（2 个子测，4 个断言）
  - 并发安全性测试

---

## Phase 5: 集成与部署 ✅ 100%

### 5.1 服务集成 ✅
- [x] Frontend Session 类型修复 (agentId, status)
- [x] Frontend API 分页响应处理 (ListResponse)
- [x] Frontend 消息 API (messagesApi)
- [x] Frontend 流式聊天 API (chatApi.stream)
- [x] Backend 认证中间件集成 (JWT 验证)

### 5.2 部署配置 ✅
- [x] Docker 容器化 (Backend, Frontend, Gateway)
- [x] Docker Compose 多服务编排 (postgres, backend, frontend, gateway)
- [x] 环境变量配置 (.env.example)
- [x] 健康检查与监控 (healthcheck for all services)

---

## Phase 6: 测试与优化 🔲 待开始

### 6.1 集成测试
- [ ] Frontend ↔ Backend 端到端测试
- [ ] 用户认证流程测试
- [ ] Agent 创建与配置测试
- [ ] Session 管理测试
- [ ] 聊天流式响应测试

### 6.2 API 文档
- [ ] OpenAPI/Swagger 规范
- [ ] API 交互式文档
- [ ] 前端 API 客户端文档

### 6.3 监控与日志
- [ ] 结构化日志 (zap/slog)
- [ ] Prometheus 指标导出
- [ ] 分布式追踪 (OpenTelemetry)
- [ ] 错误追踪 (Sentry)

### 6.4 性能优化
- [ ] 数据库连接池优化
- [ ] 缓存层 (Redis)
- [ ] CDN 静态资源
- [ ] 前端代码分割

---

## 当前进度总览

| Phase | 名称 | 状态 | 完成度 |
|-------|------|------|--------|
| **Phase 1** | CLI Agent 核心 | ✅ 完成 | 100% |
| **Phase 2** | Gateway & 接口系统 | ✅ 完成 | 100% |
| **Phase 3** | Monorepo Backend | ✅ 完成 | 100% |
| **Phase 4** | Frontend Web 应用 | ✅ 完成 | 100% |
| **Phase 5** | 集成与部署 | ✅ 完成 | 100% |
| **Phase 6** | 测试与优化 | 🔲 待开始 | 0% |

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
│   ├── interfaces/       # 接口提供者 (Feishu)
│   ├── config/           # 配置管理
│   └── ...               # 其他核心包
├── pkg/
│   ├── agent/            # 公共 API 包
│   └── anthropic/       # Anthropic 客户端
│   └── ...               # 其他工具
├── backend/              # 独立后端服务
├── frontend/             # Next.js 14 应用
├── shared/               # 共享类型
└── .claude/              # 配置目录
```

---

## 快速命令

```bash
# CLI 开发
make build          # 构建 CLI
make test           # 运行测试

# Gateway
go build -o gateway ./cmd/gateway
./gateway list       # 列出接口提供者
./gateway api        # 启动 API 服务器

# Backend
cd backend && go run cmd/server/main.go  # 启动后端服务
go test ./...    # 运行后端测试

# Frontend
cd frontend && npm run dev    # 启动开发服务器
npm run build      # 构建生产版本
npm test         # 运行前端测试

# Docker 部署
cp .env.example .env      # 准备环境变量
# 编辑 .env 填写 API Keys
docker-compose up -d        # 启动所有服务
docker-compose ps           # 查看服务状态
docker-compose logs -f       # 查看日志
docker-compose down -v       # 停止并清理
```

---

## 下一步建议

1. **集成测试** - Frontend 与 Backend API 端到端联调
2. **API 文档** - OpenAPI/Swagger 规范与交互式文档
3. **监控与日志** - 添加结构化日志和监控指标
4. **性能优化** - 缓存层、数据库连接池、前端代码分割