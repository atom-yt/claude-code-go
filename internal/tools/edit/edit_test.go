package edit

import (
	"context"
	"os"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, _ := os.CreateTemp("", "edit_test_*.txt")
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestEdit_SingleReplace(t *testing.T) {
	path := writeTemp(t, "hello world\n")
	defer os.Remove(path)

	tool := &Tool{}
	result, _ := tool.Call(context.Background(), map[string]any{
		"file_path":  path,
		"old_string": "hello",
		"new_string": "goodbye",
	})
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "goodbye world") {
		t.Errorf("replacement failed: %s", string(data))
	}
}

func TestEdit_DuplicateError(t *testing.T) {
	path := writeTemp(t, "foo\nfoo\n")
	defer os.Remove(path)

	tool := &Tool{}
	result, _ := tool.Call(context.Background(), map[string]any{
		"file_path":  path,
		"old_string": "foo",
		"new_string": "bar",
	})
	if !result.IsError {
		t.Fatal("expected error for duplicate match")
	}
}

func TestEdit_ReplaceAll(t *testing.T) {
	path := writeTemp(t, "a\na\na\n")
	defer os.Remove(path)

	tool := &Tool{}
	result, _ := tool.Call(context.Background(), map[string]any{
		"file_path":   path,
		"old_string":  "a",
		"new_string":  "b",
		"replace_all": true,
	})
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "a") {
		t.Errorf("not all replaced: %s", string(data))
	}
}
