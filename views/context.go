package views

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/lrndwy/gokil/orm"
	"github.com/lrndwy/gokil/storage"
)

type Handler func(*Context) error

type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
	DB      *orm.DB
	Storage storage.Provider
	Params  map[string]string
}

func Wrap(handler Handler) func(http.ResponseWriter, *http.Request, map[string]string) {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		ctx := &Context{
			Request: r,
			Writer:  w,
			Params:  params,
		}
		if err := handler(ctx); err != nil {
			_ = Error(ctx, http.StatusInternalServerError, err.Error())
		}
	}
}

func (c *Context) JSON(status int, payload any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(payload)
}

func Error(c *Context, status int, message string) error {
	return c.JSON(status, map[string]any{"error": message})
}

func (c *Context) NoContent() {
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (c *Context) BindJSON(v any) error {
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

func (c *Context) DBContext() context.Context {
	return c.Request.Context()
}
