package views

import "net/http"

// Response is the standard JSON envelope for successful API responses.
type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// PaginatedResponse is the standard JSON envelope for paginated list responses.
type PaginatedResponse struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Data    any      `json:"data"`
	Meta    PageMeta `json:"meta"`
}

// PageMeta holds pagination metadata.
type PageMeta struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Pages int   `json:"pages"`
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

func (c *Context) Paginated(message string, data any, meta PageMeta) error {
	return c.JSON(http.StatusOK, PaginatedResponse{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// ResourceCreated writes a 201 response using "<resource> created" message format.
func (c *Context) ResourceCreated(resource string, data any) error {
	return c.Created(resource+" created", data)
}

// ResourceOK writes a 200 response using "<resource> <action>" message format.
func (c *Context) ResourceOK(action, resource string, data any) error {
	return c.OK(resource+" "+action, data)
}

// NoContentResponse writes 204 No Content and returns nil for the handler.
func (c *Context) NoContentResponse() error {
	c.NoContent()
	return nil
}
