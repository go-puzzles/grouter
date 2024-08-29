package prouter

import (
	"embed"
	"errors"
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
	paths       []string
}

func newGroupWithRouter(router *mux.Router) RouterGroup {
	return RouterGroup{
		router:      router,
		routes:      make([]iRoute, 0),
		middlewares: make([]Middleware, 0),
		paths:       make([]string, 0),
	}
}

func (rg *RouterGroup) UseMiddleware(m ...Middleware) {
	rg.middlewares = append(rg.middlewares, m...)
}

func (rg *RouterGroup) HandlerRouter(routers ...Router) {
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

func (rg *RouterGroup) HandleRoute(method, path string, handler handlerFunc, opts ...RouteOption) {
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
		Route:       NewRoute(method, path, handler),
		router:      rg.router,
		routeOption: routeOpt,
	}
	if !rg.root {
		r.middleware = rg.middlewares
	}

	rg.initRouter(r)
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

	if len(rg.paths) > 0 {
		url = fmt.Sprintf("/%s%s", strings.Join(rg.paths, "/"), url)
	}
	plog.Infof("Method: %-6s Router: %-30s Handler: %s", method, url, handlerName)
}

func (rg *RouterGroup) Group(prefix string) *RouterGroup {
	router := rg.router.PathPrefix(prefix).Subrouter()
	g := newGroupWithRouter(router)
	g.prouter = rg.prouter

	if strings.HasPrefix(prefix, "/") {
		prefix = strings.TrimLeft(prefix, "/")
	}
	g.paths = append(rg.paths, prefix)
	return &g
}

func (rg *RouterGroup) staticHandler(prefix string, fs http.FileSystem) handlerFunc {
	return HandleFunc(func(ctx *Context) (Response, error) {
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
			return nil, errors.New("page not found")
		}
		return nil, nil
	})
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
	handler := rg.staticHandler(relativePath, fs)
	rg.GET(urlPattern, handler, opts...)
}

func (rg *RouterGroup) GET(path string, handler handlerFunc, opt ...RouteOption) {
	rg.HandleRoute(http.MethodGet, path, handler, opt...)
}

func (rg *RouterGroup) POST(path string, handler handlerFunc, opt ...RouteOption) {
	rg.HandleRoute(http.MethodPost, path, handler, opt...)
}

func (rg *RouterGroup) PUT(path string, handler handlerFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodPut, path, handler, opts...)
}

func (rg *RouterGroup) PATCH(path string, handler handlerFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodPatch, path, handler, opts...)
}

func (rg *RouterGroup) DELETE(path string, handler handlerFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodDelete, path, handler, opts...)
}

func (rg *RouterGroup) OPTIONS(path string, handler handlerFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodOptions, path, handler, opts...)
}

func (rg *RouterGroup) HEAD(path string, handler handlerFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodHead, path, handler, opts...)
}

func (rg *RouterGroup) TRACE(path string, handler handlerFunc, opts ...RouteOption) {
	rg.HandleRoute(http.MethodTrace, path, handler, opts...)
}

func (rg *RouterGroup) Any(path string, handler handlerFunc, opts ...RouteOption) {
	rg.HandleRoute("", path, handler, opts...)
}
