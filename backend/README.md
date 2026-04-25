# Backend

Go-based HTTP server providing RESTful APIs and WebSocket support for the Atom AI Platform.

## Directory Structure

```
backend/
├── cmd/
│   └── server/         # Server entry point
├── internal/
│   ├── auth/           # JWT authentication
│   ├── db/             # Database connection
│   ├── handlers/       # HTTP handlers
│   ├── models/         # Data models
│   ├── repository/     # Data access layer
│   └── services/       # Business logic
├── migrations/         # Database migrations
├── go.mod             # Go module definition
└── README.md          # This file
```

## Integration with Core Packages

The backend references core packages from the parent `internal/` directory:

- `internal/api/` - Anthropic/OpenAI API client
- `internal/agent/` - Core Agent execution engine
- `internal/tools/` - Tools registry and execution
- `internal/apiserver/` - Existing API server implementation

These are accessed via the `go.mod` replace directive:
```go
replace github.com/atom-yt/claude-code-go => ../
```

## Getting Started

### Prerequisites

- Go 1.22.10 or later
- PostgreSQL 14+
- Environment variables configured

### Environment Variables

Create a `.env` file in the backend directory:

```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/atom_ai

# Authentication
JWT_SECRET=your-secret-key-change-this
JWT_EXPIRATION=24h

# API Keys
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# CORS
FRONTEND_URL=http://localhost:3000
```

### Development

```bash
# Install dependencies
go mod download
go mod tidy

# Run the server
go run cmd/server/main.go

# Or use the monorepo Makefile from project root
make backend-dev
```

The server will start on `http://localhost:8080`

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/auth/...
```

### Building

```bash
# Build the server binary
go build -o bin/server cmd/server/main.go

# Run the binary
./bin/server
```

## API Endpoints

### Health Check
- `GET /health` - Server health status

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

## Architecture

```
┌─────────────┐
│   Handlers  │  HTTP/WebSocket layer
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Services   │  Business logic
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Repository  │  Data access
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Database   │  PostgreSQL
└─────────────┘
```

## Database

Migrations are stored in the `migrations/` directory and are managed by the Database Schema Agent.

To run migrations (when implemented):
```bash
go run cmd/migrate/main.go up
```

## Authentication

The backend uses JWT (JSON Web Tokens) for authentication:

1. User authenticates with `/api/v1/auth/login`
2. Server returns a JWT token
3. Client includes token in `Authorization: Bearer <token>` header
4. Server validates token on protected routes

Implementation is provided by the JWT Authentication Agent.

## WebSocket

The backend supports WebSocket connections for real-time communication:

- Session message streaming
- Tool execution updates
- Agent status updates

WebSocket endpoints are defined in `internal/apiserver/` and extended here.

## Development Notes

### Adding a New Endpoint

1. Define the handler in `internal/handlers/`
2. Register the route in `cmd/server/main.go`
3. Add service logic in `internal/services/`
4. Add repository methods in `internal/repository/`
5. Update OpenAPI specification

### Adding a New Model

1. Define the model struct in `internal/models/`
2. Add repository methods in `internal/repository/`
3. Create a migration in `migrations/`
4. Update the API handlers

## Dependencies

Key dependencies:
- `github.com/gorilla/mux` - HTTP router
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/golang-jwt/jwt/v5` - JWT implementation
- `github.com/google/uuid` - UUID generation

## Contributing

When contributing to the backend:

1. Follow Go best practices and the project coding standards
2. Write tests for new features
3. Update documentation
4. Ensure all tests pass before submitting