package prouter

import (
	"slices"
	"strings"

	"github.com/gorilla/mux"
)

type Route interface {
	Handler() handlerFunc
	Path() string
	Method() string
}

type OptRoute interface {
	Route
	Option(*mux.Route) *mux.Route
}

// iRoute the final iRoute group which will use to register into mux.Router
type iRoute struct {
	Route
	router      *mux.Router
	middleware  []Middleware
	routeOption RouteOption
}

type RouteOption func(*mux.Route) *mux.Route

func (r *iRoute) handleSpecifyMiddleware(handler handlerFunc) handlerFunc {
	next := handler
	for _, m := range slices.Backward(r.middleware) {
		next = m.WrapHandler(next)
	}

	return next
}

type defaultRoute struct {
	method  string
	path    string
	handler handlerFunc

	// it use in HandlerRouter while route is OptRoute
	opts []RouteOption
}

func (r *defaultRoute) Handler() handlerFunc {
	return r.handler
}

func (r *defaultRoute) Path() string {
	return r.path
}

func (r *defaultRoute) Method() string {
	return r.method
}

func (r *defaultRoute) Option(route *mux.Route) *mux.Route {
	next := route
	for _, opt := range r.opts {
		next = opt(next)
	}

	return next
}

func newHandlerFuncRoute(method, path string, handler handlerFunc, opts ...RouteOption) Route {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return &defaultRoute{method, path, handler, opts}
}

func NewRoute(method, path string, handler HandleFunc, opts ...RouteOption) Route {
	return newHandlerFuncRoute(method, path, handler, opts...)
}
