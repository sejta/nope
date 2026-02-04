package httpkit

import (
	"net/http"

	jsonkit "github.com/sejta/nope/json"
)

// JSON пишет JSON-ответ с заданным статусом (net/http style).
func JSON(w http.ResponseWriter, status int, v any) {
	jsonkit.WriteJSON(w, status, v)
}

// DecodeJSON читает и валидирует JSON-тело запроса.
func DecodeJSON(r *http.Request, dst any) error {
	return jsonkit.DecodeJSON(r, dst)
}

// WriteNoContent пишет ответ 204 без тела (net/http style).
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
