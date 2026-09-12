package errors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrValidation   = errors.New("validation_error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not_found")
	ErrConflict     = errors.New("conflict")
	ErrInternal     = errors.New("internal_error")
)

type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type AppError struct {
	err    error
	msg    string
	hint   string
	marked error
	code   int
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

func (e *AppError) WithCode(code int) *AppError {
	e.code = code
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

func (e *AppError) Unwrap() error { return e.err }

func (e *AppError) Is(target error) bool {
	if e.marked != nil && errors.Is(e.marked, target) {
		return true
	}
	return errors.Is(e.err, target)
}

func (e *AppError) Hint() string { return e.hint }

func (e *AppError) Code() int {
	if e.code != 0 {
		return e.code
	}
	switch {
	case e.Is(ErrValidation):
		return 40000
	case e.Is(ErrUnauthorized):
		return 40100
	case e.Is(ErrForbidden):
		return 40300
	case e.Is(ErrNotFound):
		return 40400
	case e.Is(ErrConflict):
		return 40900
	default:
		return 50000
	}
}

func (e *AppError) HTTPStatus() int {
	switch {
	case e.Is(ErrValidation):
		return http.StatusBadRequest
	case e.Is(ErrUnauthorized):
		return http.StatusUnauthorized
	case e.Is(ErrForbidden):
		return http.StatusForbidden
	case e.Is(ErrNotFound):
		return http.StatusNotFound
	case e.Is(ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func (e *AppError) ToEnvelope() Envelope {
	data := any(nil)
	if e.hint != "" {
		data = map[string]string{"hint": e.hint}
	}
	return Envelope{
		Code:    e.Code(),
		Message: e.Error(),
		Data:    data,
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
