package http

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	"mockapi/internal/controller/http/middleware"
	"mockapi/internal/platform/apperror"
	"mockapi/internal/platform/requestid"
)

var (
	ErrNotReady             = errors.New("service not ready")
	ErrUnsupportedMediaType = errors.New("unsupported media type")
	ErrBodyTooLarge         = errors.New("body too large")
	ErrInvalidJSON          = errors.New("invalid JSON")
	ErrInvalidID            = errors.New("invalid ID")
	ErrMissingField         = errors.New("missing required field")
)

// error handling
func HTTPErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		code := "INTERNAL_ERROR"
		message := "internal server error"
		var appErr *apperror.Error
		var fiberErr *fiber.Error

		switch {
		case errors.Is(err, context.Canceled):
			status, code, message = fiber.StatusRequestTimeout, "REQUEST_CANCELED", "request canceled"
		case errors.Is(err, context.DeadlineExceeded):
			status, code, message = fiber.StatusGatewayTimeout, "REQUEST_TIMEOUT", "request timed out"
			logger.Warn("request timed out", "code", code, "request_id", requestid.FromContext(c.UserContext()), "error", err)
		case errors.As(err, &appErr):
			status = appErr.Status
			code = appErr.Code
			message = appErr.Message
			if status >= 500 {
				logger.Error("request failed", "code", code, "request_id", requestid.FromContext(c.UserContext()), "error", appErr.Err)
			}
		case errors.Is(err, ErrUnsupportedMediaType):
			status, code, message = fiber.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json"
		case errors.Is(err, ErrBodyTooLarge), errors.Is(err, fiber.ErrRequestEntityTooLarge):
			status, code, message = fiber.StatusRequestEntityTooLarge, "BODY_TOO_LARGE", "body exceeds 1 MiB"
		case errors.Is(err, ErrInvalidJSON):
			status, code = fiber.StatusBadRequest, "INVALID_JSON"
			message = detailedMessage(err, ErrInvalidJSON, "body must be one valid JSON object with supported fields")
		case errors.Is(err, ErrInvalidID):
			status, code, message = fiber.StatusBadRequest, "INVALID_INPUT", "invalid id"
		case errors.Is(err, ErrMissingField):
			status, code, message = fiber.StatusBadRequest, "INVALID_INPUT", "missing required field"
		case errors.Is(err, ErrNotReady):
			status, code, message = fiber.StatusServiceUnavailable, "NOT_READY", "service unavailable"
		case errors.Is(err, middleware.ErrPanic):
		case errors.As(err, &fiberErr) && fiberErr.Code == fiber.StatusNotFound:
			status, code, message = fiber.StatusNotFound, "NOT_FOUND", "route not found"
		case errors.As(err, &fiberErr) && fiberErr.Code == fiber.StatusMethodNotAllowed:
			status, code, message = fiber.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed"
		case errors.As(err, &fiberErr) && fiberErr.Code < fiber.StatusInternalServerError:
			status, code, message = fiberErr.Code, "ERROR", fiberErr.Message
		default:
			logger.Error("request failed", "code", code, "request_id", requestid.FromContext(c.UserContext()), "error", err)
		}

		return c.Status(status).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    code,
				"message": message,
			},
		})
	}
}

func detailedMessage(err, target error, fallback string) string {
	if err == nil {
		return fallback
	}
	prefix := target.Error() + ": "
	msg := err.Error()
	if strings.HasPrefix(msg, prefix) {
		detail := strings.TrimPrefix(msg, prefix)
		if detail != "" {
			return detail
		}
	}
	return fallback
}
