package api

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RetryConfig controls automatic retry behaviour for transient HTTP errors.
type RetryConfig struct {
	MaxRetries int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

// DefaultRetryConfig is used when no custom config is supplied.
var DefaultRetryConfig = RetryConfig{
	MaxRetries:   3,
	InitialDelay: 1 * time.Second,
	MaxDelay:     30 * time.Second,
}

// retryable returns true for status codes that should be retried.
func retryable(code int) bool {
	switch code {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError,  // 500
		http.StatusBadGateway,           // 502
		http.StatusServiceUnavailable,   // 503
		http.StatusGatewayTimeout:       // 504
		return true
	}
	return false
}

// retryDelay computes the wait time for a given attempt, honouring Retry-After.
func retryDelay(resp *http.Response, attempt int, cfg RetryConfig) time.Duration {
	// Prefer Retry-After header if present.
	if resp != nil {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(strings.TrimSpace(ra)); err == nil && secs > 0 {
				d := time.Duration(secs) * time.Second
				if d > cfg.MaxDelay {
					d = cfg.MaxDelay
				}
				return d
			}
		}
	}
	// Exponential backoff: initialDelay * 2^attempt
	d := time.Duration(float64(cfg.InitialDelay) * math.Pow(2, float64(attempt)))
	if d > cfg.MaxDelay {
		d = cfg.MaxDelay
	}
	return d
}

// doWithRetry executes an HTTP request with automatic retries for transient errors.
// buildReq is called on each attempt to produce a fresh *http.Request (body must be re-readable).
// On success (200 OK) the *http.Response is returned with body open.
// On non-retryable failure or exhausted retries, an error is returned.
func doWithRetry(ctx context.Context, client *http.Client, buildReq func() (*http.Request, error), cfg RetryConfig) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		req, err := buildReq()
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request: %w", err)
			// Network errors are retryable.
			if attempt < cfg.MaxRetries {
				wait := retryDelay(nil, attempt, cfg)
				if !sleep(ctx, wait) {
					return nil, ctx.Err()
				}
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		// Read body for error message, then close.
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastErr = fmt.Errorf("API error %d: %s", resp.StatusCode, string(b))

		if !retryable(resp.StatusCode) || attempt >= cfg.MaxRetries {
			return nil, lastErr
		}

		wait := retryDelay(resp, attempt, cfg)
		if !sleep(ctx, wait) {
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}

// sleep waits for d or until ctx is cancelled. Returns false if cancelled.
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
