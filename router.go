package prouter

import (
	"errors"
	"net"
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

func (v *Prouter) handelrName(handler handlerFunc) string {
	funcName := plog.GetFuncName(handler)
	fs := strings.Split(funcName, ".")

	return fs[len(fs)-1]
}

func remoteIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

func clientIP(r *http.Request) string {
	remoteIP := net.ParseIP(remoteIP(r))
	if remoteIP == nil {
		return ""
	}

	return remoteIP.String()
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
			Request:   r,
			Writer:    w,
			Path:      path,
			Method:    r.Method,
			ClientIp:  clientIP(r),
			startTime: time.Now(),
		}

		vars := mux.Vars(r)
		if vars == nil {
			vars = make(map[string]string)
		}
		ctx.vars = vars
		ctx.router = v

		handlerFunc := wr.handleSpecifyMiddleware(wr.Handler())

		status, resp := v.packResponseTmpl(handlerFunc.Handle(ctx))

		if status == -1 {
			return
		}
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
	code, msg = v.parseError(resp, err)

	data = func() any {
		if resp == nil {
			return nil
		}

		return resp.GetData()
	}()

	code = func() int {
		if resp == nil {
			return 200
		}
		return resp.GetCode()
	}()

	ret = NewResponseTmpl()
	ret.SetCode(code)
	ret.SetMessage(msg)
	ret.SetData(data)

	return code, ret
}

func (v *Prouter) parseError(resp Response, err error) (code int, msg string) {
	if err != nil {
		var rErr *prouterError
		if errors.As(err, &rErr) {
			code = rErr.Code()
			msg = rErr.Message()
		} else if resp != nil {
			err = errors.Join(err, errors.New(resp.GetMessage()))
			code = resp.GetCode()
			msg = err.Error()
		} else {
			code = http.StatusInternalServerError
			msg = err.Error()
		}
	}

	if code == 0 || code == 200 {
		code = http.StatusInternalServerError
	}
	return
}
