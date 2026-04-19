# 编码规范

## 基础规则

1. **遵循 Go 标准规范**：使用 `gofmt` 格式化代码
2. **包注释**：每个包应有简洁的文档注释
3. **导出函数**：必须有文档注释，说明用途和参数
4. **错误处理**：所有错误必须处理，不忽略 `err`

## 接口设计

### Tool 接口

工具接口定义在 `internal/tools/tool.go`：

```go
type Tool interface {
    Name() string                           // 唯一标识符
    Description() string                    // Claude 看到的描述
    InputSchema() map[string]any            // JSON Schema
    Call(ctx context.Context, input map[string]any) (ToolResult, error)
    IsReadOnly() bool                       // 是否只读
    IsConcurrencySafe() bool                // 是否并发安全
}
```

### ToolResult 结构

```go
type ToolResult struct {
    Output string // 返回给 Claude 的文本
    IsError bool  // 是否为错误结果
}
```

## Agent 循环流程

1. 接收用户输入，添加到历史消息
2. 调用 API 流式获取响应
3. 解析工具调用（tool_use）
4. 顺序执行：pre-hooks → 权限检查 → 执行工具 → post-hooks
5. 将工具结果添加到消息，继续循环直到完成

## 并发执行规则

- `IsConcurrencySafe() == true` 的工具可并行执行
- 工具执行前必须通过权限检查
- Post-tool hooks 异步执行，不阻塞主流程
- 使用 `sync.WaitGroup` 协调并发工具完成

## 错误处理

- 工具执行失败不应崩溃 agent 循环
- API 错误通过 `StreamEvent{Type: EventError}` 传递
- 超时和取消通过 `context.Context` 传播

## 性能考虑

- 工具输出超过 50,000 字符会自动截断
- SSE 流式解析使用缓冲 channel
- 大量工具调用时优先并发执行
