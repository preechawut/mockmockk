package requestid

import (
	"context"
	"strings"
	"testing"
)

func TestEnsureAndContext(t *testing.T) {
	for _, id := range []string{"", "valid-123", strings.Repeat("x", 65), "unsafe\nvalue"} {
		t.Run(id[:min(len(id), 10)], func(t *testing.T) {
			captured := Ensure(id)
			ctx := WithContext(context.Background(), captured)
			if !Valid(captured) || captured != FromContext(ctx) {
				t.Fatal("invalid correlation ID")
			}

			if Valid(id) && captured != id {
				t.Fatal("valid ID was replaced")
			}
			if !Valid(id) && len(captured) != 32 {
				t.Fatal("expected generated ID")
			}
		})
	}
}
