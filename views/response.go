package views

import "net/http"

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type PaginatedResponse struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Data    any      `json:"data"`
	Meta    PageMeta `json:"meta"`
}

type PageMeta struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Pages int   `json:"pages"`
}

func (c *Context) OK(message string, data any) error {
	return c.Success(http.StatusOK, message, data)
}

func (c *Context) Created(message string, data any) error {
	return c.Success(http.StatusCreated, message, data)
}

func (c *Context) Paginated(message string, data any, meta PageMeta) error {
	return c.JSON(PaginatedResponse{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func (c *Context) ResourceCreated(resource string, data any) error {
	return c.Created(resource+" created", data)
}

func (c *Context) ResourceOK(action, resource string, data any) error {
	return c.OK(resource+" "+action, data)
}

func (c *Context) NoContentResponse() error {
	c.NoContent()
	return nil
}
