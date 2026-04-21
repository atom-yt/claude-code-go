package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atom-yt/claude-code-go/internal/agent"
	"github.com/atom-yt/claude-code-go/internal/api"
	"github.com/atom-yt/claude-code-go/internal/config"
	"github.com/atom-yt/claude-code-go/internal/tools"
	toolglob "github.com/atom-yt/claude-code-go/internal/tools/glob"
	toolgrep "github.com/atom-yt/claude-code-go/internal/tools/grep"
	toolread "github.com/atom-yt/claude-code-go/internal/tools/read"
)

// Consolidator manages memory consolidation operations.
type Consolidator struct {
	cfg              config.Settings
	client           api.Streamer
	memDir           string
	minHours         int
	minSessions      int
}

// NewConsolidator creates a new consolidator with the given configuration.
func NewConsolidator(cfg config.Settings, client api.Streamer) (*Consolidator, error) {
	memDir, err := MemoryRootDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory root dir: %w", err)
	}

	// Create memory directory if it doesn't exist
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create memory directory: %w", err)
	}

	minHours := cfg.MinConsolidateHours
	if minHours <= 0 {
		minHours = 24 // Default to 24 hours
	}

	minSessions := cfg.MinConsolidateSessions
	if minSessions <= 0 {
		minSessions = 5 // Default to 5 sessions
	}

	return &Consolidator{
		cfg:         cfg,
		client:      client,
		memDir:      memDir,
		minHours:    minHours,
		minSessions: minSessions,
	}, nil
}

// ShouldConsolidate checks if consolidation should be triggered based on:
// 1. Time since last consolidation (>= minHours)
// 2. Number of sessions since last consolidation (>= minSessions)
// 3. Not currently running
func (c *Consolidator) ShouldConsolidate(ctx context.Context) (bool, string) {
	lockState, err := GetLockState()
	if err != nil {
		return false, fmt.Sprintf("failed to get lock state: %v", err)
	}

	if lockState.IsRunning {
		return false, "consolidation already running"
	}

	// Check time threshold
	timeSinceLast := time.Since(lockState.LastConsolidation)
	if timeSinceLast.Hours() < float64(c.minHours) {
		return false, fmt.Sprintf("too soon since last consolidation (%.1f < %d hours)",
			timeSinceLast.Hours(), c.minHours)
	}

	// Check session count threshold
	if lockState.LastConsolidation.IsZero() {
		// First consolidation - always allow if there are enough sessions
		count, err := CountSessionsSince(time.Time{})
		if err != nil {
			return false, fmt.Sprintf("failed to count sessions: %v", err)
		}
		if count < c.minSessions {
			return false, fmt.Sprintf("not enough sessions yet (%d < %d)", count, c.minSessions)
		}
	} else {
		count, err := CountSessionsSince(lockState.LastConsolidation)
		if err != nil {
			return false, fmt.Sprintf("failed to count sessions: %v", err)
		}
		if count < c.minSessions {
			return false, fmt.Sprintf("not enough new sessions (%d < %d)", count, c.minSessions)
		}
	}

	return true, "ready to consolidate"
}

// Consolidate performs memory consolidation.
// It uses a read-only agent to analyze sessions and memory, then returns
// the suggested memory file content for the caller to write.
func (c *Consolidator) Consolidate(ctx context.Context) (string, error) {
	// Acquire lock
	acquired, err := AcquireLock()
	if err != nil {
		return "", fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !acquired {
		return "", fmt.Errorf("consolidation already in progress")
	}
	defer ReleaseLock()

	// Build consolidation prompt
	prompt, err := BuildConsolidationPrompt()
	if err != nil {
		return "", fmt.Errorf("failed to build consolidation prompt: %w", err)
	}

	// Create a read-only tool registry
	registry := tools.NewRegistry()
	registry.Register(&toolread.Tool{})
	registry.Register(&toolglob.Tool{})
	registry.Register(&toolgrep.Tool{})

	// Create a temporary agent for consolidation
	consolidatorAgent := agent.New(c.client, c.cfg.Model, c.cfg.Provider, registry, nil, nil)

	// Run the consolidation query
	eventCh := consolidatorAgent.Query(ctx, prompt)

	// Collect the full response
	var fullResponse strings.Builder
	for event := range eventCh {
		switch event.Type {
		case agent.EventTextDelta:
			fullResponse.WriteString(event.Text)
		case agent.EventError:
			return "", fmt.Errorf("consolidation query error: %w", event.Error)
		case agent.EventDone:
			// Done collecting
		}
	}

	response := fullResponse.String()

	// Extract memory file updates from the response
	// The agent should output formatted content that can be written to memory files
	// We look for markdown code blocks with file paths
	filesUpdated, err := parseAndWriteMemoryFiles(ctx, response)
	if err != nil {
		return "", fmt.Errorf("failed to write memory files: %w", err)
	}

	result := response
	if filesUpdated > 0 {
		result = fmt.Sprintf("%s\n\n[Consolidation complete: %d memory file(s) updated]", response, filesUpdated)
	}

	return result, nil
}

// parseAndWriteMemoryFiles extracts file paths and content from the agent's response
// and writes them to the memory directory.
// Expected format: ```markdown:filename.md or ```filename
func parseAndWriteMemoryFiles(ctx context.Context, response string) (int, error) {
	memDir, err := MemoryRootDir()
	if err != nil {
		return 0, err
	}

	// Ensure directory exists
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		return 0, err
	}

	filesWritten := 0

	// Parse for markdown code blocks with file specifications
	// Format: ```filename.md or ```markdown:filename.md
	lines := strings.Split(response, "\n")
	var currentFile string
	var currentContent strings.Builder
	inCodeBlock := false

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// End of code block, write file
				if currentFile != "" {
					filePath := filepath.Join(memDir, currentFile)
					if err := os.WriteFile(filePath, []byte(currentContent.String()), 0o644); err != nil {
						return filesWritten, err
					}
					filesWritten++
				}
				currentFile = ""
				currentContent.Reset()
				inCodeBlock = false
			} else {
				// Start of code block, extract filename
				inCodeBlock = true
				blockHeader := strings.TrimPrefix(line, "```")
				blockHeader = strings.TrimSpace(blockHeader)

				// Check for format: ```markdown:filename.md
				if strings.Contains(blockHeader, ":") {
					parts := strings.SplitN(blockHeader, ":", 2)
					if len(parts) == 2 {
						// Handle markdown:filename
						if parts[0] == "markdown" || strings.HasSuffix(parts[0], "markdown") {
							currentFile = parts[1]
						} else {
							currentFile = blockHeader
						}
					}
				} else if blockHeader != "" && !strings.EqualFold(blockHeader, "markdown") {
					// Format: ```filename.md
					currentFile = blockHeader
				}
			}
		} else if inCodeBlock {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	return filesWritten, nil
}

// RunBackgroundConsolidation runs consolidation in the background.
// It logs errors but does not block the caller.
func (c *Consolidator) RunBackgroundConsolidation(ctx context.Context) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Log panic if needed
			}
		}()

		should, _ := c.ShouldConsolidate(ctx)
		if !should {
			return
		}

		// Perform consolidation
		result, err := c.Consolidate(ctx)
		if err != nil {
			return
		}

		// Log or process the result
		// For now, we could write to a log file or trigger a notification
		_ = result
	}()
}

// EnsureMemoryDirExists creates the memory directory if it doesn't exist.
func (c *Consolidator) EnsureMemoryDirExists() error {
	return os.MkdirAll(c.memDir, 0o755)
}

// WriteMemoryFile writes a memory file to the memory directory.
func (c *Consolidator) WriteMemoryFile(filename, content string) error {
	path := filepath.Join(c.memDir, filename)
	return os.WriteFile(path, []byte(content), 0o644)
}

// ListMemoryFiles returns a list of all memory files (excluding hidden files).
func (c *Consolidator) ListMemoryFiles() ([]string, error) {
	entries, err := os.ReadDir(c.memDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || len(e.Name()) == 0 || e.Name()[0] == '.' {
			continue
		}
		files = append(files, e.Name())
	}
	return files, nil
}

// UpdateMemoryIndex updates or creates the MEMORY.md index file.
// It adds a reference to the new memory file.
func UpdateMemoryIndex(newFilename string, description string) error {
	memIndexPath, err := MemoryIndexPath()
	if err != nil {
		return err
	}

	// Read existing content
	var existingContent []byte
	if content, err := os.ReadFile(memIndexPath); err == nil {
		existingContent = content
	}

	// Build new entry
	timestamp := time.Now().Format("2006-01-02")
	newEntry := fmt.Sprintf("\n- [%s](%s) - %s", newFilename, newFilename, description)

	// Write updated content
	var updatedContent strings.Builder
	updatedContent.Write(existingContent)

	if !strings.Contains(string(existingContent), newFilename) {
		// Add timestamp header if needed
		if !strings.Contains(string(existingContent), timestamp) {
			updatedContent.WriteString(fmt.Sprintf("\n## %s\n", timestamp))
		}
		updatedContent.WriteString(newEntry)
	}

	return os.WriteFile(memIndexPath, []byte(updatedContent.String()), 0o644)
}
