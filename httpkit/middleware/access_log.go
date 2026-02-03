package middleware

import (
	"net/http"
	"time"
)

// Logger — минимальный интерфейс логгера.
type Logger interface {
	Printf(format string, args ...any)
}

// AccessLog логирует запрос по завершению.
func AccessLog(l Logger) func(next http.Handler) http.Handler {
	if l == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lw := &logWriter{ResponseWriter: w}
			next.ServeHTTP(lw, r)
			if lw.status == 0 {
				lw.status = http.StatusOK
			}
			durMS := time.Since(start).Milliseconds()
			reqID := ReqID(r.Context())
			if reqID != "" {
				l.Printf("method=%s path=%s status=%d dur_ms=%d bytes=%d req_id=%s", r.Method, r.URL.Path, lw.status, durMS, lw.bytes, reqID)
				return
			}
			l.Printf("method=%s path=%s status=%d dur_ms=%d bytes=%d", r.Method, r.URL.Path, lw.status, durMS, lw.bytes)
		})
	}
}

type logWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *logWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *logWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}
