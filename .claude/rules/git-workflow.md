# Git 工作流

## 分支命名

使用以下命名规范：

| 类型 | 格式 | 示例 |
|------|------|------|
| 新功能 | `feature/xxx` | `feature/websearch-tool` |
| Bug 修复 | `bugfix/xxx` | `bugfix/file-read-crash` |
| 文档 | `docs/xxx` | `docs/update-readme` |
| 重构 | `refactor/xxx` | `refactor/tool-registry` |

## 提交信息

提交信息应简明扼要，说明"为什么"而非"做了什么"：

```
Add WebSearch tool for real-time information retrieval

- Implement DuckDuckGo HTML search
- Add WebSearch tool registry
- Update documentation
```

避免使用：
```
Update code
Fix bug
Add feature
```

## 主分支

- `main` - 主分支，始终保持稳定
- 新功能从 `main` 分支切出
- 完成后通过 PR 合并回 `main`

## Commit 检查

提交前确保：

1. 代码已通过 `make test`
2. 代码已通过 `gofmt` 格式化
3. 没有遗留的 `TODO` 或 `FIXME`（除非必要）
4. 新增代码有对应的测试

## Pull Request

PR 标题应遵循约定提交格式：

```
feat: add WebSearch tool
fix: resolve file path traversal issue
docs: update provider configuration guide
```

PR 描述应包含：
- 变更概述
- 相关 issue 链接
- 测试方法
- 破坏性变更说明（如有）
