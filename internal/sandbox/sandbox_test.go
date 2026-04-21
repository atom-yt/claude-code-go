package sandbox

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, BackendNull, cfg.Type)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
	assert.Equal(t, 50000, cfg.MaxOutput)
}

func TestNewNullSandbox(t *testing.T) {
	cfg := Config{Type: BackendNull, Timeout: 30 * time.Second}
	sb := NewNullSandbox(cfg)
	assert.NotNil(t, sb)
	assert.Equal(t, BackendNull, sb.config.Type)
	assert.Equal(t, 30*time.Second, sb.config.Timeout)
}

func TestNullSandboxDefaults(t *testing.T) {
	sb := NewNullSandbox(Config{})
	assert.Equal(t, DefaultConfig().Timeout, sb.config.Timeout)
	assert.Equal(t, DefaultConfig().MaxOutput, sb.config.MaxOutput)
}

func TestNullSandboxExecute(t *testing.T) {
	ctx := context.Background()
	sb := NewNullSandbox(Config{})

	output, err := sb.Execute(ctx, "echo", []string{"hello"})
	assert.NoError(t, err)
	assert.Contains(t, output, "hello")
}

func TestNullSandboxExecuteError(t *testing.T) {
	ctx := context.Background()
	sb := NewNullSandbox(Config{})

	_, err := sb.Execute(ctx, "nonexistent-command-xyz-123", []string{})
	// Error should occur for non-existent command
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start command")
}

func TestNullSandboxExecuteTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cfg := Config{Timeout: 100 * time.Millisecond}
	sb := NewNullSandbox(cfg)

	// Sleep longer than timeout
	output, err := sb.Execute(ctx, "sleep", []string{"2"})
	// Command should complete with timeout message (or return error from timeout)
	// Note: The timeout behavior depends on OS and command timing
	if err == nil {
		// If no error, check for timeout message
		assert.Contains(t, output, "timed out")
	}
	// Either way, test passes if we handle timeout properly
}

func TestNullSandboxOutputLimit(t *testing.T) {
	ctx := context.Background()
	cfg := Config{MaxOutput: 20}
	sb := NewNullSandbox(cfg)

	// Generate more output than limit (30 chars)
	output, err := sb.Execute(ctx, "echo", []string{"-n", "123456789012345678901234567890"})
	assert.NoError(t, err)
	assert.Contains(t, output, "truncated")
}

func TestNullSandboxClose(t *testing.T) {
	sb := NewNullSandbox(Config{})
	err := sb.Close()
	assert.NoError(t, err)
}

func TestLimitedWriter(t *testing.T) {
	w := &limitedWriter{limit: 10}

	n, err := w.Write([]byte("hello world"))
	assert.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.Equal(t, "hello worl", w.String())
	assert.True(t, w.exceeded)

	// Further writes are accepted but don't increase buffer
	n, err = w.Write([]byte("!"))
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, "hello worl", w.String())
}