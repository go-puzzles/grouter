package prouter

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-puzzles/plog"
	"github.com/go-puzzles/plog/log"
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

	logMsg = "statusCode=%v duration=%v clientIp=%s method=%s path=%s"
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
		logger: log.New(),
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

func (lm *LogMiddleware) log(ctx context.Context, r *http.Request, rw *ResponseWriter, resp Response, err error, start time.Time) {
	if err != nil && rw.StatusCode() == http.StatusOK {
		rw.WriteHeader(http.StatusInternalServerError)
	}

	path := r.URL.Path
	raw := r.URL.RawQuery
	if raw != "" {
		path = path + "?" + raw
	}

	statusCode := rw.StatusCode()
	clientIp := lm.ClientIP(r)
	method := r.Method
	spendTime := time.Since(start)

	status := "success"
	if statusCode != http.StatusOK {
		status = "failed"
	}

	var logFunc func(ctx context.Context, msg string, v ...any)
	switch {
	case statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices:
		logFunc = lm.logger.Infoc
	case statusCode >= http.StatusMultipleChoices && statusCode < http.StatusBadRequest:
		logFunc = lm.logger.Warnc
	case statusCode >= http.StatusBadRequest && statusCode <= http.StatusNetworkAuthenticationRequired:
		logFunc = lm.logger.Errorc
	default:
		logFunc = lm.logger.Errorc
	}

	args := []any{
		status,
		"statusCode", statusCode,
		"duration", spendTime,
		"clientIp", clientIp,
		"method", method,
		"path", path,
	}

	if err != nil {
		args = append(args, "err", err)
		if resp != nil {
			args = append(args, "errMsg", resp.GetMessage())
		}
	}

	logFunc(ctx, "handle route %v.", args...)
}

func (lm *LogMiddleware) WrapHandler(handler HandleFunc) HandleFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) (Response, error) {
		fmt.Println("into log middleware")
		rw := WrapResponseWriter(w)

		var (
			resp Response
			err  error
		)
		start := time.Now()
		defer func() {
			lm.log(ctx, r, rw, resp, err, start)
		}()

		resp, err = handler(ctx, rw, r, vars)

		return resp, err
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
