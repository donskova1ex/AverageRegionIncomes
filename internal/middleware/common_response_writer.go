package middleware

import "net/http"

type CommonResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewCommonResponseWriter(w http.ResponseWriter) *CommonResponseWriter {
	return &CommonResponseWriter{w, http.StatusOK}
}

func (lrw *CommonResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
