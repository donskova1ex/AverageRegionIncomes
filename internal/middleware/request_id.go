package middleware

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

type RequestID string

const RequestIDCtxKey RequestID = "request_id"

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		ctx := context.WithValue(r.Context(), RequestIDCtxKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
