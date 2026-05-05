package log

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}

	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestNewPlainFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	logger.Info("hello", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "hello") {
		t.Errorf("expected output to contain 'hello', got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected output to contain 'key=value', got: %s", output)
	}
}

func TestNewJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	logger.Info("hello", "key", "value")

	output := buf.String()
	if !strings.Contains(output, `"msg":"hello"`) {
		t.Errorf("expected JSON output with msg field, got: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("expected JSON output with key field, got: %s", output)
	}
}

func TestNewWithOptions(t *testing.T) {
	logger := New(Options{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	})
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNewLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	logger.Info("should be filtered")
	if buf.Len() != 0 {
		t.Errorf("expected info message to be filtered at warn level, got: %s", buf.String())
	}

	logger.Warn("should appear")
	if !strings.Contains(buf.String(), "should appear") {
		t.Errorf("expected warn message to appear, got: %s", buf.String())
	}
}

func TestNewFileOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	logger := New(Options{
		Level:  "info",
		Format: "plain",
		Output: path,
	})

	logger.Info("file test", "num", 42)

	data, err := os.ReadFile(path) //nolint:gosec // test file path is safe
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if !strings.Contains(string(data), "file test") {
		t.Errorf("expected log file to contain 'file test', got: %s", string(data))
	}
}

func TestNewDefaults(t *testing.T) {
	logger := New(Options{})
	if logger == nil {
		t.Fatal("expected non-nil logger with empty options")
	}
}
