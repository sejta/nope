package app

import "net/http"

// WithHealth оборачивает handler так, чтобы добавить GET /healthz.
func WithHealth(h http.Handler) http.Handler {
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
