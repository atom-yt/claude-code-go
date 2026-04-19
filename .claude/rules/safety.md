# 安全性与错误边界

## 安全性

### 命令注入

所有 shell 输入需转义或使用参数化执行：

```go
// 错误：直接拼接字符串
cmd := exec.Command("bash", "-c", "rm "+userInput)

// 正确：使用参数化
cmd := exec.Command("rm", "-rf", userInput)
```

### 路径遍历

验证文件路径在允许范围内：

```go
func validatePath(path string) error {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return err
    }

    // 检查路径是否在允许的目录内
    if !strings.HasPrefix(absPath, allowedDir) {
        return errors.New("path outside allowed directory")
    }

    return nil
}
```

### API Key 保护

- 从不记录或打印敏感信息
- 使用环境变量或配置文件存储
- 不在日志中输出完整的 API Key

```go
// 日志中只显示前几位
log.Printf("Using API key: %s...", apiKey[:8]+"...")
```

### 输入验证

所有外部输入必须验证：

```go
func safeInput(input string) error {
    if input == "" {
        return errors.New("input cannot be empty")
    }
    if len(input) > maxLen {
        return errors.New("input too long")
    }
    // 其他验证...
    return nil
}
```

## 错误边界

### 工具执行错误

工具执行失败不应崩溃 agent 循环：

```go
func (a *Agent) executeTool(ctx context.Context, tu api.ToolUse) api.ToolResult {
    t, ok := a.registry.GetByName(tu.Name)
    if !ok {
        return api.ToolResult{
            Output:  fmt.Sprintf("unknown tool: %q", tu.Name),
            IsError: true,
        }
    }

    result, err := t.Call(ctx, tu.Input)
    if err != nil {
        return api.ToolResult{
            Output:  fmt.Sprintf("tool execution error: %v", err),
            IsError: true,
        }
    }

    return api.ToolResult{
        Output:  result.Output,
        IsError: result.IsError,
    }
}
```

### API 错误传递

API 错误通过 `StreamEvent` 传递：

```go
case api.EventError:
    ch <- StreamEvent{Type: EventError, Error: ev.Error}
    return
```

### 超时和取消

使用 `context.Context` 传播取消信号：

```go
select {
case <-ctx.Done():
    return tools.ToolResult{
        Output:  "operation cancelled",
        IsError: true,
    }
case result := <-workCh:
    return result
}
```

### 资源清理

使用 `defer` 确保资源清理：

```go
func doWork() error {
    file, err := os.Open("file.txt")
    if err != nil {
        return err
    }
    defer file.Close()

    // 使用 file...
    return nil
}
```
