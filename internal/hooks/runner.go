package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const defaultTimeoutSec = 30

// run executes a single HookCommand and returns its Result.
func run(ctx context.Context, cmd HookCommand, input Input) (Result, error) {
	timeout := time.Duration(cmd.Timeout) * time.Second
	if timeout <= 0 {
		timeout = defaultTimeoutSec * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	switch cmd.Type {
	case TypeCommand:
		return runShell(ctx, cmd.Command, input)
	case TypeHTTP:
		return runHTTP(ctx, cmd.URL, cmd.Headers, input)
	default:
		return Result{}, fmt.Errorf("unknown hook type: %q", cmd.Type)
	}
}

// runShell executes a shell command, passing hook input as JSON via stdin
// and HOOK_INPUT env var.
func runShell(ctx context.Context, command string, input Input) (Result, error) {
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return Result{}, fmt.Errorf("marshal hook input: %w", err)
	}

	c := exec.CommandContext(ctx, "bash", "-c", command)
	c.Env = append(c.Environ(), fmt.Sprintf("HOOK_INPUT=%s", inputJSON))
	c.Stdin = bytes.NewReader(inputJSON)

	var out bytes.Buffer
	c.Stdout = &out
	c.Stderr = &out

	if err := c.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return Result{Output: out.String()}, fmt.Errorf("hook timed out: %s", command)
		}
		// Non-zero exit is treated as a deny decision.
		return parseHookOutput(out.String(), true), nil
	}
	return parseHookOutput(out.String(), false), nil
}

// runHTTP POSTs hook input as JSON to the given URL.
func runHTTP(ctx context.Context, url string, headers map[string]string, input Input) (Result, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return Result{}, fmt.Errorf("marshal hook input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return Result{}, fmt.Errorf("build http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Result{Output: err.Error()}, fmt.Errorf("http hook request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	isDeny := resp.StatusCode >= 400

	return parseHookOutput(string(respBody), isDeny), nil
}

// parseHookOutput interprets hook stdout/body.
// If the output is JSON with a "decision" key, it is honoured.
// Otherwise a non-zero exit / 4xx response is treated as deny.
func parseHookOutput(output string, defaultDeny bool) Result {
	trimmed := strings.TrimSpace(output)

	// Try to parse JSON response from hook.
	var parsed struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
		if parsed.Decision != "" {
			return Result{
				Decision: parsed.Decision,
				Reason:   parsed.Reason,
				Output:   output,
			}
		}
	}

	if defaultDeny {
		reason := trimmed
		if reason == "" {
			reason = "hook denied the request"
		}
		return Result{Decision: "deny", Reason: reason, Output: output}
	}
	return Result{Output: output}
}
