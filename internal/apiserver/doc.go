// Package apiserver provides HTTP API server for external Agent access.
//
// This allows external services to interact with Claude Code agents via REST API.
//
// Deployment Modes:
//
// 1. Single Mode: All requests use a single Agent instance
//    - Lowest resource usage
//    - Shared conversation context
//    - Best for single-user or internal tools
//
// 2. Per-Session Mode: Each session gets its own Agent instance
//    - Full isolation between conversations
//    - Higher memory usage
//    - Best for multi-user scenarios
//
// 3. Pool Mode: Uses a pool of Agent instances with round-robin load balancing
//    - Balanced resource usage
//    - Concurrent request handling
//    - Best for high-traffic production use
//
// API Endpoints:
//
//   GET    /health                - Health check
//   POST   /api/v1/sessions       - Create new session
//   GET    /api/v1/sessions/{id}  - Get session info
//   DELETE /api/v1/sessions/{id}  - Delete session
//   POST   /api/v1/chat/completions - Chat (REST API, non-streaming)
//   POST   /api/v1/chat/stream     - Chat (Server-Sent Events for streaming)
//   GET    /api/v1/stats            - Server statistics
//
// Usage:
//
//   cfg := &apiserver.Config{
//       Addr: ":8080",
//       DeploymentMode: apiserver.ModePool,
//       PoolSize: 4,
//       EnableAuth: true,
//       APIKey: "your-api-key",
//   }
//   server := apiserver.NewServer(cfg)
//   server.Start(ctx)
//
// Multi-Instance Deployment:
//
// For production deployments, run multiple API server instances behind
// a load balancer (nginx, AWS ALB, etc.). Each instance can
// use different deployment modes based on your needs.
//
// Session Management:
//
//   Sessions are automatically tracked and cleaned up. In per-session
//   mode, each session has its own agent instance. In pool mode,
//   agents are reused across sessions with proper isolation.
package apiserver