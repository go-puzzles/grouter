package prouter

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-puzzles/plog"
	"github.com/gorilla/mux"
)

const (
	DebugMode = iota
	ReleaseMode
)

var (
	prouterMode = DebugMode
)

func SetMode(value int64) {
	switch value {
	case DebugMode:
		prouterMode = DebugMode
	case ReleaseMode:
		prouterMode = ReleaseMode
	default:
		panic("Prouter mode unknown: " + strconv.FormatInt(value, 10) + " (available mode: debug release test)")
	}
}

type Prouter struct {
	RouterGroup
	host        string
	scheme      string
	middlewares []Middleware
}

type RouterOption func(v *Prouter)

func WithHost(host string) RouterOption {
	return func(v *Prouter) {
		v.host = host
	}
}

func WithScheme(scheme string) RouterOption {
	return func(v *Prouter) {
		v.scheme = scheme
	}
}

func WithNotFoundHandler(handler http.Handler) RouterOption {
	return func(v *Prouter) {
		v.router.NotFoundHandler = handler
	}
}

func WithMethodNotAllowedHandler(handler http.Handler) RouterOption {
	return func(v *Prouter) {
		v.router.MethodNotAllowedHandler = handler
	}
}

func (v *Prouter) parseOptions(opts ...RouterOption) {
	for _, opt := range opts {
		opt(v)
	}
	if v.host != "" {
		v.router = v.router.Host(v.host).Subrouter()
	}

	if v.scheme != "" {
		v.router = v.router.Schemes(v.host).Subrouter()
	}
}

func New(opts ...RouterOption) *Prouter {
	m := mux.NewRouter()

	v := &Prouter{
		RouterGroup: newGroupWithRouter(m),
		middlewares: make([]Middleware, 0),
	}
	v.RouterGroup.root = true
	v.RouterGroup.prouter = v
	v.parseOptions(opts...)

	return v
}

func NewProuter(opts ...RouterOption) *Prouter {
	v := New(opts...)
	v.UseMiddleware(NewLogMiddleware())

	notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = WriteJSON(w, http.StatusNotFound, ErrorResponse(http.StatusNotFound, "page not found"))
	})
	v.router.NotFoundHandler = notFoundHandler
	v.router.MethodNotAllowedHandler = notFoundHandler
	return v
}
func (v *Prouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.router.ServeHTTP(w, r)
}

func (v *Prouter) ServeHandler() *mux.Router {
	return v.router
}

func (v *Prouter) Run(addr string) error {
	srv := http.Server{
		Addr:    addr,
		Handler: v,
	}
	return srv.ListenAndServe()
}

func (v *Prouter) initRouter(r iRoute) {
	f := v.makeHttpHandler(r)

	vr := r.router.Path(r.Path())
	if r.Method() != "" {
		vr = vr.Methods(r.Method())
	}

	if r.routeOption != nil {
		vr = r.routeOption(vr)
	}

	mr := vr.Handler(f)
	v.debugPrintRoute(r.Method(), mr, r.Handler())
}

func (v *Prouter) UseMiddleware(m ...Middleware) {
	v.middlewares = append(v.middlewares, m...)
}

func (v *Prouter) handleGlobalMiddleware(handler HandleFunc) HandleFunc {
	h := handler
	for _, m := range v.middlewares {
		h = m.WrapHandler(h)
	}

	return h
}

func (v *Prouter) handelrName(handler HandleFunc) string {
	funcName := plog.GetFuncName(handler)
	fs := strings.Split(funcName, ".")

	return fs[len(fs)-1]
}

func (v *Prouter) makeHttpHandler(wr iRoute) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := plog.With(context.Background(), "handler", v.handelrName(wr.Handler()))
		r = r.WithContext(ctx)

		// TODO: parse body data

		vars := mux.Vars(r)
		if vars == nil {
			vars = make(map[string]string)
		}

		handlerFunc := v.handleGlobalMiddleware(wr.Handler())
		handlerFunc = wr.handleSpecifyMiddleware(handlerFunc)

		resp := handlerFunc(ctx, w, r, vars)
		if resp == nil {
			return
		}

		_ = WriteJSON(w, resp.GetCode(), resp.GetData())
	}
}

func (v *Prouter) debugPrintRoute(method string, route *mux.Route, handler HandleFunc) {
	if prouterMode != DebugMode {
		return
	}
	if method == "" {
		method = "ANY"
	}

	handlerName := v.handelrName(handler)
	url, err := route.GetPathTemplate()
	if err != nil {
		plog.Errorf("get handler: %v iRoute url error: %v", handlerName, err)
	}
	plog.Infof("Method: %-6s Router: %-26s Handler: %s", method, url, handlerName)
}
