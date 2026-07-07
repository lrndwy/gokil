package views

import "net/http"

// Response is the standard JSON envelope for successful API responses.
type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func (c *Context) Success(status int, message string, data any) error {
	return c.JSON(status, Response{
		Status:  status,
		Message: message,
		Data:    data,
	})
}

func (c *Context) OK(message string, data any) error {
	return c.Success(http.StatusOK, message, data)
}

func (c *Context) Created(message string, data any) error {
	return c.Success(http.StatusCreated, message, data)
}
