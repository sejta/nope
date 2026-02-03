package middleware

import (
	"bufio"
	"io"
	"net"
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
			if reqID == "" {
				reqID = lw.Header().Get(requestIDHeader)
				if reqID == "" {
					reqID = r.Header.Get(requestIDHeader)
				}
			}
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

func (w *logWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *logWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return h.Hijack()
}

func (w *logWriter) Push(target string, opts *http.PushOptions) error {
	p, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return p.Push(target, opts)
}

func (w *logWriter) ReadFrom(r io.Reader) (int64, error) {
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		if w.status == 0 {
			w.status = http.StatusOK
		}
		n, err := rf.ReadFrom(r)
		w.bytes += int(n)
		return n, err
	}
	if w.status == 0 {
		w.status = http.StatusOK
	}
	buf := make([]byte, 32*1024)
	var total int64
	for {
		nr, er := r.Read(buf)
		if nr > 0 {
			nw, ew := w.Write(buf[:nr])
			total += int64(nw)
			if ew != nil {
				return total, ew
			}
			if nw != nr {
				return total, io.ErrShortWrite
			}
		}
		if er != nil {
			if er == io.EOF {
				return total, nil
			}
			return total, er
		}
	}
}
