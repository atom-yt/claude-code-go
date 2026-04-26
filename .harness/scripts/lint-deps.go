// lint-deps checks Go import statements against layer architecture rules.
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Layer mapping based on .harness/docs/ARCHITECTURE.md
var layers = map[string]int{
	"pkg/":                       0,
	"internal/messages/":         1,
	"internal/pathutil/":         1,
	"internal/urlutil/":          1,
	"internal/config/":           2,
	"internal/permissions/":      2,
	"internal/hooks/":            2,
	"internal/memory/":           2,
	"internal/providers/":        2,
	"internal/compact/":          2,
	"internal/mcpresource/":       2,
	"internal/plugin/":           2,
	"internal/sandbox/":          2,
	"internal/interfaces/":       2,
	"internal/tools/":            3,
	"internal/api/":              3,
	"internal/commands/":        3,
	"internal/skills/":          3,
	"internal/mcp/":             3,
	"internal/prompt/":          3,
	"internal/cmdutil/":         3,
	"internal/agent/":           4,
	"internal/tui/":             4,
	"internal/session/":          4,
	"internal/subagent/":         4,
	"internal/runtime/":          4,
	"internal/taskstore/":        4,
	"internal/apiserver/":       4,
}

// getLayer returns the layer number for a given path, or -1 if not in layers
func getLayer(path string) int {
	normalized := filepath.ToSlash(path)
	// Remove leading "./" if present
	if strings.HasPrefix(normalized, "./") {
		normalized = normalized[2:]
	}

	// Find the longest matching layer prefix
	for prefix, layer := range layers {
		if strings.HasPrefix(normalized, prefix) {
			return layer
		}
	}
	return -1
}

// checkFile checks imports in a single Go file
func checkFile(filePath string, srcLayer int) ([]string, error) {
	var violations []string

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	for _, imp := range node.Imports {
		// Extract import path (remove quotes)
		importPath := strings.Trim(imp.Path.Value, `"`)

		// Skip external and stdlib imports
		if !strings.HasPrefix(importPath, "github.com/atom-yt/claude-code-go") {
			continue
		}

		// Convert import path to directory path
		relPath := strings.TrimPrefix(importPath, "github.com/atom-yt/claude-code-go/")

		dstLayer := getLayer(relPath)
		if dstLayer == -1 {
			continue
		}

		// Check layer rule: srcLayer should be >= dstLayer
		// Higher layer (greater number) can import lower layer (smaller number)
		if srcLayer < dstLayer {
			pos := fset.Position(imp.Pos())
			violations = append(violations,
				fmt.Sprintf("%s:%d: layer violation: %s (L%d) imports %s (L%d)",
					filePath, pos.Line, filepath.Dir(filePath), srcLayer, relPath, dstLayer))
		}
	}

	return violations, nil
}

// findGoFiles finds all .go files recursively
func findGoFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, .git, and hidden directories
		if info.IsDir() {
			name := filepath.Base(path)
			if name == "vendor" || name == ".git" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func main() {
	// Find all Go files
	goFiles, err := findGoFiles(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding Go files: %v\n", err)
		os.Exit(1)
	}

	var allViolations []string

	// Check each file
	for _, filePath := range goFiles {
		// Get source layer
		dir := filepath.Dir(filePath)
		srcLayer := getLayer(dir)
		if srcLayer == -1 {
			continue
		}

		violations, err := checkFile(filePath, srcLayer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking %s: %v\n", filePath, err)
			continue
		}
		allViolations = append(allViolations, violations...)
	}

	if len(allViolations) > 0 {
		fmt.Println("Layer architecture violations found:")
		for _, v := range allViolations {
			fmt.Println("  " + v)
		}
		fmt.Printf("\nTotal violations: %d\n", len(allViolations))
		os.Exit(1)
	}

	fmt.Println("✓ No layer architecture violations found")
	os.Exit(0)
}