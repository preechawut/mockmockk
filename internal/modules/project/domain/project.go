package domain

import (
	"mockapi/internal/platform/apperror"
	"regexp"
	"strings"
	"time"
)

type Project struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

var (
	ErrInvalid  = apperror.BadRequest("INVALID_PROJECT", "name and lowercase slug are required (slug: letters, digits, hyphens; max 63)")
	ErrNotFound = apperror.NotFound("PROJECT_NOT_FOUND", "project not found")
	ErrConflict = apperror.Conflict("PROJECT_CONFLICT", "project slug already exists")
	slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
)

func Validate(name, slug string) error {
	if strings.TrimSpace(name) == "" || len(name) > 200 || !slugPattern.MatchString(slug) {
		return ErrInvalid
	}
	return nil
}
