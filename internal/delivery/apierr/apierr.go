package apierr

import (
	"betera-tz/pkg/errs"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("already exists")
	ErrBadRequest     = errors.New("bad request")
	ErrInternalServer = errors.New("internal server error")
	ErrRequestTimeout = errors.New("request timeout")
)

type ApiErr struct {
	Code    int
	Message any
}

func (ae ApiErr) Error() string {
	return fmt.Sprintf("error: %s, code: %d", ae.Message, ae.Code)
}

func NewApiError(code int, err error) ApiErr {
	return ApiErr{
		Code:    code,
		Message: err.Error(),
	}
}

func ToApiError(err error) ApiErr {
	switch {
	case errors.Is(err, errs.ErrNotFoundBase):
		return NotFound()
	case errors.Is(err, errs.ErrAlreadyExistsBase):
		return AlreadyExists()
	case errors.Is(err, errs.ErrInvalidValuesBase):
		return InvalidValues()
	default:
		return InternalServerError()
	}
}

func InternalServerError() ApiErr {
	return NewApiError(http.StatusInternalServerError, ErrInternalServer)
}

func InvalidRequest() ApiErr {
	return NewApiError(http.StatusBadRequest, ErrBadRequest)
}

func NotFound() ApiErr {
	return NewApiError(http.StatusNotFound, ErrNotFound)
}

func AlreadyExists() ApiErr {
	return NewApiError(http.StatusConflict, ErrAlreadyExists)
}

func RequestTimeout() ApiErr {
	return NewApiError(http.StatusRequestTimeout, ErrRequestTimeout)
}

func InvalidValues() ApiErr {
	return NewApiError(http.StatusBadRequest, errs.ErrInvalidValuesBase)
}
