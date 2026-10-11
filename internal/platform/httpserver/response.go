package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"mockapi/internal/platform/logger"
	"mockapi/internal/shared/apperror"
)

type successResponse struct {
	OK   bool `json:"ok"`
	Data any  `json:"data"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	OK    bool        `json:"ok"`
	Error errorDetail `json:"error"`
}

func JSON(c fiber.Ctx, status int, value any) error {
	return c.Status(status).JSON(successResponse{
		OK:   true,
		Data: value,
	})
}

func ErrorStatus(c fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(errorResponse{
		Error: errorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func Error(c fiber.Ctx, err error) error {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(c.Context().Err(), context.DeadlineExceeded) {
		return ErrorStatus(c, fiber.StatusServiceUnavailable, "request_timeout", "request timed out")
	}

	status := fiber.StatusInternalServerError
	code := "internal_error"
	message := "an unexpected error occurred"

	var known *apperror.Error
	var framework *fiber.Error

	switch {
	case errors.As(err, &known):
		code = string(known.Kind)
		message = known.Message

		switch known.Kind {
		case apperror.Invalid:
			status = fiber.StatusBadRequest
		case apperror.NotFound:
			status = fiber.StatusNotFound
		case apperror.Conflict:
			status = fiber.StatusConflict
		case apperror.Unauthorized:
			status = fiber.StatusUnauthorized
		case apperror.Forbidden:
			status = fiber.StatusForbidden
		case apperror.RateLimited:
			status = fiber.StatusTooManyRequests
			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(max(1, known.RetryAfter)))
		}

	case errors.As(err, &framework) && framework.Code < fiber.StatusInternalServerError:
		status = framework.Code

		switch status {
		case fiber.StatusRequestEntityTooLarge:
			code, message = "body_too_large", "body must not exceed 64 KiB"
		case fiber.StatusNotFound:
			code, message = "not_found", "route not found"
		case fiber.StatusMethodNotAllowed:
			code, message = "method_not_allowed", "method not allowed"
		case fiber.StatusRequestTimeout:
			status, code, message = fiber.StatusServiceUnavailable, "request_timeout", "request timed out"
		default:
			code, message = "invalid", framework.Message
		}

	default:
		logger.FromContext(c.Context()).Error("request failed", "error", err)
	}

	return ErrorStatus(c, status, code, message)
}

func Decode(c fiber.Ctx, target any) error {
	media, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || media != fiber.MIMEApplicationJSON {
		return apperror.New(apperror.Invalid, "Content-Type must be application/json")
	}

	body := c.Body()
	if len(body) > 64*1024 {
		return apperror.New(apperror.Invalid, "body must not exceed 64 KiB")
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(target)
	if err != nil {
		return apperror.New(apperror.Invalid, "invalid JSON body")
	}

	if decoder.Decode(new(any)) != io.EOF {
		return apperror.New(apperror.Invalid, "body must contain one JSON value")
	}

	return nil
}
