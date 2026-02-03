package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout устанавливает deadline на обработку запроса.
func Timeout(d time.Duration) func(next http.Handler) http.Handler {
	if d <= 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
