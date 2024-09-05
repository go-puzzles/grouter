package prouter

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/go-puzzles/plog"
	"github.com/gorilla/mux"
)

type RouterGroup struct {
	// prouter is the root instance
	prouter *Prouter
	// router is which used by the current group
	router      *mux.Router
	routes      []iRoute
	middlewares []Middleware
	root        bool
}

func newGroupWithRouter(router *mux.Router) RouterGroup {
	return RouterGroup{
		router:      router,
		routes:      make([]iRoute, 0),
		middlewares: make([]Middleware, 0),
	}
}

func (rg *RouterGroup) Use(middlewares ...HandleFunc) {
	ms := make([]Middleware, 0, len(middlewares))
	for _, m := range middlewares {
		ms = append(ms, m)
	}
	rg.middlewares = append(rg.middlewares, ms...)
}

func (rg *RouterGroup) UseMiddleware(m ...Middleware) {
	rg.middlewares = append(rg.middlewares, m...)
}

func (rg *RouterGroup) HandleRouter(routers ...Router) {
	wrapRoutes := func(routes []Route) {
		for _, r := range routes {
			var opt RouteOption
			switch tr := r.(type) {
			case OptRoute:
				opt = tr.Option
			default:
			}

			rg.initRouter(iRoute{
				Route:       r,
				router:      rg.router,
				middleware:  rg.middlewares,
				routeOption: opt,
			})
		}
	}

	for _, router := range routers {
		wrapRoutes(router.Routes())
	}
}

func (rg *RouterGroup) handleRoute(method, path string, handler handlerFunc, opts ...RouteOption) {
	routeOpt := func(r *mux.Route) *mux.Route {

		if opts == nil {
			return r
		}

		next := r
		for _, opt := range opts {
			next = opt(next)
		}

		return next
	}

	r := iRoute{
		Route:       newHandlerFuncRoute(method, path, handler),
		router:      rg.router,
		routeOption: routeOpt,
	}
	r.middleware = rg.middlewares

	rg.initRouter(r)

}

func (rg *RouterGroup) HandleRoute(method, path string, handler HandleFunc, opts ...RouteOption) {
	rg.handleRoute(method, path, handler, opts...)
}

func (rg *RouterGroup) initRouter(r iRoute) {
	f := rg.prouter.makeHttpHandler(r)

	vr := r.router.PathPrefix(r.Path())
	if r.Method() != "" {
		vr = vr.Methods(r.Method())
	}

	if r.routeOption != nil {
		vr = r.routeOption(vr)
	}

	mr := vr.Handler(f)
	rg.debugPrintRoute(r.Method(), mr, r.Handler())
}

func (rg *RouterGroup) debugPrintRoute(method string, route *mux.Route, handler handlerFunc) {
	if prouterMode != DebugMode {
		return
	}
	if method == "" {
		method = "ANY"
	}

	handlerName := handler.Name()
	url, err := route.GetPathTemplate()
	if err != nil {
		plog.Errorf("get handler: %v iRoute url error: %v", handlerName, err)
	}

	plog.Infof("Method: %-6s Router: %-30s Handler: %s", method, url, handlerName)
}

func (rg *RouterGroup) Group(prefix string, middlewares ...HandleFunc) *RouterGroup {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	router := rg.router.PathPrefix(prefix).Subrouter()
	g := newGroupWithRouter(router)
	g.middlewares = append(g.middlewares, rg.middlewares...)
	g.prouter = rg.prouter

	g.Use(middlewares...)

	return &g
}

func (rg *RouterGroup) staticHandler(prefix string, fs http.FileSystem) HandleFunc {
	return func(ctx *Context) (Response, error) {
		r := ctx.Request
		w := ctx.Writer

		p := strings.TrimPrefix(r.URL.Path, prefix)
		rp := strings.TrimPrefix(r.URL.RawPath, prefix)

		if len(p) < len(r.URL.Path) && (r.URL.RawPath == "" || len(rp) < len(r.URL.RawPath)) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			r2.URL.RawPath = rp

			http.FileServer(fs).ServeHTTP(w, r2)
		} else {
			return nil, MsgError(http.StatusNotFound, fmt.Sprintf("%v static file not found", p))
		}
		return nil, nil
	}
}

func (rg *RouterGroup) Static(path, root string, opts ...RouteOption) {
	rg.StaticFS(path, http.Dir(root), opts...)
}

func (rg *RouterGroup) StaticFsEmbed(path, fileRelativePath string, fsEmbed embed.FS, opts ...RouteOption) {
	subFs, err := fs.Sub(fsEmbed, fileRelativePath)
	if err != nil {
		panic(err)
	}

	rg.StaticFS(path, http.FS(subFs))
}

func (rg *RouterGroup) StaticFS(relativePath string, fs http.FileSystem, opts ...RouteOption) {
	if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}

	urlPattern := path.Join(relativePath, "{filepath}")
	handler := &wrapHandler{
		name:    "StaticFSHandler",
		handler: rg.staticHandler(relativePath, fs),
	}

	rg.handleRoute(http.MethodGet, urlPattern, handler, opts...)
}

func (rg *RouterGroup) GET(path string, handler HandleFunc, opt ...RouteOption) {
	rg.HandleRoute(http.MethodGet, path, handler, opt...)
}

func (rg *RouterGroup) POST(path string, handler HandleFunc, opt ...RouteOption) {
	rg.HandleRoute(http.MethodPost, path, handler, opt...)
}

func (rg *RouterGroup) PUT(path string, handler HandleFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodPut, path, handler, opts...)
}

func (rg *RouterGroup) PATCH(path string, handler HandleFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodPatch, path, handler, opts...)
}

func (rg *RouterGroup) DELETE(path string, handler HandleFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodDelete, path, handler, opts...)
}

func (rg *RouterGroup) OPTIONS(path string, handler HandleFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodOptions, path, handler, opts...)
}

func (rg *RouterGroup) HEAD(path string, handler HandleFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodHead, path, handler, opts...)
}

func (rg *RouterGroup) TRACE(path string, handler HandleFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodTrace, path, handler, opts...)
}

func (rg *RouterGroup) Any(path string, handler HandleFunc, opts ...RouteOption) {
	rg.HandleRoute("", path, handler, opts...)
}
