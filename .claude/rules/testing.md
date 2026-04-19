# 测试规范

## 测试文件

- 测试文件与源文件同包
- 命名格式：`xxx_test.go`

## 测试框架

- 使用 `testing` 标准库
- 使用 `testify/assert` 进行断言

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)
```

## 测试要求

### 单元测试

- 关键路径必须有测试覆盖
- 公开函数必须有测试
- 错误分支必须有测试

### 并发测试

```go
func TestConcurrent(t *testing.T) {
    t.Parallel()
    // 测试代码
}
```

### 表格驱动测试

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"simple", "hello", "HELLO", false},
        {"empty", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := DoSomething(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

## 开发完成后验证

每次开发完成后，必须进行以下验证步骤：

### 1. 编译检查

```bash
make build
```

### 2. 单元测试

```bash
make test
```

### 3. 简单功能测试

```bash
# 查看版本
./claude version

# 查看帮助
./claude --help

# 快速启动测试（如果环境变量已配置）
./claude "hello"
```

### 4. 新增功能专项测试

针对本次开发的功能点进行手动验证。

## 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/tools/...

# 显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
