package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	MaxBodySize = 1 << 20
)

func DecodeJSON(c *fiber.Ctx, target any) error {
	if !c.Is("json") {
		return ErrUnsupportedMediaType
	}
	dec := json.NewDecoder(bytes.NewReader(c.Body()))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return nil
}

func PathUUID(c *fiber.Ctx, name string) (string, error) {
	val := strings.TrimSpace(c.Params(name))
	if val == "" {
		return "", ErrInvalidID
	}
	if _, err := uuid.Parse(val); err != nil {
		return "", ErrInvalidID
	}
	return strings.Clone(val), nil
}
