# Git Workflow

## Branching Strategy

### Branch Naming

| Type | Format | Example |
|------|--------|---------|
| Feature | `feature/xxx` | `feature/websearch-tool` |
| Bugfix | `bugfix/xxx` | `bugfix/file-read-crash` |
| Documentation | `docs/xxx` | `docs/update-readme` |
| Refactoring | `refactor/xxx` | `refactor/tool-registry` |

### Main Branch

- `main` is always stable
- Feature branches are created from `main`
- Completed work is merged back to `main` via PR

## Commit Messages

Follow conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring (no functional change)
- `test`: Test additions/changes
- `chore`: Maintenance tasks

### Examples

```
feat(tools): add WebSearch tool for DuckDuckGo integration

- Implement DuckDuckGo HTML scraping
- Add URL validation via urlutil package
- Register tool in registry

Closes #123
```

```
fix(permissions): resolve path traversal vulnerability

Add proper path validation in Read tool to prevent
accessing files outside working directory.

Security: high
```

## Pull Requests

### PR Requirements

- All tests pass
- Code follows gofmt formatting
- New code has test coverage
- PR description includes:
  - Summary of changes
  - Related issue numbers
  - Testing instructions
  - Breaking changes (if any)

### PR Title Format

```
feat: add WebSearch tool
fix: resolve file path traversal issue
docs: update provider configuration guide
```

## Pre-Commit Hooks

Run before committing:
- `go fmt ./...`
- `go test ./...`
- `go vet ./...`

## Pre-Push Checks

Before pushing:
- Run full validation: `go run .harness/scripts/validate.go`
- Ensure no TODO/FIXME left (unless documented)
- Verify branch is up to date with main