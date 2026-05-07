// Package log provides a structured logging wrapper built on Go's stdlib log/slog.
package log

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Options configures the logger.
type Options struct {
	// Level is the minimum log level: debug, info, warn, error. Default: info.
	Level string
	// Format is the output format: json or plain. Default: plain.
	Format string
	// Output is the destination: stdout, stderr, or a file path. Default: stderr.
	Output string
}

// ParseLevel converts a level string to slog.Level.
// Unrecognized values default to slog.LevelInfo.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// New creates a configured *slog.Logger from the given Options.
// If Output is a file path, the file is opened for append (created if needed).
// The caller is responsible for closing any file-based output if needed.
func New(opts Options) *slog.Logger {
	w := resolveOutput(opts.Output)
	level := ParseLevel(opts.Level)

	handlerOpts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(opts.Format)) {
	case "json":
		handler = slog.NewJSONHandler(w, handlerOpts)
	default:
		handler = slog.NewTextHandler(w, handlerOpts)
	}

	return slog.New(handler)
}

func resolveOutput(output string) io.Writer {
	switch strings.ToLower(strings.TrimSpace(output)) {
	case "", "stderr":
		return os.Stderr
	case "stdout":
		return os.Stdout
	default:
		f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) //nolint:gosec // path is intentionally caller-provided
		if err != nil {
			return os.Stderr
		}
		return f
	}
}
