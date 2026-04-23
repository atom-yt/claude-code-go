# Safety Rules

These rules must be followed to ensure code security and stability.

## Critical Rules (Violations must be fixed)

### 1. No Command Injection
All shell input must be parameterized, never concatenated:

```go
// ✗ WRONG
cmd := exec.Command("bash", "-c", "rm "+userInput)

// ✓ CORRECT
cmd := exec.Command("rm", userInput)
```

### 2. Path Traversal Protection
Always validate file paths are within allowed directories:

```go
func validatePath(path string) error {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return err
    }
    if !strings.HasPrefix(absPath, allowedDir) {
        return errors.New("path outside allowed directory")
    }
    return nil
}
```

### 3. No Secret Logging
Never log API keys, tokens, or credentials:

```go
// ✗ WRONG
log.Printf("API Key: %s", apiKey)

// ✓ CORRECT
log.Printf("Using API key: %s...", apiKey[:8]+"...")
```

### 4. Error Handling
All errors must be handled:

```go
// ✗ WRONG
file, _ := os.Open("file.txt")

// ✓ CORRECT
file, err := os.Open("file.txt")
if err != nil {
    return fmt.Errorf("failed to open file: %w", err)
}
```

## Important Rules

### 5. Context Propagation
Always pass `context.Context` through the call stack:

```go
func doWork(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // do work
    }
}
```

### 6. Resource Cleanup
Always clean up resources using `defer`:

```go
file, err := os.Open("file.txt")
if err != nil {
    return err
}
defer file.Close()
```

### 7. No Panics in Production
Panics should only be used for truly unrecoverable conditions. Prefer returning errors.

## Tool Execution Safety

- Tools must never crash the agent loop
- Timeout and cancellation must be supported
- Tool execution errors return `ToolResult{IsError: true}`

## Web Security

- URL validation before making HTTP requests
- SSRF protection for internal network access
- User input validation for file operations