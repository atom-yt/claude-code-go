# 斜杠命令

## 可用命令

| 命令 | 说明 |
|------|------|
| `/help` | 显示帮助信息 |
| `/clear` | 清除屏幕 |
| `/model [provider/model]` | 切换 provider 或 model |
| `/cost` | 显示当前会话的 token 使用和成本 |
| `/compact` | 压缩历史消息 |

## 新增命令

### 步骤

1. 在 `internal/commands/` 下实现 `Command` 接口
2. 在 `internal/commands/builtin.go` 中注册

### Command 接口

```go
type Command interface {
    Name() string
    Execute(ctx context.Context, args string, state *State) (string, error)
}
```

### 示例

创建 `internal/commands/mycommand.go`：

```go
package commands

import "context"

type MyCommand struct{}

func (c *MyCommand) Name() string {
    return "mycommand"
}

func (c *MyCommand) Execute(ctx context.Context, args string, state *State) (string, error) {
    return "My command executed", nil
}
```

在 `internal/commands/builtin.go` 中注册：

```go
func NewBuiltin() []Command {
    return []Command{
        &Help{},
        &Clear{},
        &MyCommand{},  // 注册新命令
    }
}
```

## 命令参数解析

命令参数通过 `args` 字符串传递，需要自行解析：

```go
func (c *MyCommand) Execute(ctx context.Context, args string, state *State) (string, error) {
    if args == "" {
        return "Usage: /mycommand <arg>", nil
    }
    // 使用 args...
    return "", nil
}
```

## 命令执行上下文

- `ctx context.Context`：上下文，支持取消
- `args string`：命令参数
- `state *State`：当前会话状态，包含 history、settings 等
