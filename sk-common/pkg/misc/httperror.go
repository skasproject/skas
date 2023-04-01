package misc

import (
	"fmt"
	"net/http"
)

type HttpError interface {
	error
	GetStatusCode() int
}

var _ HttpError = &httpError{}

type httpError struct {
	statusCode int
	message    string
}

func NewHttpError(message string, httpCode int) HttpError {
	return &httpError{
		statusCode: httpCode,
		message:    message,
	}
}

func (he *httpError) Error() string {
	if he.message == "" {
		return fmt.Sprintf("Http error: %s (%d)", http.StatusText(he.statusCode), he.statusCode)
	}
	return fmt.Sprintf("%s (%d:%s)", he.message, he.statusCode, http.StatusText(he.statusCode))
}

func (he *httpError) GetStatusCode() int {
	return he.statusCode
}
