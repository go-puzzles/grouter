package prouter

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
	
	"github.com/go-puzzles/plog"
	"github.com/go-puzzles/plog/slog"
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
	
	logMsg = "| %s %v %s | %13v | %15s | %s %4v %s| %#v"
)

type LogMiddleware struct {
	logger plog.Logger
}

type LogOption func(*LogMiddleware)

func WithLogger(l plog.Logger) LogOption {
	return func(lm *LogMiddleware) {
		lm.logger = l
	}
}

func NewLogMiddleware() *LogMiddleware {
	return &LogMiddleware{
		logger: slog.New(),
	}
}

func (lm *LogMiddleware) RemoteIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

func (lm *LogMiddleware) ClientIP(r *http.Request) string {
	remoteIP := net.ParseIP(lm.RemoteIP(r))
	if remoteIP == nil {
		return ""
	}
	
	return remoteIP.String()
}

func (lm *LogMiddleware) log(start time.Time, statusCode int, r *http.Request) {
	path := r.URL.Path
	raw := r.URL.RawQuery
	
	clientIp := lm.ClientIP(r)
	method := r.Method
	if raw != "" {
		path = path + "?" + raw
	}
	
	spendTime := time.Since(start)
	
	var logFunc func(msg string, v ...any)
	switch {
	case statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices:
		logFunc = lm.logger.Infof
	case statusCode >= http.StatusMultipleChoices && statusCode < http.StatusBadRequest:
		logFunc = lm.logger.Warnf
	case statusCode >= http.StatusBadRequest && statusCode <= http.StatusNetworkAuthenticationRequired:
		logFunc = lm.logger.Errorf
	default:
		logFunc = lm.logger.Errorf
	}
	
	var statusColor, mthColor, resetColor string
	statusColor = statusCodeColor(statusCode)
	mthColor = methodColor(method)
	resetColor = reset
	
	logFunc(
		logMsg,
		statusColor, statusCode, resetColor,
		spendTime,
		clientIp,
		mthColor, method, resetColor,
		path,
	)
}

func (lm *LogMiddleware) WrapHandler(handler HandleFunc) HandleFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) Response {
		rw := WrapResponseWriter(w)
		
		start := time.Now()
		defer func() {
			lm.log(start, rw.StatusCode(), r)
		}()
		
		return handler(ctx, rw, r, vars)
	}
}

func statusCodeColor(code int) string {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

// MethodColor is the ANSI color for appropriately logging http method to a terminal.
func methodColor(method string) string {
	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}
