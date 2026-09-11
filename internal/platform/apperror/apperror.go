package apperror

import (
	"fmt"
)

// error handling
type Error struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Err != nil && e.Message != "" {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" {
		return e.Code
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "application error"
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Status == t.Status && e.Code == t.Code
}

func New(status int, code, message string) *Error {
	return &Error{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, status int, code, message string) *Error {
	return &Error{
		Status:  status,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func BadRequest(code, message string) *Error {
	return New(400, code, message)
}

func NotFound(code, message string) *Error {
	return New(404, code, message)
}

func Conflict(code, message string) *Error {
	return New(409, code, message)
}

func Internal(message string, err error) *Error {
	return Wrap(err, 500, "INTERNAL_ERROR", message)
}
