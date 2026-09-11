package logger

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFilteringAndFields(t *testing.T) {
	var out bytes.Buffer
	logger, err := New(&out, "warn", "json")
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("hidden")
	if out.Len() != 0 {
		t.Fatal("info was not filtered")
	}
	logger.With("service", "api").Error("failure", "code", "INTERNAL_ERROR")
	var entry map[string]any
	if err := json.Unmarshal(out.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry["service"] != "api" || entry["level"] != "ERROR" || entry["code"] != "INTERNAL_ERROR" {
		t.Fatalf("unexpected log: %s", out.String())
	}
}

func TestInvalidOptions(t *testing.T) {
	for _, pair := range [][2]string{{"nope", "json"}, {"info", "xml"}} {
		if _, err := New(&bytes.Buffer{}, pair[0], pair[1]); err == nil {
			t.Fatal("invalid option accepted")
		}
	}
}
