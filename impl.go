package prouter

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	statusCodeKey = "prouter-status-key"
)

type ResponseTmpl interface {
	SetCode(int)
	SetData(any)
	SetMessage(string)
}

// iRoute the final iRoute group which will use to register into mux.Router
type iRoute struct {
	Route
	router      *mux.Router
	middleware  []Middleware
	routeOption RouteOption
}

type RouteOption func(*mux.Route) *mux.Route

func (r *iRoute) handleSpecifyMiddleware(handler HandleFunc) HandleFunc {
	next := handler
	for _, m := range r.middleware {
		next = m.WrapHandler(next)
	}

	return next
}

type Response interface {
	GetCode() int
	GetMessage() string
	GetData() any
}

type HandleFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) (Response, error)

type Router interface {
	Routes() []Route
}

type Route interface {
	Handler() HandleFunc
	Path() string
	Method() string
}

type OptRoute interface {
	Route
	Option(*mux.Route) *mux.Route
}

type Middleware interface {
	WrapHandler(handler HandleFunc) HandleFunc
}

func (f HandleFunc) WrapHandler(handler HandleFunc) HandleFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) (Response, error) {
		resp, err := f(ctx, w, r, vars)
		if err != nil {
			return resp, err
		}

		return handler(ctx, w, r, vars)
	}
}

type defaultRoute struct {
	method  string
	path    string
	handler HandleFunc
	opts    []RouteOption
}

func (r *defaultRoute) Handler() HandleFunc {
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

func NewRoute(method, path string, handler HandleFunc, opts ...RouteOption) Route {
	return &defaultRoute{method, path, handler, opts}
}

func NewGetRoute(path string, handler HandleFunc, opts ...RouteOption) Route {
	return NewRoute(http.MethodGet, path, handler, opts...)
}

func NewPostRoute(path string, handler HandleFunc, opts ...RouteOption) Route {
	return NewRoute(http.MethodPost, path, handler, opts...)
}

func NewPutRoute(path string, handler HandleFunc, opts ...RouteOption) Route {
	return NewRoute(http.MethodPut, path, handler, opts...)
}

func NewDeleteRoute(path string, handler HandleFunc, opts ...RouteOption) Route {
	return NewRoute(http.MethodDelete, path, handler, opts...)
}

func NewOptionsRoute(path string, handler HandleFunc, opts ...RouteOption) Route {
	return NewRoute(http.MethodOptions, path, handler, opts...)
}

func NewHeadRoute(path string, handler HandleFunc, opts ...RouteOption) Route {
	return NewRoute(http.MethodHead, path, handler, opts...)
}

func NewPatchRoute(path string, handler HandleFunc, opts ...RouteOption) Route {
	return NewRoute(http.MethodPatch, path, handler, opts...)
}

func NewAnyRoute(path string, handler HandleFunc, opts ...RouteOption) Route {
	return NewRoute("", path, handler, opts...)
}
