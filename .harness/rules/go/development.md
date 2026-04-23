# Go Development Rules

## Code Style

### Formatting

Always format code with `gofmt`:

```bash
gofmt -s -w .
```

Use `goimports` for import organization:

```bash
goimports -w .
```

### Naming

- Exported names: PascalCase (`type Tool interface`)
- Unexported names: camelCase (`type toolRegistry struct`)
- Interfaces: `-er` suffix for behavior (`Reader`, `Writer`)

### Error Handling

Always handle errors, never ignore them:

```go
// ✗ WRONG
file, _ := os.Open("file.txt")

// ✓ CORRECT
file, err := os.Open("file.txt")
if err != nil {
    return fmt.Errorf("failed to open file: %w", err)
}
```

Use `fmt.Errorf` with `%w` for error wrapping:

```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

## Structs and Interfaces

### Interface Design

- Keep interfaces small (few methods)
- Accept interfaces, return structs
- Define interfaces where they're used, not where they're implemented

```go
// Good: interface defined where used
func Process(ctx context.Context, r io.Reader) error {
    // ...
}
```

### Struct Initialization

Use field names for clarity:

```go
user := User{
    Name:  "John",
    Email: "john@example.com",
}
```

## Concurrency

### Goroutines

Always handle goroutine lifecycle:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    for {
        select {
        case <-ctx.Done():
            return
        case val := <-ch:
            // process
        }
    }
}()
```

### Channels

Prefer buffered channels for producers:

```go
ch := make(chan Result, 10)
```

Always close channels:

```go
defer close(ch)
```

## Testing

### Test Organization

Use table-driven tests:

```go
func TestParse(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Result
        wantErr bool
    }{
        {"valid", "test", Result{}, false},
        {"invalid", "", Result{}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Parse(tt.input)
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

### Test Naming

Use `Test<FunctionName>` format. Use `t.Run()` for subtests.

## Context

Always accept `context.Context` as first parameter:

```go
func DoSomething(ctx context.Context, arg string) (Result, error) {
    select {
    case <-ctx.Done():
        return Result{}, ctx.Err()
    default:
        // do work
    }
}
```

## Performance

### Profiling

```bash
# CPU profile
go test -cpuprofile=cpu.prof -bench=.

# Memory profile
go test -memprofile=mem.prof -bench=.
```

### Benchmarks

Write benchmarks for critical paths:

```go
func BenchmarkProcess(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Process(testInput)
    }
}
```

## Dependencies

- Keep dependencies minimal
- Update regularly with `go get -u`
- Run `go mod tidy` before commits
- Review dependency changes in PRs