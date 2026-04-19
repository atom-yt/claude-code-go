package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/memory"
)

// ---- /memory ----

type memoryCmd struct{}

func (c *memoryCmd) Name() string      { return "memory" }
func (c *memoryCmd) Aliases() []string { return []string{"mem"} }
func (c *memoryCmd) Description() string {
	return "Show memory files and directory info"
}

func (c *memoryCmd) Execute(_ context.Context, args []string, _ *Context) (string, error) {
	// Get memory directory
	memoryDir, err := memory.MemoryRootDir()
	if err != nil {
		return fmt.Sprintf("Memory not available: %v\n\nMemory is only available in git repositories.", err), nil
	}

	// Show memory directory info
	lockPath, _ := memory.LockFilePath()
	runningLockPath, _ := memory.RunningLockFilePath()

	lines := []string{fmt.Sprintf("Memory directory: %s", memoryDir)}
	lines = append(lines, "")

	// Check for consolidation lock
	lockInfo := "Not locked"
	if info, err := os.Stat(lockPath); err == nil {
		lockInfo = fmt.Sprintf("Locked (since %s)", info.ModTime().Format("2006-01-02 15:04"))
	}
	lines = append(lines, fmt.Sprintf("Consolidation lock: %s", lockInfo))

	if info, err := os.Stat(runningLockPath); err == nil {
		lines = append(lines, fmt.Sprintf("Consolidation in progress since: %s", info.ModTime().Format("2006-01-02 15:04")))
	}

	// List memory files
	entries, err := os.ReadDir(memoryDir)
	if err != nil {
		return fmt.Sprintf("Failed to read memory directory: %v", err), nil
	}

	if len(entries) == 0 {
		lines = append(lines, "")
		lines = append(lines, "No memory files found.")
		return "<!-- raw -->\n" + strings.Join(lines, "\n"), nil
	}

	lines = append(lines, "")
	lines = append(lines, "Memory files:")

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		// Skip lock files
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		lines = append(lines, fmt.Sprintf("  %s (%s, %s)",
			e.Name(),
			formatFileSize(info.Size()),
			info.ModTime().Format("2006-01-02 15:04")))
	}

	return "<!-- raw -->\n" + strings.Join(lines, "\n"), nil
}

func formatFileSize(bytes int64) string {
	const kb = 1024
	const mb = kb * 1024

	if bytes < kb {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < mb {
		return fmt.Sprintf("%.1f KB", float64(bytes)/kb)
	}
	return fmt.Sprintf("%.2f MB", float64(bytes)/mb)
}