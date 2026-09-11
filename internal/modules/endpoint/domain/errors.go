package domain

import (
	"mockapi/internal/platform/apperror"
)

var (
	ErrNotFound     = apperror.NotFound("NOT_FOUND", "endpoint not found")
	ErrNoMatch      = apperror.NotFound("NOT_FOUND", "no mock endpoint configured")
	ErrConflict     = apperror.Conflict("CONFLICT", "endpoint already exists")
	ErrInvalidInput = apperror.BadRequest("INVALID_INPUT", "invalid input")
)
