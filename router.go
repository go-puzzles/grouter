package prouter

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-puzzles/puzzles/plog"
	"github.com/gorilla/mux"
)

const (
	DebugMode = iota
	ReleaseMode
)

var (
	prouterMode = DebugMode
)

type Router interface {
	Routes() []Route
}

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
	host   string
	scheme string
	// middlewares []Middleware
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
	}
	v.RouterGroup.root = true
	v.RouterGroup.prouter = v
	v.parseOptions(opts...)

	return v
}

func NewProuter(opts ...RouterOption) *Prouter {
	v := New(opts...)
	v.UseMiddleware(
		NewLogMiddleware(),
		NewRecoveryMiddleware(),
	)

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

func (v *Prouter) handlerName(handler handlerFunc) string {
	funcName := plog.GetFuncName(handler)
	fs := strings.Split(funcName, ".")

	return fs[len(fs)-1]
}

func (v *Prouter) makeHttpHandler(wr iRoute) http.HandlerFunc {
	handlerName := wr.Handler().Name()

	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		raw := r.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		ctx := &Context{
			Context:   plog.With(r.Context(), "handler", handlerName),
			Writer:    WrapResponseWriter(w),
			Path:      path,
			Method:    r.Method,
			ClientIp:  clientIP(r),
			startTime: time.Now(),
		}
		r = r.Clone(ctx)
		ctx.Request = r

		vars := mux.Vars(ctx.Request)
		if vars == nil {
			vars = make(map[string]string)
		}
		ctx.vars = vars
		ctx.router = v

		handlerFunc := wr.handleSpecifyMiddleware(wr.Handler())

		code, resp := v.packResponseTmpl(handlerFunc.Handle(ctx))
		if code == -1 {
			return
		}

		status := mapCodeToStatus(code)
		_ = WriteJSON(w, status, resp)
	}
}

func (v *Prouter) packResponseTmpl(resp Response, err error) (status int, ret ResponseTmpl) {
	if resp == nil && err == nil {
		return -1, nil
	}

	var (
		code int
		data any
		msg  string
	)

	// resp == nil indicates that an error was sent on this request
	data = func() any {
		if resp == nil {
			return nil
		}

		return resp.GetData()
	}()

	code, msg = v.parseError(resp, err)

	ret = NewResponseTmpl()
	ret.SetCode(code)
	ret.SetMessage(msg)
	ret.SetData(data)

	return code, ret
}

func (v *Prouter) parseError(resp Response, err error) (code int, msg string) {
	if resp != nil {
		code = resp.GetCode()
	}

	if err != nil {
		rErr := new(prouterError)
		if errors.As(err, &rErr) {
			msg = rErr.Message()
			code = rErr.Code()
		} else {
			msg = err.Error()
		}

		if code == 0 {
			code = http.StatusInternalServerError
		}
	}
	return
}
