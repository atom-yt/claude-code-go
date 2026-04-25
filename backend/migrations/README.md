# Database Migrations

This directory contains PostgreSQL database migrations for the atom-ai-platform.

## Prerequisites

1. **PostgreSQL** (version 14+)
2. **pgvector extension** - Install from: https://github.com/pgvector/pgvector
3. **golang-migrate** - Install with:
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

## Installation

### Install pgvector extension

```bash
# On macOS with Homebrew
brew install pgvector

# On Ubuntu/Debian
# Clone and compile pgvector
git clone --branch v0.5.1 https://github.com/pgvector/pgvector.git
cd pgvector
make
sudo make install
```

### Enable pgvector in PostgreSQL

```bash
# Connect to your database
psql -U postgres -d atom_ai

# Create extension
CREATE EXTENSION IF NOT EXISTS vector;
```

## Migrations List

| Migration | Description |
|-----------|-------------|
| 000001 | Create users table |
| 000002 | Create agents table |
| 000003 | Create chat_sessions table |
| 000004 | Create messages table |
| 000005 | Create knowledge_bases table |
| 000006 | Create knowledge_documents table |
| 000007 | Create knowledge_chunks table (with pgvector) |
| 000008 | Create updated_at triggers |

## Running Migrations

### Using the Go helper

```bash
# Set database URL
export DB_URL="postgres://user:password@localhost:5432/atom_ai?sslmode=disable"

# Run all pending migrations
go run migrate.go -up

# Rollback all migrations
go run migrate.go -down

# Rollback specific number of migrations
go run migrate.go -down -steps 1

# Create a new migration
go run migrate.go -create add_user_preferences

# Show current migration version
go run migrate.go

# Dry run (show what would be done)
go run migrate.go -up -dry-run
```

### Using Make

```bash
# Set database URL
export DB_URL="postgres://user:password@localhost:5432/atom_ai?sslmode=disable"

# Run all migrations
make migrate-up

# Rollback all migrations
make migrate-down

# Rollback specific number of migrations
make migrate-steps STEPS=1

# Create new migration
make migrate-create NAME=add_user_preferences

# Show current version
make migrate-version
```

### Using golang-migrate CLI directly

```bash
# Run all up migrations
migrate -path ./migrations -database "postgres://user:password@localhost:5432/atom_ai?sslmode=disable" up

# Run all down migrations
migrate -path ./migrations -database "postgres://user:password@localhost:5432/atom_ai?sslmode=disable" down

# Rollback specific number of migrations
migrate -path ./migrations -database "postgres://user:password@localhost:5432/atom_ai?sslmode=disable" down 1

# Show current version
migrate -path ./migrations -database "postgres://user:password@localhost:5432/atom_ai?sslmode=disable" version

# Create new migration
migrate create -ext sql -dir ./migrations -seq add_user_preferences
```

## Database Schema

### Users Table
- `id` - UUID primary key
- `email` - Unique email address
- `password_hash` - Bcrypt hashed password
- `display_name` - Display name
- `role` - User role (user/admin)
- `created_at` - Creation timestamp

### Agents Table
- `id` - UUID primary key
- `user_id` - Foreign key to users
- `name` - Agent name
- `description` - Agent description
- `system_prompt` - System prompt
- `model` - AI model name
- `provider` - Provider name
- `temperature` - Temperature parameter
- `max_tokens` - Max tokens
- `tools` - Array of tool names
- `knowledge_ids` - Array of knowledge base IDs
- `created_at`, `updated_at` - Timestamps

### Chat Sessions Table
- `id` - UUID primary key
- `user_id` - Foreign key to users
- `agent_id` - Foreign key to agents
- `title` - Session title
- `status` - Session status (active/archived/deleted)
- `created_at`, `updated_at` - Timestamps

### Messages Table
- `id` - UUID primary key
- `session_id` - Foreign key to chat_sessions
- `role` - Message role (user/assistant/system)
- `content` - JSONB content
- `tool_calls` - JSONB tool use blocks
- `created_at` - Creation timestamp

### Knowledge Bases Table
- `id` - UUID primary key
- `user_id` - Foreign key to users
- `name` - Knowledge base name
- `description` - Description
- `created_at` - Creation timestamp

### Knowledge Documents Table
- `id` - UUID primary key
- `kb_id` - Foreign key to knowledge_bases
- `filename` - Document filename
- `content` - Document content
- `status` - Processing status (processing/ready/failed)
- `error_message` - Error message if failed
- `file_size` - File size
- `mime_type` - MIME type
- `created_at` - Creation timestamp

### Knowledge Chunks Table
- `id` - UUID primary key
- `document_id` - Foreign key to knowledge_documents
- `chunk_index` - Chunk position in document
- `content` - Chunk content
- `embedding` - Vector embedding (1536 dimensions)
- `metadata` - Additional metadata
- `created_at` - Creation timestamp

## Indexes

### Performance Indexes
- `users(email)` - For login queries
- `agents(user_id)` - For user's agents
- `chat_sessions(user_id, agent_id)` - For session queries
- `messages(session_id, created_at)` - For message history
- `knowledge_documents(kb_id)` - For document queries
- `knowledge_chunks(document_id)` - For chunk queries
- `knowledge_chunks(embedding)` - HNSW index for vector similarity search
- `knowledge_chunks(metadata)` - GIN index for metadata queries

### JSONB Indexes
- `messages(content)` - GIN index for content queries
- `messages(tool_calls)` - GIN index for tool calls
- `agents(knowledge_ids)` - GIN index for knowledge base lookups

## Environment Variables

```bash
# Database connection
DB_URL="postgres://user:password@localhost:5432/atom_ai?sslmode=disable"

# Alternative: individual variables
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="password"
DB_NAME="atom_ai"
DB_SSLMODE="disable"
```

## Troubleshooting

### pgvector not found

If you get an error about the vector type not being found:

```bash
# Connect to database
psql -U postgres -d atom_ai

# Install extension
CREATE EXTENSION IF NOT EXISTS vector;
```

### Migration conflicts

If you need to force a specific migration version:

```bash
# Force to version 5
go run migrate.go -force 5

# Or using migrate CLI
migrate -path ./migrations -database "$DB_URL" force 5
```

### Rollback issues

If you need to drop all tables and start fresh:

```sql
-- In PostgreSQL
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO public;
```

Then run migrations from the beginning:

```bash
go run migrate.go -up
```

## Testing

To test migrations locally with Docker:

```bash
# Start PostgreSQL with pgvector
docker run --name atom-postgres \n  -e POSTGRES_PASSWORD=password \n  -e POSTGRES_DB=atom_ai \n  -p 5432:5432 \n  -d pgvector/pgvector:pg16

# Run migrations
export DB_URL="postgres://postgres:password@localhost:5432/atom_ai?sslmode=disable"
go run migrate.go -up

# Connect to database
docker exec -it atom-postgres psql -U postgres -d atom_ai
```

## Best Practices

1. **Always create both .up.sql and .down.sql files** - This ensures migrations can be rolled back
2. **Test migrations in development** before applying to production
3. **Use transactions** - Wrap migration changes in BEGIN/COMMIT blocks
4. **Index foreign keys** - Add indexes on frequently queried foreign key columns
5. **Consider data integrity** - Use appropriate constraints (NOT NULL, CHECK, UNIQUE)
6. **Document changes** - Add comments to tables and columns
7. **Version control** - Commit migration files to git