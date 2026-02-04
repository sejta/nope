package httpkit

import (
	"net/http"

	apperrors "github.com/sejta/nope/errors"
	jsonkit "github.com/sejta/nope/json"
)

// Adapt преобразует Handler в http.HandlerFunc.
//
// Поведение:
// - успех → JSON-ответ (200/201/204)
// - ошибка → errors.WriteError (единый error contract)
func Adapt(h Handler) http.HandlerFunc {
	if h == nil {
		panic("httpkit: nil handler")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := h(r.Context(), r)
		if err != nil {
			if setter, ok := w.(interface{ SetErr(error) }); ok {
				setter.SetErr(err)
			}
			apperrors.WriteError(w, r, err)
			return
		}

		if resResult, ok := res.(result); ok {
			status := resResult.status()
			if status == http.StatusNoContent {
				w.WriteHeader(status)
				return
			}
			jsonkit.WriteJSON(w, status, resResult.payload())
			return
		}

		jsonkit.WriteJSON(w, http.StatusOK, res)
	}
}
