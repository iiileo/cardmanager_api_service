package errors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrValidation = errors.New("validation_error")
	ErrNotFound   = errors.New("not_found")
	ErrInternal   = errors.New("internal_error")
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

type AppError struct {
	err    error
	msg    string
	hint   string
	marked error
}

func NewError(msg string) *AppError {
	return &AppError{msg: msg, err: errors.New(msg)}
}

func WithError(err error) *AppError {
	return &AppError{err: err, msg: err.Error()}
}

func (e *AppError) WithHint(hint string) *AppError {
	e.hint = hint
	return e
}

func (e *AppError) Mark(mark error) *AppError {
	e.marked = mark
	return e
}

func (e *AppError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	if e.err != nil {
		return e.err.Error()
	}
	return "unknown error"
}

func (e *AppError) Unwrap() error {
	return e.err
}

func (e *AppError) Is(target error) bool {
	if e.marked != nil && errors.Is(e.marked, target) {
		return true
	}
	return errors.Is(e.err, target)
}

func (e *AppError) Hint() string {
	return e.hint
}

func (e *AppError) HTTPStatus() int {
	switch {
	case e.Is(ErrValidation):
		return http.StatusBadRequest
	case e.Is(ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func (e *AppError) ToResponse() ErrorResponse {
	code := "internal_error"
	switch {
	case e.Is(ErrValidation):
		code = ErrValidation.Error()
	case e.Is(ErrNotFound):
		code = ErrNotFound.Error()
	case e.Is(ErrInternal):
		code = ErrInternal.Error()
	}
	return ErrorResponse{
		Error:   code,
		Message: e.Error(),
		Hint:    e.hint,
	}
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}
