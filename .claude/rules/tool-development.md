# 工具开发约定

## 新增工具步骤

1. 在 `internal/tools/` 下创建子包（如 `internal/tools/mytool/`）
2. 实现 `Tool` 接口
3. 在 `internal/tools/registry.go` 中注册
4. 添加单元测试（`*_test.go`）
5. 添加包文档（`doc.go`）

## 工具要求

### InputSchema

- 使用标准 JSON Schema 格式
- 明确字段类型和必填项
- 添加字段描述

```go
func (t *MyTool) InputSchema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "path": map[string]any{
                "type":        "string",
                "description": "File path to read",
            },
        },
        "required": []string{"path"},
    }
}
```

### 错误返回

- 通过 `ToolResult.IsError` 标记
- 提供清晰的错误信息

```go
return tools.ToolResult{
    Output:  fmt.Sprintf("failed to read file: %v", err),
    IsError: true,
}
```

### 超时控制

- 使用 `context.Context` 支持取消
- 长时间操作应检查 `ctx.Done()`

```go
select {
case <-ctx.Done():
    return tools.ToolResult{
        Output:  "operation cancelled",
        IsError: true,
    }
default:
    // 继续执行
}
```

### 路径处理

- 优先使用绝对路径
- 验证路径在允许范围内
- 处理路径遍历攻击

## 工具注册

在 `internal/tools/registry.go` 中注册：

```go
func NewRegistry() *Registry {
    r := &Registry{tools: make(map[string]tools.Tool)}
    r.Register(&read.Read{})
    r.Register(&write.Write{})
    // 注册新工具
    r.Register(&mytool.MyTool{})
    return r
}
```

## 工具并发安全

- 只读工具（如 Read、Glob、Grep）应返回 `IsReadOnly() == true`
- 无状态工具应返回 `IsConcurrencySafe() == true`
- 有状态工具（如 Write）应返回 `IsConcurrencySafe() == false`
