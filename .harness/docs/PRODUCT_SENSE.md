# Product Sense

## Vision

claude-code-go is a Go implementation of Claude Code - an AI-powered terminal CLI tool that helps developers write better code faster.

## Target Audience

- Go developers who prefer CLI tools
- DevOps engineers working in terminal environments
- Teams using Go in CI/CD pipelines
- Users who need fast, compiled binaries without runtime dependencies

## Core Value Propositions

1. **Performance** - Compiled binary, fast startup, low memory footprint
2. **Reliability** - Strong typing, compile-time error checking
3. **Portability** - Single binary distribution, no runtime dependencies
4. **Integration** - Native Go ecosystem integration

## Key Features

### Agent Capabilities
- Multi-provider LLM support (Anthropic, OpenAI-compatible)
- Tool calling with parallel execution
- Session persistence and memory
- MCP (Model Context Protocol) support

### Built-in Tools
- File operations: Read, Write, Edit, Glob, Grep
- Shell execution: Bash
- Web operations: WebFetch, WebSearch
- Task management: Todo, Task, Ask
- Plan mode for complex tasks

### UI/UX
- Interactive TUI using bubbletea
- Syntax highlighting with glamour
- Autocomplete and help system
- Permission control and hooks

## Technical Priorities

1. **Safety First** - Path validation, SSRF protection, permission checks
2. **Agent Stability** - Graceful error handling, no crashes
3. **Performance** - Efficient tool execution, minimal overhead
4. **Extensibility** - Easy to add tools, providers, skills

## Quality Standards

- Code coverage > 70% for critical paths
- No panics in production code
- All errors handled explicitly
- Zero data loss scenarios

## Future Roadmap

- [ ] Improved TUI with split views
- [ ] Enhanced plan mode visualization
- [ ] More MCP server integrations
- [ ] Performance profiling and optimization
- [ ] Multi-language support (beyond Go)