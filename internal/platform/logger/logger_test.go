package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestLoggerFormatsLevelAndContext(t *testing.T) {
	for _, format := range []string{"json", "text"} {
		var output bytes.Buffer
		log, err := New(&output, format, "WARN")
		if err != nil {
			t.Fatal(err)
		}
		log.Info("hidden")
		ctx := WithContext(context.Background(), log.With("request_id", "test-id"))
		FromContext(ctx).Warn("visible")
		if strings.Contains(output.String(), "hidden") || !strings.Contains(output.String(), "test-id") {
			t.Fatal("level/context not applied")
		}
		if format == "json" {
			var entry map[string]any
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := New(&bytes.Buffer{}, "xml", "INFO"); err == nil {
		t.Fatal("accepted invalid format")
	}
	if _, err := New(&bytes.Buffer{}, "json", "invalid"); err == nil {
		t.Fatal("accepted invalid level")
	}
}
