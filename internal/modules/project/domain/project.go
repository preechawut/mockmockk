package domain

import (
	"regexp"
	"strings"
	"time"

	"mockapi/internal/shared/apperror"
)

type Project struct {
	ID        string
	Slug      string
	Name      string
	CreatedAt time.Time
}

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

func (p *Project) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	p.Slug = strings.TrimSpace(p.Slug)

	if p.Name == "" || len(p.Name) > 200 {
		return apperror.New(apperror.Invalid, "name is required and must not exceed 200 characters")
	}

	if !slugPattern.MatchString(p.Slug) {
		return apperror.New(apperror.Invalid, "slug is required (letters, digits, hyphens; max 63 characters)")
	}

	return nil
}

func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 200 {
		return apperror.New(apperror.Invalid, "name is required and must not exceed 200 characters")
	}
	return nil
}
