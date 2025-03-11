package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := NewCommonResponseWriter(w)
			next.ServeHTTP(rw, r)

			durationMs := time.Since(start).Milliseconds()

			log := logger.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			queryParams := r.URL.Query()
			if len(queryParams) > 0 {
				for key, values := range queryParams {
					log = log.With(slog.String(key, strings.Join(values, ",")))
				}
			}

			requestID, _ := r.Context().Value(RequestIDCtxKey).(RequestID)
			log = log.With(
				slog.Int64("status", int64(rw.statusCode)),
				slog.String("request_id", string(requestID)),
				slog.Int64("duration_ms", durationMs),
			)

			log.Info("http request")
		})
	}
}
