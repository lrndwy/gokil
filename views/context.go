package views

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/lrndwy/gokil/orm"
)

type Handler func(*Context) error

type Middleware func(Handler) Handler

type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
	Params  map[string]string
}

func (c *Context) JSON(data any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(c.Writer).Encode(orm.ProjectForJSON(data))
}

func (c *Context) Success(status int, message string, data any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(map[string]any{
		"success": true,
		"message": message,
		"data":    orm.ProjectForJSON(data),
	})
}

func (c *Context) Error(status int, message string) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(map[string]any{
		"success": false,
		"message": message,
	})
}

func (c *Context) ValidationError(status int, message string, errors map[string][]string) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(map[string]any{
		"success": false,
		"message": message,
		"errors":  errors,
	})
}

func (c *Context) NotFound() error {
	return c.Error(http.StatusNotFound, "not found")
}

func (c *Context) NoContent() {
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (c *Context) Bind(v any) error {
	defer c.Request.Body.Close()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func (c *Context) Param(name string) string {
	if c.Params == nil {
		return ""
	}
	return c.Params[name]
}

func (c *Context) Query(name string) string {
	return c.Request.URL.Query().Get(name)
}

func (c *Context) DB() *orm.DB {
	if c.Request == nil {
		return nil
	}
	return orm.DBFromContext(c.Request.Context())
}

func (c *Context) ParseMultipart(maxMemory int64) error {
	return c.Request.ParseMultipartForm(maxMemory)
}

func (c *Context) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return c.Request.FormFile(name)
}
