package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/apiserver"
	"github.com/atom-yt/claude-code-go/internal/config"
	"github.com/atom-yt/claude-code-go/internal/interfaces"
	_ "github.com/atom-yt/claude-code-go/internal/interfaces/feishu" // Auto-register feishu provider
	"github.com/atom-yt/claude-code-go/internal/tools"
	"github.com/spf13/cobra"
)

func main() {
	// Root command
	rootCmd := &cobra.Command{
		Use:   "gateway",
		Short: "claude-code-go interface provider gateway",
		Long: `Gateway manages interface providers (Feishu, WeChat, etc.) for Claude Code.

Providers are automatically discovered through registration. To add a new provider,
implement the interfaces.Provider interface and register it via init().

Configuration is loaded from:
  - Project-level .claude/settings.json
  - User-level ~/.claude/settings.json

Example settings.json:
  {
    "interfaces": {
      "feishu": {
        "enabled": true,
        "config": {
          "app_id": "cli_xxx",
          "app_secret": "xxx",
          "mode": "dual"
        }
      }
    }
  }
`,
	}

	// List providers command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available interface providers",
		Long: `List all registered interface providers.

Providers marked with * are enabled in configuration.`,
		Run: listProviders,
	}
	rootCmd.AddCommand(listCmd)

	// Start command
	startCmd := &cobra.Command{
		Use:   "start [provider...]",
		Short: "Start interface provider(s)",
		Long: `Start one or more interface providers.

If no provider name is given, starts all providers that are enabled in configuration.
If provider names are given, starts only those providers.

Examples:
  gateway start              # Start all enabled providers
  gateway start feishu       # Start only feishu provider
  gateway start feishu wechat # Start feishu and wechat providers`,
		Args: cobra.ArbitraryArgs,
		Run:  startProviders,
	}
	rootCmd.AddCommand(startCmd)

	// API server command
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Start HTTP API server for external access",
		Long: `Start an HTTP API server for external services to interact with agents.

The API server supports multiple deployment modes:
  - single: All requests use a single agent (lowest resource usage)
  - per-session: Each session gets its own agent (full isolation)
  - pool: Uses a pool of agents with load balancing (production use)

Endpoints:
  GET    /health                - Health check
  POST   /api/v1/sessions       - Create new session
  GET    /api/v1/sessions/{id}  - Get session info
  DELETE /api/v1/sessions/{id}  - Delete session
  POST   /api/v1/chat/completions - Chat (REST API)
  POST   /api/v1/chat/stream    - Chat (Server-Sent Events)

Examples:
  gateway api                           # Start API server on :8080
  gateway api --addr :3000              # Start on port 3000
  gateway api --mode pool --pool-size 4 # Use pool mode with 4 agents
  gateway api --auth --api-key secret   # Enable authentication`,
		Run: startAPIServer,
	}
	apiCmd.Flags().String("addr", ":8080", "API server listen address")
	apiCmd.Flags().String("mode", "pool", "Deployment mode: single, per-session, or pool")
	apiCmd.Flags().Int("pool-size", 4, "Number of agents in pool (for pool mode)")
	apiCmd.Flags().Bool("auth", false, "Enable API authentication")
	apiCmd.Flags().String("api-key", "", "API key for authentication")
	apiCmd.Flags().Bool("sse", true, "Enable Server-Sent Events")
	apiCmd.Flags().StringSlice("cors", []string{"*"}, "Allowed CORS origins")
	rootCmd.AddCommand(apiCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// listProviders lists all registered interface providers.
func listProviders(cmd *cobra.Command, args []string) {
	// Load configuration to check enabled status
	cfg := config.Load(config.CLIFlags{})

	fmt.Println("Registered Interface Providers:")
	fmt.Println()

	providers := interfaces.List()
	if len(providers) == 0 {
		fmt.Println("  No providers registered")
		fmt.Println()
		fmt.Println("To register a provider, import its package which should")
		fmt.Println("call interfaces.Register() in its init() function.")
		return
	}

	for _, name := range providers {
		provider, ok := interfaces.Get(name)
		if !ok {
			continue
		}

		// Check if enabled in config
		status := "disabled"
		marker := " "
		if interfaceCfg, exists := cfg.Interfaces[name]; exists && interfaceCfg.Enabled {
			status = "enabled"
			marker = "*"
		}

		fmt.Printf("  %s %-12s %s (%s)\n", marker, name, provider.Description(), status)
	}

	fmt.Println()
	fmt.Println("Run 'gateway start' to start all enabled providers")
	fmt.Println("Run 'gateway start <name>' to start a specific provider")
}

// startProviders starts the specified interface provider(s).
func startProviders(cmd *cobra.Command, args []string) {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load configuration
	cfg := config.Load(config.CLIFlags{})

	// Determine which providers to start
	var providersToStart []string
	if len(args) > 0 {
		// Explicit providers specified
		providersToStart = args
	} else {
		// Start all enabled providers from config
		for name, interfaceCfg := range cfg.Interfaces {
			if interfaceCfg.Enabled {
				providersToStart = append(providersToStart, name)
			}
		}
	}

	if len(providersToStart) == 0 {
		fmt.Println("No providers to start.")
		fmt.Println()
		fmt.Println("To enable a provider, add it to settings.json:")
		fmt.Println(`  {
    "interfaces": {
      "feishu": {
        "enabled": true,
        "config": {
          "app_id": "cli_xxx",
          "app_secret": "xxx"
        }
      }
    }
  }`)
		fmt.Println()
		fmt.Println("Run 'gateway list' to see all registered providers")
		return
	}

	// Start each provider
	var startedCount int
	var errors []error

	for _, name := range providersToStart {
		// Get provider from registry
		provider, ok := interfaces.Get(name)
		if !ok {
			fmt.Fprintf(os.Stderr, "Provider not registered: %s\n", name)
			fmt.Fprintf(os.Stderr, "Run 'gateway list' to see available providers\n")
			errors = append(errors, fmt.Errorf("provider not registered: %s", name))
			continue
		}

		// Build provider config
		providerCfg, err := buildProviderConfig(cfg, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] Config error: %v\n", name, err)
			errors = append(errors, fmt.Errorf("[%s] config error: %w", name, err))
			continue
		}

		// Start provider
		fmt.Printf("[%s] Starting provider...\n", name)
		fmt.Printf("[%s] Description: %s\n", name, provider.Description())

		// Start in a goroutine so we can run multiple providers
		go func(pName string, p interfaces.Provider, pc interfaces.ProviderConfig) {
			if err := p.Start(ctx, pc); err != nil {
				fmt.Fprintf(os.Stderr, "[%s] Error: %v\n", pName, err)
			}
		}(name, provider, providerCfg)

		startedCount++
	}

	if startedCount == 0 {
		fmt.Println("No providers started successfully.")
		os.Exit(1)
	}

	fmt.Printf("\nStarted %d provider(s). Press Ctrl+C to stop.\n", startedCount)

	// Wait for context cancellation (Ctrl+C)
	<-ctx.Done()
	fmt.Println("\nShutting down...")
}

// buildProviderConfig builds provider configuration from settings.
func buildProviderConfig(cfg config.Settings, providerName string) (interfaces.ProviderConfig, error) {
	// Get interface provider configuration
	interfaceCfg, exists := cfg.Interfaces[providerName]
	if !exists {
		return interfaces.ProviderConfig{}, fmt.Errorf("provider %s not configured in settings.json", providerName)
	}

	// Check if enabled (only warn, don't block if explicitly requested)
	if !interfaceCfg.Enabled {
		fmt.Printf("[%s] Warning: provider is disabled in config, but starting anyway\n", providerName)
	}

	// Create API client
	apiClient, err := createAPIClient(cfg)
	if err != nil {
		return interfaces.ProviderConfig{}, fmt.Errorf("failed to create API client: %w", err)
	}

	// Create tool registry
	toolReg := tools.NewRegistry()

	// Build provider config
	providerCfg := interfaces.ProviderConfig{
		Model:        cfg.Model,
		SystemPrompt: "",    // Use default
		MaxHistory:   50,    // Default
		Provider:     cfg.Provider,
		APIClient:    apiClient,
		ToolRegistry: toolReg,
		Settings:     interfaceCfg.Config,
	}

	return providerCfg, nil
}

// createAPIClient creates the API client based on configuration.
func createAPIClient(cfg config.Settings) (api.Streamer, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	// Create client based on provider
	switch cfg.Provider {
	case "anthropic", "":
		if cfg.BaseURL != "" {
			return api.NewWithBaseURL(cfg.APIKey, cfg.BaseURL), nil
		}
		return api.New(cfg.APIKey), nil

	case "openai", "kimi", "deepseek", "qwen", "codex", "ark", "ark-openai":
		return api.NewOpenAI(cfg.APIKey, cfg.BaseURL), nil

	case "ark-anthropic":
		if cfg.BaseURL != "" {
			return api.NewWithBaseURL(cfg.APIKey, cfg.BaseURL), nil
		}
		return api.New(cfg.APIKey), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

// startAPIServer starts the HTTP API server.
func startAPIServer(cmd *cobra.Command, args []string) {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load configuration
	cfg := config.Load(config.CLIFlags{})

	// Get flags
	addr, _ := cmd.Flags().GetString("addr")
	modeStr, _ := cmd.Flags().GetString("mode")
	poolSize, _ := cmd.Flags().GetInt("pool-size")
	auth, _ := cmd.Flags().GetBool("auth")
	apiKey, _ := cmd.Flags().GetString("api-key")
	enableSSE, _ := cmd.Flags().GetBool("sse")
	corsOrigins, _ := cmd.Flags().GetStringSlice("cors")

	// Map mode string to enum
	var mode apiserver.DeploymentMode
	switch modeStr {
	case "single":
		mode = apiserver.ModeSingle
	case "per-session":
		mode = apiserver.ModePerSession
	case "pool":
		mode = apiserver.ModePool
	default:
		fmt.Fprintf(os.Stderr, "Invalid mode: %s (use single, per-session, or pool)\n", modeStr)
		os.Exit(1)
	}

	// Validate authentication
	if auth && apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: --api-key required when --auth is enabled")
		os.Exit(1)
	}

	// Create API client
	_, err := createAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
		os.Exit(1)
	}

	// Create tool registry
	_ = tools.NewRegistry()

	// Build API server config
	serverCfg := &apiserver.Config{
		Addr:         addr,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second, // Long timeout for AI responses
		IdleTimeout:  60 * time.Second,

		Model:        cfg.Model,
		Provider:     cfg.Provider,
		BaseURL:      cfg.BaseURL,
		APIKey:       cfg.APIKey,
		SystemPrompt: "",

		MCPServers:      cfg.MCPServers,
		Permissions:     cfg.Permissions,
		Hooks:           cfg.Hooks,

		DeploymentMode: mode,
		PoolSize:       poolSize,

		EnableSSE:   enableSSE,
		EnableAuth:  auth,
		APIKeyAuth: apiKey,

		AllowedOrigins: corsOrigins,
		WorkingDir:    "",
	}

	// Create server
	server := apiserver.NewServer(serverCfg)

	fmt.Printf("Starting API server on %s\n", addr)
	fmt.Printf("Deployment mode: %s\n", mode)
	fmt.Printf("Pool size: %d\n", poolSize)
	if auth {
		fmt.Println("Authentication: enabled")
	} else {
		fmt.Println("Authentication: disabled")
	}
	fmt.Printf("CORS origins: %v\n", corsOrigins)

	// Start server
	if err := server.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nAPI server stopped")
}