package logger

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

func New(out io.Writer, level, format string) (*slog.Logger, error) {
	if out == nil {
		return nil, fmt.Errorf("log output is required")
	}
	var severity slog.Level
	switch strings.ToLower(level) {
	case "", "info":
		severity = slog.LevelInfo
	case "debug":
		severity = slog.LevelDebug
	case "warn":
		severity = slog.LevelWarn
	case "error":
		severity = slog.LevelError
	default:
		return nil, fmt.Errorf("LOG_LEVEL must be debug, info, warn or error")
	}
	options := &slog.HandlerOptions{Level: severity}
	switch strings.ToLower(format) {
	case "", "json":
		return slog.New(slog.NewJSONHandler(out, options)), nil
	case "text":
		return slog.New(slog.NewTextHandler(out, options)), nil
	default:
		return nil, fmt.Errorf("LOG_FORMAT must be json or text")
	}
}
