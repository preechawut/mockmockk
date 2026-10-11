package apperror

import "errors"

type Kind string

const (
	Invalid      Kind = "invalid"
	NotFound     Kind = "not_found"
	Conflict     Kind = "conflict"
	Unauthorized Kind = "unauthorized"
	Forbidden    Kind = "forbidden"
	RateLimited  Kind = "rate_limited"
)

type Error struct {
	Kind       Kind
	Message    string
	RetryAfter int
}

func (e *Error) Error() string {
	return e.Message
}

func New(kind Kind, message string) error {
	return &Error{
		Kind:    kind,
		Message: message,
	}
}

func Is(err error, kind Kind) bool {
	var target *Error
	return errors.As(err, &target) && target.Kind == kind
}
