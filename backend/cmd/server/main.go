package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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