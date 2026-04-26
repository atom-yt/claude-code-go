package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"

	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
	"github.com/atom-yt/atom-ai-platform/backend/internal/db"
	"github.com/atom-yt/atom-ai-platform/backend/internal/handlers"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
	"github.com/atom-yt/atom-ai-platform/backend/internal/services"
	"github.com/atom-yt/claude-code-go/pkg/agent"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get configuration
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/atom_ai?sslmode=disable")
	serverAddr := getEnv("SERVER_ADDR", ":8080")

	// AI API configuration
	apiKey := getEnv("ANTHROPIC_API_KEY", "")
	baseURL := getEnv("BASE_URL", "")
	defaultProvider := getEnv("DEFAULT_PROVIDER", "anthropic")
	defaultModel := getEnv("DEFAULT_MODEL", "claude-sonnet-4-6")

	// Initialize database
	database, err := db.NewFromURL(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("Database connected successfully")

	// Run migrations (simplified: execute SQL directly)
	if err := runMigrations(database.Pool()); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(database.Pool())
	sessionRepo := repository.NewSessionRepository(database.Pool())
	messageRepo := repository.NewMessageRepository(database.Pool())
	agentRepo := repository.NewAgentRepository(database.Pool())
	skillRepo := repository.NewSkillRepository(database.Pool())
	artifactRepo := repository.NewArtifactRepository(database.Pool())
	scheduleRepo := repository.NewScheduleRepository(database.Pool())
	knowledgeRepo := repository.NewKnowledgeRepository(database.Pool())

	// Initialize services
	authUserStore := auth.NewUserStoreAdapter(userRepo)
	authService := auth.NewService(jwtSecret, authUserStore)
	agentService := services.NewAgentService(agentRepo)
	sessionService := services.NewSessionService(sessionRepo, agentRepo)
	messageService := services.NewMessageService(messageRepo, sessionRepo)
	skillService := services.NewSkillService(skillRepo)
	artifactService := services.NewArtifactService(artifactRepo)
	scheduleService := services.NewScheduleService(scheduleRepo)
	knowledgeService := services.NewKnowledgeService(knowledgeRepo)

	// Initialize agent factory
	agentFactory := agent.NewConfigFactory(apiKey, baseURL, defaultProvider, defaultModel)

	// Initialize router
	router := handlers.NewRouter(&handlers.Config{
		AuthService:      authService,
		AgentService:     agentService,
		SessionService:   sessionService,
		MessageService:   messageService,
		SkillService:     skillService,
		ArtifactService:  artifactService,
		ScheduleService:  scheduleService,
		KnowledgeService: knowledgeService,
		AgentFactory:     agentFactory,
	})

	// Create server
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router.GetRouter(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// runMigrations executes SQL migration files directly
func runMigrations(db *pgxpool.Pool) error {
	// Create tables if they don't exist
	tables := map[string]string{
		"users": `CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			display_name VARCHAR(255),
			role VARCHAR(50) DEFAULT 'user',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		"agents": `CREATE TABLE IF NOT EXISTS agents (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description VARCHAR(500),
			system_prompt TEXT,
			model VARCHAR(255) NOT NULL,
			provider VARCHAR(50) NOT NULL,
			temperature NUMERIC DEFAULT 0.7,
			max_tokens INT DEFAULT 4096,
			tools JSONB DEFAULT '[]'::jsonb,
			knowledge_ids JSONB DEFAULT '[]'::jsonb,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		"chat_sessions": `CREATE TABLE IF NOT EXISTS chat_sessions (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			agent_id VARCHAR(255),
			title VARCHAR(500) DEFAULT 'New Chat',
			status VARCHAR(50) DEFAULT 'active',
			messages JSONB DEFAULT '[]',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		"messages": `CREATE TABLE IF NOT EXISTS messages (
			id VARCHAR(255) PRIMARY KEY,
			session_id VARCHAR(255) NOT NULL,
			role VARCHAR(50) DEFAULT 'user',
			content TEXT,
			tool_calls JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		"skills": `CREATE TABLE IF NOT EXISTS skills (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			category VARCHAR(50),
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		"artifacts": `CREATE TABLE IF NOT EXISTS artifacts (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			session_id VARCHAR(255),
			content TEXT NOT NULL,
			type VARCHAR(50) DEFAULT 'code',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		"scheduled_tasks": `CREATE TABLE IF NOT EXISTS scheduled_tasks (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			session_id VARCHAR(255),
			name VARCHAR(255) NOT NULL,
			message TEXT,
			cron_expression VARCHAR(100),
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		"knowledge_bases": `CREATE TABLE IF NOT EXISTS knowledge_bases (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		"knowledge_documents": `CREATE TABLE IF NOT EXISTS knowledge_documents (
			id VARCHAR(255) PRIMARY KEY,
			base_id VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		"knowledge_chunks": `CREATE TABLE IF NOT EXISTS knowledge_chunks (
			id VARCHAR(255) PRIMARY KEY,
			document_id VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for tableName, sql := range tables {
		if _, err := db.Exec(context.Background(), sql); err != nil {
			log.Printf("Failed to create table %s: %v", tableName, err)
		} else {
			log.Printf("Created table: %s", tableName)
		}
	}

	return nil
}