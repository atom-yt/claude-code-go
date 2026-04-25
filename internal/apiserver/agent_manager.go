package apiserver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/atom-yt/claude-code-go/internal/agent"
)

// AgentFactory creates new agent instances.
type AgentFactory func(ctx context.Context, sessionID string) (*agent.Agent, error)

// AgentManager manages multiple agent instances based on deployment mode.
type AgentManager struct {
	mode   DeploymentMode
	single *agent.Agent           // Single agent instance
	pool   []*AgentWrapper        // Pool of agents
	poolIdx int                    // Round-robin index
	mu      sync.RWMutex

	// Factory for creating new agents
	factory AgentFactory
}

// AgentWrapper wraps an agent with usage tracking.
type AgentWrapper struct {
	agent   *agent.Agent
	session string
	busy    bool
	mu       sync.Mutex
	requests int
	lastUsed time.Time
}

// NewAgentManager creates a new agent manager.
func NewAgentManager(mode DeploymentMode, factory AgentFactory) *AgentManager {
	return &AgentManager{
		mode:    mode,
		factory:  factory,
	}
}

// GetAgent gets an agent instance based on deployment mode.
func (am *AgentManager) GetAgent(ctx context.Context, sessionID string) (*agent.Agent, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	switch am.mode {
	case ModeSingle:
		// Single mode: all requests use same agent
		if am.single == nil {
			agent, err := am.factory(ctx, sessionID)
			if err != nil {
				return nil, fmt.Errorf("failed to create agent: %w", err)
			}
			am.single = agent
		}
		return am.single, nil

	case ModePerSession:
		// Per-session mode: each session gets its own agent
		// This is handled at the server level - the caller should
		// manage the agent lifecycle per session
		return am.factory(ctx, sessionID)

	case ModePool:
		// Pool mode: use round-robin to get an available agent
		if len(am.pool) == 0 {
			return nil, fmt.Errorf("agent pool is empty")
		}

		// Find next available agent (simple round-robin)
		for i := 0; i < len(am.pool); i++ {
			idx := (am.poolIdx + i) % len(am.pool)
			wrapper := am.pool[idx]

			wrapper.mu.Lock()
			busy := wrapper.busy
			wrapper.mu.Unlock()

			if !busy {
				// Mark as busy
				wrapper.mu.Lock()
				wrapper.busy = true
				wrapper.requests++
				wrapper.lastUsed = time.Now()
				wrapper.session = sessionID
				wrapper.mu.Unlock()

				// Update round-robin index
				am.poolIdx = (idx + 1) % len(am.pool)

				return wrapper.agent, nil
			}
		}

		// All agents busy
		return nil, fmt.Errorf("all agents in pool are busy")

	default:
		return nil, fmt.Errorf("unknown deployment mode: %s", am.mode)
	}
}

// ReleaseAgent releases an agent back to the pool.
func (am *AgentManager) ReleaseAgent(agent *agent.Agent, sessionID string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.mode == ModeSingle {
		return
	}

	if am.mode == ModePool {
		// Find and release the agent wrapper
		for _, wrapper := range am.pool {
			if wrapper.agent == agent && wrapper.session == sessionID {
				wrapper.mu.Lock()
				wrapper.busy = false
				wrapper.session = ""
				wrapper.mu.Unlock()
				return
			}
		}
	}

	// Per-session mode: caller manages the agent
}

// InitializePool initializes the agent pool.
func (am *AgentManager) InitializePool(ctx context.Context, size int) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.mode != ModePool {
		return nil
	}

	if size <= 0 {
		size = 4 // Default pool size
	}

	am.pool = make([]*AgentWrapper, 0, size)

	// Create agents for the pool
	for i := 0; i < size; i++ {
		agent, err := am.factory(ctx, fmt.Sprintf("pool-agent-%d", i))
		if err != nil {
			return fmt.Errorf("failed to create pool agent %d: %w", i, err)
		}

		am.pool = append(am.pool, &AgentWrapper{
			agent:   agent,
			busy:    false,
			requests: 0,
			lastUsed: time.Now(),
		})
	}

	return nil
}

// Stats returns agent manager statistics.
func (am *AgentManager) Stats() map[string]any {
	am.mu.RLock()
	defer am.mu.RUnlock()

	stats := map[string]any{
		"mode":  string(am.mode),
		"pool_size":  len(am.pool),
	}

	if am.mode == ModePool {
		busyCount := 0
		totalRequests := 0
		for _, wrapper := range am.pool {
			wrapper.mu.Lock()
			if wrapper.busy {
				busyCount++
			}
			totalRequests += wrapper.requests
			wrapper.mu.Unlock()
		}

		stats["busy_agents"] = busyCount
		stats["available_agents"] = len(am.pool) - busyCount
		stats["total_requests"] = totalRequests
	}

	return stats
}

// Close cleans up all agent instances.
func (am *AgentManager) Close() {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Close single agent
	if am.single != nil {
		// No explicit close method on agent
		// Just nil the reference
		am.single = nil
	}

	// Close pool agents
	// Note: Agents don't have explicit Close in the current implementation
	// This would be a placeholder for future cleanup
	am.pool = nil
}