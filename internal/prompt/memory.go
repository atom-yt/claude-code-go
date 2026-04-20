package prompt

import (
	"os"
	"strings"

	"github.com/atom-yt/claude-code-go/internal/memory"
)

const maxMemoryChars = 8_000

// DiscoverMemorySnippet loads the durable project memory index when available.
func DiscoverMemorySnippet() string {
	path, err := memory.MemoryIndexPath()
	if err != nil || path == "" {
		return ""
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return ""
	}
	if len(content) > maxMemoryChars {
		content = content[:maxMemoryChars]
	}
	return content
}
