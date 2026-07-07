package views

import (
	"errors"
	"net/http"
)

// HTTPError is returned by handlers to send a specific HTTP status to the client.
type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func BadRequest(message string) error {
	return &HTTPError{Status: http.StatusBadRequest, Message: message}
}

func NotFound(message string) error {
	return &HTTPError{Status: http.StatusNotFound, Message: message}
}

func Conflict(message string) error {
	return &HTTPError{Status: http.StatusConflict, Message: message}
}

// HandleError writes a JSON error response. HTTPError uses its status code; other errors become 500.
func HandleError(c *Context, err error) error {
	if err == nil {
		return nil
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return Error(c, httpErr.Status, httpErr.Message)
	}
	return Error(c, http.StatusInternalServerError, err.Error())
}
