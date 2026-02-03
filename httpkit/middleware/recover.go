package middleware

import (
	"net/http"

	apperrors "github.com/sejta/nope/errors"
)

// Recover перехватывает panic и возвращает безопасный JSON-ответ 500.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				err := apperrors.E(http.StatusInternalServerError, apperrors.CodeInternal, apperrors.MsgInternal)
				apperrors.WriteError(w, r, err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
