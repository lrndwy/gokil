package router

import (
	"net/http"
	"sort"
	"strings"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request, params map[string]string)

type Route struct {
	Method  string
	Path    string
	Handler HandlerFunc
}

type Router struct {
	routes     []Route
	middleware []func(http.Handler) http.Handler
}

func New() *Router {
	return &Router{}
}

func (r *Router) Use(mw func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, mw)
}

func (r *Router) Handle(method, path string, handler HandlerFunc) {
	r.routes = append(r.routes, Route{
		Method:  strings.ToUpper(strings.TrimSpace(method)),
		Path:    normalizePath(path),
		Handler: handler,
	})
}

func (r *Router) GET(path string, handler HandlerFunc)    { r.Handle(http.MethodGet, path, handler) }
func (r *Router) POST(path string, handler HandlerFunc)   { r.Handle(http.MethodPost, path, handler) }
func (r *Router) PUT(path string, handler HandlerFunc)    { r.Handle(http.MethodPut, path, handler) }
func (r *Router) PATCH(path string, handler HandlerFunc)  { r.Handle(http.MethodPatch, path, handler) }
func (r *Router) DELETE(path string, handler HandlerFunc) { r.Handle(http.MethodDelete, path, handler) }

func (r *Router) Mount(routes ...Route) {
	for _, route := range routes {
		r.Handle(route.Method, route.Path, route.Handler)
	}
}

func (r *Router) Handler() http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := normalizePath(req.URL.Path)
		for _, route := range r.routes {
			if route.Method != req.Method {
				continue
			}
			if params, ok := matchPath(route.Path, path); ok {
				route.Handler(w, req, params)
				return
			}
		}
		writeNotFound(w)
	})

	var h http.Handler = handler
	for i := len(r.middleware) - 1; i >= 0; i-- {
		h = r.middleware[i](h)
	}
	return h
}

// RouteInfo is a lightweight representation of a registered route.
type RouteInfo struct {
	Method string
	Path   string
}

// Routes returns all registered routes.
func (r *Router) Routes() []RouteInfo {
	routes := make([]RouteInfo, 0, len(r.routes))
	for _, route := range r.routes {
		routes = append(routes, RouteInfo{Method: route.Method, Path: route.Path})
	}
	return routes
}

func (r *Router) Paths() []string {
	paths := make([]string, 0, len(r.routes))
	for _, route := range r.routes {
		paths = append(paths, route.Method+" "+route.Path)
	}
	sort.Strings(paths)
	return paths
}

func matchPath(pattern, path string) (map[string]string, bool) {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if pattern == "/" && path == "/" {
		return map[string]string{}, true
	}
	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := map[string]string{}
	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			key := strings.TrimSuffix(strings.TrimPrefix(part, ":"), "/")
			params[key] = pathParts[i]
			continue
		}
		if part != pathParts[i] {
			return nil, false
		}
	}
	return params, true
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func writeNotFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"error":"not found"}`))
}
