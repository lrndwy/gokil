package framework

import (
	"github.com/lrndwy/gokil/views"
)

type pendingRoute struct {
	Method     string
	Path       string
	Handler    views.Handler
	Middleware []views.Middleware
}

var pendingRoutes []pendingRoute

func RegisterRoute(method, path string, handler views.Handler, mws ...views.Middleware) {
	pendingRoutes = append(pendingRoutes, pendingRoute{
		Method:     method,
		Path:       path,
		Handler:    handler,
		Middleware: mws,
	})
}

func (app *App) setupRoutes() {
	for _, route := range pendingRoutes {
		h := route.Handler
		for i := len(route.Middleware) - 1; i >= 0; i-- {
			h = route.Middleware[i](h)
		}
		app.Router.Handle(route.Method, route.Path, app.Wrap(h))
	}
}
