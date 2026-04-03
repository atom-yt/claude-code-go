package read

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestRead_Basic(t *testing.T) {
	f, _ := os.CreateTemp("", "read_test_*.txt")
	f.WriteString("line1\nline2\nline3\n")
	f.Close()
	defer os.Remove(f.Name())

	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{"file_path": f.Name()})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if !strings.Contains(result.Output, "1\tline1") {
		t.Errorf("expected line numbers, got: %s", result.Output)
	}
}

func TestRead_OffsetLimit(t *testing.T) {
	f, _ := os.CreateTemp("", "read_test_*.txt")
	for i := 1; i <= 10; i++ {
		f.WriteString("line\n")
	}
	f.Close()
	defer os.Remove(f.Name())

	tool := &Tool{}
	result, _ := tool.Call(context.Background(), map[string]any{
		"file_path": f.Name(),
		"offset":    float64(3),
		"limit":     float64(2),
	})
	lines := strings.Split(strings.TrimSpace(result.Output), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %q", len(lines), result.Output)
	}
	if !strings.HasPrefix(lines[0], "3\t") {
		t.Errorf("expected line 3, got: %s", lines[0])
	}
}

func TestRead_Truncation(t *testing.T) {
	f, _ := os.CreateTemp("", "read_test_*.txt")
	for i := 0; i < defaultMaxLines+10; i++ {
		f.WriteString("x\n")
	}
	f.Close()
	defer os.Remove(f.Name())

	tool := &Tool{}
	result, _ := tool.Call(context.Background(), map[string]any{"file_path": f.Name()})
	if !strings.Contains(result.Output, "truncated") {
		t.Error("expected truncation notice")
	}
}
