package prouter

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *ResponseWriter) WriteHeader(code int) {
	if code > 0 && w.statusCode != code {
		w.statusCode = code
	}
}

func (w *ResponseWriter) StatusCode() int {
	return w.statusCode
}

func WrapResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}
