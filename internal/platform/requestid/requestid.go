package requestid

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

const Header = "X-Request-ID"

type contextKey struct{}

func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(contextKey{}).(string)
	return id
}

func WithContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

func Valid(id string) bool {
	return len(id) > 0 && len(id) <= 64 && !strings.ContainsAny(id, " \t\r\n")
}

func Generate() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

func Ensure(id string) string {
	if !Valid(id) {
		return Generate()
	}
	return id
}
