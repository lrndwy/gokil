package framework

import (
	"github.com/lrndwy/gokil/views"
)

type pendingRoute struct {
	Method  string
	Path    string
	Handler views.Handler
}

var pendingRoutes []pendingRoute

func RegisterRoute(method, path string, handler views.Handler) {
	pendingRoutes = append(pendingRoutes, pendingRoute{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
}

func (app *App) setupRoutes() {
	for _, route := range pendingRoutes {
		app.Router.Handle(route.Method, route.Path, app.Wrap(route.Handler))
	}
}
