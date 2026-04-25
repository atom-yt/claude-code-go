package feishu

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting for API calls and message processing.
type RateLimiter struct {
	mu          sync.RWMutex
	limiter     *rate.Limiter
	userLimits  map[string]*rate.Limiter
	burst       int
	rps         int // requests per second
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(rps, burst int) *RateLimiter {
	return &RateLimiter{
		limiter:    rate.NewLimiter(rate.Limit(rps), burst),
		userLimits: make(map[string]*rate.Limiter),
		burst:      burst,
		rps:        rps,
	}
}

// Wait waits until the rate limiter allows a request.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}

// Allow checks if a request is allowed without waiting.
func (rl *RateLimiter) Allow() bool {
	return rl.limiter.Allow()
}

// WaitUser waits until the user-specific rate limiter allows a request.
func (rl *RateLimiter) WaitUser(ctx context.Context, userID string) error {
	limiter := rl.getUserLimiter(userID)
	return limiter.Wait(ctx)
}

// AllowUser checks if a user-specific request is allowed without waiting.
func (rl *RateLimiter) AllowUser(userID string) bool {
	limiter := rl.getUserLimiter(userID)
	return limiter.Allow()
}

// getUserLimiter gets or creates a user-specific limiter.
func (rl *RateLimiter) getUserLimiter(userID string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.userLimits[userID]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := rl.userLimits[userID]; exists {
		return limiter
	}

	// Create new user limiter (same limits as global)
	limiter = rate.NewLimiter(rate.Limit(rl.rps), rl.burst)
	rl.userLimits[userID] = limiter

	return limiter
}

// ClearUser removes the rate limiter for a specific user.
func (rl *RateLimiter) ClearUser(userID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.userLimits, userID)
}

// CleanupInactive removes rate limiters for inactive users.
func (rl *RateLimiter) CleanupInactive(activeUsers map[string]struct{}) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for userID := range rl.userLimits {
		if _, active := activeUsers[userID]; !active {
			delete(rl.userLimits, userID)
		}
	}
}

// Stats returns rate limiter statistics.
func (rl *RateLimiter) Stats() RateLimiterStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return RateLimiterStats{
		RPS:          rl.rps,
		Burst:        rl.burst,
		UserLimiters: len(rl.userLimits),
	}
}

// RateLimiterStats contains rate limiter statistics.
type RateLimiterStats struct {
	RPS          int
	Burst        int
	UserLimiters int
}