package views

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/lrndwy/gokil/orm"
)

type Handler func(*Context) error

type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
	Params  map[string]string
}

type DBContextKey struct{}

func (c *Context) JSON(data any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(c.Writer).Encode(orm.ProjectForJSON(data))
}

func (c *Context) Success(status int, message string, data any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(map[string]any{
		"status":  status,
		"message": message,
		"data":    orm.ProjectForJSON(data),
	})
}

func (c *Context) Error(status int, message string) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(map[string]any{
		"error": message,
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

func (c *Context) DB() interface{} {
	return c.Request.Context().Value(DBContextKey{})
}
