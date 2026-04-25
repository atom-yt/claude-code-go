# Monorepo Structure

```
atom-ai-platform/
├── .claude/                    # Claude Code rules and configuration
│   └── rules/                  # Development rules and standards
├── backend/                    # Go backend services
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Server entry point
│   ├── internal/
│   │   ├── auth/               # JWT authentication
│   │   │   ├── handlers.go     # Auth HTTP handlers
│   │   │   ├── jwt.go          # JWT token logic
│   │   │   ├── middleware.go   # Auth middleware
│   │   │   ├── models.go       # Auth models
│   │   │   ├── password.go     # Password hashing
│   │   │   └── service.go      # Auth service
│   │   ├── db/                 # Database connection
│   │   ├── handlers/           # HTTP handlers
│   │   ├── models/             # Data models
│   │   ├── repository/         # Data access layer
│   │   └── services/           # Business logic
│   ├── migrations/             # Database migrations
│   ├── go.mod                  # Go module
│   ├── Makefile                # Backend Makefile
│   └── README.md               # Backend documentation
├── frontend/                   # Next.js frontend
│   ├── src/
│   │   ├── app/                # Next.js App Router
│   │   ├── components/         # React components
│   │   ├── lib/                # Utilities
│   │   ├── stores/             # State management
│   │   └── types/              # TypeScript types
│   ├── package.json
│   ├── tsconfig.json
│   └── README.md
├── shared/                     # Shared types and specs
│   ├── types/
│   │   ├── doc.go
│   │   └── types.go            # Common type definitions
│   ├── specs/
│   │   └── doc.go
└── internal/                   # Core claude-code-go packages
    ├── api/                    # API client (Anthropic/OpenAI)
    ├── agent/                  # Agent execution engine
    ├── tools/                  # Tools registry
    ├── apiserver/              # API server (extended)
    ├── config/                 # Configuration
    ├── session/                # Session management
    ├── messages/               # Message types
    ├── hooks/                  # Hooks system
    ├── permissions/            # Permission system
    ├── commands/               # Slash commands
    ├── skills/                 # Skill system
    ├── memory/                 # Memory management
    ├── mcp/                    # MCP protocol
    └── tui/                    # Terminal UI
```

## Key Files

### Root Level
- `/Users/yangtong07/Desktop/code/llm-agent/claude-code-go/MONOREPO.md` - Monorepo documentation
- `/Users/yangtong07/Desktop/code/llm-agent/claude-code-go/Makefile` - Root Makefile with all targets
- `/Users/yangtong07/Desktop/code/llm-agent/claude-code-go/Makefile.monorepo` - Standalone monorepo Makefile
- `/Users/yangtong07/Desktop/code/llm-agent/claude-code-go/README.md` - Project README

### Backend
- `/Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/go.mod` - Go module definition
- `/Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/cmd/server/main.go` - Server entry point
- `/Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/migrations/` - Database migrations

### Frontend
- `/Users/yangtong07/Desktop/code/llm-agent/claude-code-go/frontend/README.md` - Frontend documentation
- `/Users/yangtong07/Desktop/code/llm-agent/claude-code-go/frontend/package.json` - NPM dependencies

### Shared
- `/Users/yangtong07/Desktop/code/llm-agent/claude-code-go/shared/types/types.go` - Common type definitions

## Integration Points

### Backend → Core Packages
The backend references core packages via `go.mod`:
```go
replace github.com/atom-yt/claude-code-go => ../
```

This allows the backend to use:
- `internal/api/` for API client
- `internal/agent/` for agent execution
- `internal/tools/` for tool registry
- `internal/apiserver/` for existing server functionality

### Frontend → Shared Types
The frontend uses TypeScript types that mirror the Go types in `shared/types/`.

## Agent Coordination

The monorepo is developed by multiple specialized agents:

1. **Architecture Agent** - Overall design and structure
2. **Database Agent** - Schema design and migrations (COMPLETE)
3. **Auth Agent** - JWT authentication service (COMPLETE)
4. **OpenAPI Agent** - API specification definition
5. **Frontend Agent** - Next.js application (COMPLETE)

Each agent works in their respective directories while coordinating through shared types and APIs.