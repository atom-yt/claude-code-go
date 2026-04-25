# Atom AI Platform

A full-stack AI Agent platform built with Go backend and Next.js frontend, extending the claude-code-go CLI tool with web-based capabilities.

## Architecture

This is a monorepo containing three main components:

```
atom-ai-platform/
├── backend/           # Go backend services
├── frontend/          # Next.js frontend application
├── shared/            # Shared types and API specifications
└── internal/          # Core claude-code-go packages (CLI agent)
```

## Components

### Backend (`backend/`)

Go-based HTTP server providing RESTful APIs and WebSocket support for the platform.

**Structure:**
- `internal/db/` - Database connection and utilities
- `internal/auth/` - JWT authentication and authorization
- `internal/models/` - Data models (User, Session, Agent, etc.)
- `internal/repository/` - Data access layer
- `internal/services/` - Business logic layer
- `internal/handlers/` - HTTP request handlers
- `migrations/` - Database migrations
- `cmd/server/` - Server entry point

**Key Integrations:**
- References `internal/api/` - Anthropic/OpenAI API client
- References `internal/agent/` - Core Agent execution engine
- References `internal/tools/` - Tools registry and execution
- References `internal/apiserver/` - Existing API server implementation
- Extends `internal/apiserver/` with new endpoints

### Frontend (`frontend/`)

Next.js 14+ application with App Router and TypeScript.

**To be implemented by:**
- Frontend development agent

**Planned Features:**
- Chat interface for AI interactions
- Session management UI
- Agent configuration dashboard
- Real-time tool execution visualization
- User authentication flows

### Shared (`shared/`)

Shared types and specifications used across backend and frontend.

**Structure:**
- `types/` - Common data structures and API types
- `specs/` - OpenAPI specifications

**Purpose:**
- Ensure type safety across boundaries
- Single source of truth for API contracts
- Enable code generation for clients

### Internal (`internal/`)

Core claude-code-go CLI packages, serving as the foundation for the platform.

**Key Packages:**
- `api/` - API client (Anthropic/OpenAI)
- `agent/` - Agent main loop and execution
- `tools/` - Tool registry and implementations
- `apiserver/` - API server (WebSocket support, handlers)
- `config/` - Configuration management
- `session/` - Session persistence
- `messages/` - Message types
- `hooks/` - Hooks system
- `permissions/` - Permission system
- `commands/` - Slash commands
- `skills/` - Skill system
- `memory/` - Memory management
- `mcp/` - MCP protocol support
- `tui/` - Terminal UI (CLI only)

## Getting Started

### Prerequisites

- Go 1.22.10 or later
- Node.js 18+ (for frontend)
- PostgreSQL 14+ (for backend database)
- Anthropic API key or compatible provider

### Backend Development

```bash
# Navigate to backend directory
cd backend

# Install dependencies
go mod download

# Run database migrations (to be implemented)
# go run cmd/migrate/main.go up

# Start development server
go run cmd/server/main.go

# Run tests
go test ./...

# Build
go build -o bin/server cmd/server/main.go
```

### Frontend Development

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Run tests
npm test
```

## Database Schema

The platform uses PostgreSQL with the following main entities:

- `users` - User accounts and authentication
- `sessions` - AI conversation sessions
- `agents` - Agent configurations and states
- `tool_executions` - Tool execution history
- `permissions` - User permissions and roles

*Database schema is being designed by the database schema agent.*

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/refresh` - Refresh JWT token

### Sessions
- `POST /api/v1/sessions` - Create new session
- `GET /api/v1/sessions` - List user sessions
- `GET /api/v1/sessions/{id}` - Get session details
- `DELETE /api/v1/sessions/{id}` - Delete session

### Agents
- `GET /api/v1/agents` - List available agents
- `POST /api/v1/agents` - Create custom agent
- `GET /api/v1/agents/{id}` - Get agent details
- `POST /api/v1/agents/{id}/execute` - Execute agent

### WebSocket
- `WS /api/v1/ws` - Real-time session communication

*Full OpenAPI specification is being defined by the OpenAPI spec agent.*

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                     Next.js Frontend                     │
│  (Chat UI, Session Management, Agent Dashboard)         │
└──────────────────────┬──────────────────────────────────┘
                       │ HTTP/WebSocket
                       ▼
┌─────────────────────────────────────────────────────────┐
│                      Go Backend                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Handlers   │→ │   Services   │→ │ Repository   │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                       │                  │               │
│                       ▼                  ▼               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │     Auth     │  │   Database   │  │  Agent Core  │  │
│  │   (JWT)      │  │  (PostgreSQL)│  │   (internal) │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                         │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│                  External APIs                          │
│  ┌──────────────┐  ┌──────────────┐                     │
│  │  Anthropic   │  │    OpenAI    │                     │
│  │     API      │  │     API      │                     │
│  └──────────────┘  └──────────────┘                     │
└─────────────────────────────────────────────────────────┘
```

## Development Workflow

This monorepo is developed by multiple specialized agents:

- **Architecture Agent** - Overall design and structure
- **Database Agent** - Schema design and migrations
- **Auth Agent** - JWT authentication service
- **OpenAPI Agent** - API specification definition
- **Frontend Agent** - Next.js application

Each agent works on their respective area while coordinating through this shared monorepo.

## Environment Variables

Create a `.env` file in the backend directory:

```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/atom_ai

# Authentication
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

# API Keys
ANTHROPIC_API_KEY=your-anthropic-key
OPENAI_API_KEY=your-openai-key

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Frontend URL (for CORS)
FRONTEND_URL=http://localhost:3000
```

## Testing

```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
npm test

# E2E tests (to be implemented)
npm run test:e2e
```

## Deployment

### Backend

```bash
cd backend
go build -o server cmd/server/main.go
./server
```

### Frontend

```bash
cd frontend
npm run build
# Output in .next/ directory
```

### Docker

Docker configurations to be added.

## Contributing

1. Create a feature branch from `main`
2. Make your changes in the relevant component directory
3. Ensure all tests pass
4. Submit a pull request

## License

MIT License - See LICENSE file for details