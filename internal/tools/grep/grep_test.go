package grep

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTool_Name(t *testing.T) {
	tool := &Tool{}
	assert.Equal(t, "Grep", tool.Name())
}

func TestTool_IsReadOnly(t *testing.T) {
	tool := &Tool{}
	assert.True(t, tool.IsReadOnly())
}

func TestTool_IsConcurrencySafe(t *testing.T) {
	tool := &Tool{}
	assert.True(t, tool.IsConcurrencySafe())
}

func TestTool_Description(t *testing.T) {
	tool := &Tool{}
	desc := tool.Description()
	assert.Contains(t, desc, "regular expression")
	assert.Contains(t, desc, "pattern")
	assert.Contains(t, desc, "context")
}

func TestTool_InputSchema(t *testing.T) {
	tool := &Tool{}
	schema := tool.InputSchema()
	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema["type"])

	props := schema["properties"].(map[string]any)
	assert.Contains(t, props, "pattern")
	assert.Contains(t, props, "path")
	assert.Contains(t, props, "include")
	assert.Contains(t, props, "context")

	required := schema["required"].([]string)
	assert.Contains(t, required, "pattern")
}

func TestTool_Call_MissingPattern(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "pattern is required")
}

func TestTool_Call_InvalidRegex(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{"pattern": "[invalid(regex)"})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "invalid regex")
}

func TestTool_Call_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()

	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{
		"pattern": "xyzzy123",
		"path":    tmpDir,
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "No matches found")
}

func TestTool_Call_SearchExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := strings.Join([]string{
		"line 1",
		"hello world",
		"line 3",
		"hello again",
		"line 5",
	}, "\n")
	assert.NoError(t, os.WriteFile(testFile, []byte(content), 0o644))

	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{
		"pattern": "hello",
		"path":    tmpDir,
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "test.txt")
	assert.Contains(t, result.Output, "hello world")
	assert.Contains(t, result.Output, "hello again")
}

func TestTool_Call_SearchWithInclude(t *testing.T) {
	tmpDir := t.TempDir()

	testGo := filepath.Join(tmpDir, "test.go")
	assert.NoError(t, os.WriteFile(testGo, []byte("package main\nfunc main() {}"), 0o644))

	testTxt := filepath.Join(tmpDir, "test.txt")
	assert.NoError(t, os.WriteFile(testTxt, []byte("package main\nfunc main() {}"), 0o644))

	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{
		"pattern": "package",
		"path":    tmpDir,
		"include": "*.go",
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "test.go")
	assert.NotContains(t, result.Output, "test.txt")
}

func TestTool_Call_SearchWithContext(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := strings.Join([]string{
		"line 1 before",
		"target line",
		"line 3 after",
	}, "\n")
	assert.NoError(t, os.WriteFile(testFile, []byte(content), 0o644))

	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{
		"pattern": "target",
		"path":    tmpDir,
		"context": 1,
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "target line")
}

func TestTool_Call_SearchCurrentDir(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	assert.NoError(t, os.WriteFile(testFile, []byte("search me"), 0o644))

	tool := &Tool{}
	oldWD, _ := os.Getwd()
	defer os.Chdir(oldWD)
	assert.NoError(t, os.Chdir(tmpDir))

	result, err := tool.Call(nil, map[string]any{"pattern": "search"})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "search me")
}

func TestTool_Call_SearchInSubdir(t *testing.T) {
	tmpDir := t.TempDir()
	subdir := filepath.Join(tmpDir, "subdir")
	assert.NoError(t, os.Mkdir(subdir, 0o755))

	testFile := filepath.Join(subdir, "test.txt")
	assert.NoError(t, os.WriteFile(testFile, []byte("nested file"), 0o644))

	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{
		"pattern": "nested",
		"path":    tmpDir,
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "test.txt")
	assert.Contains(t, result.Output, "nested file")
}

func TestTool_Call_IgnoresDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	subdir := filepath.Join(tmpDir, "subdir")
	assert.NoError(t, os.Mkdir(subdir, 0o755))

	testFile := filepath.Join(subdir, "test.txt")
	assert.NoError(t, os.WriteFile(testFile, []byte("test content"), 0o644))

	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{
		"pattern": "test",
		"path":    subdir,
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "test.txt")
}

func TestTool_Call_TruncatesResults(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	var lines []string
	for i := 0; i < 600; i++ {
		lines = append(lines, "match line")
	}
	content := strings.Join(lines, "\n")
	assert.NoError(t, os.WriteFile(testFile, []byte(content), 0o644))

	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{
		"pattern": "match",
		"path":    tmpDir,
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Truncated")
	assert.Contains(t, result.Output, "first 500 matches")
}

func TestTool_Call_InvalidPath(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{
		"pattern": "test",
		"path":    "/nonexistent/path/that/does/not/exist",
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "No matches found")
}

func TestTool_Call_CaseInsensitivePattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	assert.NoError(t, os.WriteFile(testFile, []byte("Hello World hello WORLD"), 0o644))

	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{
		"pattern": "(?i)hello",
		"path":    tmpDir,
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Hello")
	assert.Contains(t, result.Output, "hello")
}

func TestTool_Call_MultipleMatchesInFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := strings.Join([]string{
		"line with PATTERN",
		"another line",
		"line with PATTERN again",
		"line with PATTERN third time",
		"end line",
	}, "\n")
	assert.NoError(t, os.WriteFile(testFile, []byte(content), 0o644))

	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{
		"pattern": "PATTERN",
		"path":    tmpDir,
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)

	count := strings.Count(result.Output, "PATTERN")
	assert.Equal(t, 3, count)
}

func TestTool_Call_SpecialRegexPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := strings.Join([]string{
		"test@example.com",
		"user@domain.org",
		"not an email",
	}, "\n")
	assert.NoError(t, os.WriteFile(testFile, []byte(content), 0o644))

	tool := &Tool{}
	result, err := tool.Call(nil, map[string]any{
		"pattern": `\w+@\w+\.\w+`,
		"path":    tmpDir,
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "test@example.com")
	assert.Contains(t, result.Output, "user@domain.org")
	assert.NotContains(t, result.Output, "not an email")
}
