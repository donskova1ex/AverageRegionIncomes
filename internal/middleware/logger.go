package middleware

import (
	"log/slog"
	"net/http"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := NewCommonResponseWriter(w)
			next.ServeHTTP(rw, r)
			requestID, _ := r.Context().Value(RequestIDCtxKey).(RequestID)
			logger.Info(
				"http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int64("status", int64(rw.statusCode)),
				slog.String("request_id", string(requestID)),
			)
		})
	}
}
