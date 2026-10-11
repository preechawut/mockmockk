package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

type contextKey struct{}

func New(output io.Writer, format, level string) (*slog.Logger, error) {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("invalid LOG_LEVEL: %w", err)
	}

	options := &slog.HandlerOptions{Level: logLevel}
	switch strings.ToLower(format) {
	case "json":
		return slog.New(slog.NewJSONHandler(output, options)), nil
	case "text":
		return slog.New(slog.NewTextHandler(output, options)), nil
	default:
		return nil, fmt.Errorf("LOG_FORMAT must be json or text")
	}
}

func WithContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(contextKey{}).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return log
}
