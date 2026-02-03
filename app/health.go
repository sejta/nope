package app

import "net/http"

// WithHealth оборачивает handler так, чтобы добавить GET /healthz.
// Если путь занят — health не добавлять и вернуть исходный handler.
func WithHealth(h http.Handler) http.Handler {
	if hasHealthz(h) {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		h.ServeHTTP(w, r)
	})
}

type handlerLookup interface {
	Handler(*http.Request) (http.Handler, string)
}

func hasHealthz(h http.Handler) bool {
	checker, ok := h.(handlerLookup)
	if !ok {
		return false
	}

	req, err := http.NewRequest(http.MethodGet, "/healthz", nil)
	if err != nil {
		return false
	}
	_, pattern := checker.Handler(req)
	return pattern == "/healthz"
}
