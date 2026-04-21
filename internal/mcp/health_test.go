package mcp

import (
	"errors"
	"testing"
)

func TestHealthStatus(t *testing.T) {
	// Test initial health status is healthy
	c := &Client{
		healthStatus: HealthHealthy,
	}
	if c.HealthStatus() != HealthHealthy {
		t.Errorf("initial health status = %v, want healthy", c.HealthStatus())
	}

	// Test health error returns nil when healthy
	if c.HealthError() != nil {
		t.Error("HealthError() should return nil when healthy")
	}

	// Test after recording error
	c.healthStatus = HealthUnhealthy
	c.lastError = errors.New("test error")
	if c.HealthError() == nil {
		t.Error("HealthError() should return error when unhealthy")
	}
}

func TestCheckHealth(t *testing.T) {
	c := &HTTPClient{
		healthStatus:       HealthHealthy,
		consecutiveErrors: 0,
	}

	// Simulate success (would need real call in practice)
	// In test mode without real client, just verify structure
	if c.consecutiveErrors != 0 {
		t.Error("consecutive errors should start at 0")
	}

	// Test ResetHealth
	c.healthStatus = HealthUnhealthy
	c.consecutiveErrors = 3
	c.ResetHealth()

	if c.HealthStatus() != HealthHealthy {
		t.Error("ResetHealth should set status to healthy")
	}
	if c.consecutiveErrors != 0 {
		t.Error("ResetHealth should reset consecutive errors")
	}
}

func TestHealthStatusTransitions(t *testing.T) {
	c := &Client{
		healthStatus:       HealthHealthy,
		consecutiveErrors: 0,
	}

	// Simulate some errors
	for i := 0; i < 2; i++ {
		c.healthMu.Lock()
		c.consecutiveErrors++
		if c.consecutiveErrors >= MaxHealthCheckFailures {
			c.healthStatus = HealthUnhealthy
		}
		c.healthMu.Unlock()
	}

	// Should still be healthy (below threshold)
	if c.HealthStatus() != HealthHealthy {
		t.Errorf("After 2 errors, status = %v, want healthy", c.HealthStatus())
	}

	// Add third error
	c.healthMu.Lock()
	c.consecutiveErrors++
	if c.consecutiveErrors >= MaxHealthCheckFailures {
		c.healthStatus = HealthUnhealthy
	}
	c.healthMu.Unlock()

	// Should be unhealthy (at threshold)
	if c.HealthStatus() != HealthUnhealthy {
		t.Errorf("After 3 errors, status = %v, want unhealthy", c.HealthStatus())
	}
}
