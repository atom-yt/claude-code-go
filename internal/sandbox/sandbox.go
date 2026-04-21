// Package sandbox provides isolation abstraction for tool execution.
package sandbox

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

// BackendType defines the isolation level.
type BackendType string

const (
	BackendNull      BackendType = "null"      // No isolation, direct execution
	BackendContainer  BackendType = "container"  // Docker/container isolation (not implemented)
	BackendRestricted BackendType = "restricted" // Restricted environment (not implemented)
)

// Config holds sandbox configuration.
type Config struct {
	Type      BackendType
	Timeout   time.Duration
	MaxOutput int // Maximum output bytes to capture
}

// DefaultConfig returns safe defaults.
func DefaultConfig() Config {
	return Config{
		Type:      BackendNull,
		Timeout:   60 * time.Second,
		MaxOutput: 50000, // 50KB limit (same as current implementation)
	}
}

// Sandbox defines the interface for isolated command execution.
type Sandbox interface {
	// Execute runs a command with the sandbox's isolation policy.
	Execute(ctx context.Context, command string, args []string) (string, error)

	// Close cleans up any resources used by the sandbox.
	Close() error
}

// NullSandbox provides no isolation - executes commands directly.
// Useful for development and trusted environments.
type NullSandbox struct {
	config Config
}

// NewNullSandbox creates a sandbox with no isolation.
func NewNullSandbox(cfg Config) *NullSandbox {
	if cfg.Type == "" {
		cfg.Type = BackendNull
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultConfig().Timeout
	}
	if cfg.MaxOutput == 0 {
		cfg.MaxOutput = DefaultConfig().MaxOutput
	}
	return &NullSandbox{config: cfg}
}

// Execute runs the command directly without isolation.
func (s *NullSandbox) Execute(ctx context.Context, command string, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start command: %w", err)
	}

	// Combine stdout and stderr with timeout
	done := make(chan struct{})
	go func() {
		io.Copy(io.Discard, stderr)
		done <- struct{}{}
	}()

	output := &limitedWriter{limit: s.config.MaxOutput}
	_, err = io.Copy(output, stdout)
	<-done

	cmd.Wait()

	if output.exceeded {
		return "[output truncated]", nil
	}

	// Check for context timeout
	if ctx.Err() == context.DeadlineExceeded {
		return output.String() + fmt.Sprintf("\n[Command timed out after %.0fs]", s.config.Timeout.Seconds()), nil
	}

	return output.String(), err
}

// Close is a no-op for NullSandbox.
func (s *NullSandbox) Close() error {
	return nil
}

// limitedWriter limits the number of bytes written.
type limitedWriter struct {
	buffer   []byte
	limit    int
	exceeded bool
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.exceeded {
		return len(p), nil // Silently accept but don't store
	}
	remaining := w.limit - len(w.buffer)
	if len(p) > remaining {
		w.buffer = append(w.buffer, p[:remaining]...)
		w.exceeded = true
		return len(p), nil
	}
	w.buffer = append(w.buffer, p...)
	return len(p), nil
}

func (w *limitedWriter) String() string {
	return string(w.buffer)
}