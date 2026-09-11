package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"

	controllerhttp "mockapi/internal/controller/http"
	"mockapi/internal/platform/apperror"
)

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestHTTPErrorHandler(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{fmt.Errorf("validation: %w", apperror.BadRequest("INVALID_INPUT", "invalid input")), 400, "INVALID_INPUT"},
		{fmt.Errorf("lookup: %w", apperror.NotFound("NOT_FOUND", "mockmock not found")), 404, "NOT_FOUND"},
		{apperror.Conflict("RESOURCE_EXISTS", "resource already exists"), 409, "RESOURCE_EXISTS"},
		{apperror.New(401, "UNAUTHORIZED", "authentication required"), 401, "UNAUTHORIZED"},
		{apperror.New(403, "FORBIDDEN", "access denied"), 403, "FORBIDDEN"},
		{context.DeadlineExceeded, 504, "REQUEST_TIMEOUT"},
		{fmt.Errorf("private operation: %w", apperror.Internal("internal error", errors.New("private sql"))), 500, "INTERNAL_ERROR"},
		{fmt.Errorf("private context: %w", apperror.Wrap(errors.New("private query"), 404, "RESOURCE_NOT_FOUND", "resource not found")), 404, "RESOURCE_NOT_FOUND"},

		{context.Canceled, 408, "REQUEST_CANCELED"},
		{fiber.ErrRequestEntityTooLarge, 413, "BODY_TOO_LARGE"},
		{errors.New("private database detail"), 500, "INTERNAL_ERROR"},
		{controllerhttp.ErrInvalidJSON, 400, "INVALID_JSON"},
		{controllerhttp.ErrUnsupportedMediaType, 415, "UNSUPPORTED_MEDIA_TYPE"},
		{controllerhttp.ErrMissingField, 400, "INVALID_INPUT"},
		{controllerhttp.ErrNotReady, 503, "NOT_READY"},
		{fiber.NewError(500, "private fiber detail"), 500, "INTERNAL_ERROR"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
				ErrorHandler:          controllerhttp.HTTPErrorHandler(logger),
			})
			app.Get("/", func(c *fiber.Ctx) error {
				return tc.err
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			bodyBytes, _ := io.ReadAll(resp.Body)
			var body errorResponse
			if err := json.Unmarshal(bodyBytes, &body); err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.status || body.Error.Code != tc.code || body.Error.Message == "" {
				t.Fatalf("unexpected error: %d %s", resp.StatusCode, string(bodyBytes))
			}
			if strings.Contains(string(bodyBytes), "private") {
				t.Fatal("internal detail leaked")
			}
			if tc.status == 500 && logs.Len() == 0 {
				t.Fatal("internal error was not logged")
			}
		})
	}
}

func TestInternalDetailsStayServerSide(t *testing.T) {
	dbError := &pgconn.PgError{
		Code: "23505", Message: "duplicate key violates private_constraint",
		Detail:    "Key (email)=(private@example.test) already exists",
		TableName: "private_users", ConstraintName: "private_constraint",
	}
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          controllerhttp.HTTPErrorHandler(logger),
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return fmt.Errorf("SELECT * FROM private_users: %w", dbError)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	const expected = `{"error":{"code":"INTERNAL_ERROR","message":"internal server error"}}`
	if resp.StatusCode != 500 || string(bodyBytes) != expected {
		t.Fatalf("unsafe response: %d %s", resp.StatusCode, string(bodyBytes))
	}
	if !strings.Contains(logs.String(), "private_constraint") {
		t.Fatal("missing server-side diagnostic")
	}
}
