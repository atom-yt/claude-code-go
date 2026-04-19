# Auto-Dream 配置示例

Auto-Dream 是一个自动记忆整合功能，可以自动审查最近的会话并将知识提取到持久化的记忆文件中。

## 启用 Auto-Dream

在 `~/.claude/settings.json` 或项目目录的 `.claude/settings.json` 中添加以下配置：

```json
{
  "autoDreamEnabled": true,
  "minConsolidateHours": 24,
  "minConsolidateSessions": 5,
  "autoMemoryDirectory": ""
}
```

## 配置选项

| 选项 | 类型 | 默认值 | 说明 |
|------|------|---------|------|
| `autoDreamEnabled` | boolean | `false` | 是否启用 auto-dream |
| `minConsolidateHours` | int | `24` | 距上次整合的最小小时数 |
| `minConsolidateSessions` | int | `5` | 触发整合的最小会话数 |
| `autoMemoryDirectory` | string | `""` | 可选的自定义记忆目录路径 |

## 使用方式

### 手动触发

使用 `/dream` 命令手动触发记忆整合：

```
/dream
```

### 查看状态

使用 `/dream status` 命令查看整合状态：

```
/dream status
```

## 记忆目录

记忆文件存储在以下位置：

```
~/.claude/projects/<sanitized-git-root>/memory/
```

其中 `<sanitized-git-root>` 是 git 根目录的净化版本：
- `/` 替换为 `-`
- 前缀为 `-`

例如：`/Users/tong/project` -> `~/.claude/projects/-Users-tong-project/memory/`

## 记忆文件

### MEMORY.md

索引文件，包含所有记忆文件的引用。

### 主题记忆文件

例如：
- `patterns.md` - 编码模式
- `architecture.md` - 项目架构
- `debugging.md` - 调试技巧

## 安全性

- 默认禁用，需要显式启用
- 整合过程只使用只读工具（Read、Glob、Grep）
- 不会修改项目代码，只会更新记忆文件
- 使用文件锁防止并发整合

## 自动触发

当满足以下条件时，会话结束后会自动触发整合：

1. 距上次整合 ≥ `minConsolidateHours` 小时
2. 累积 ≥ `minConsolidateSessions` 个会话记录
3. 没有其他整合进程正在运行
