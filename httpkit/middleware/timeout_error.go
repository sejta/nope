package middleware

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"sync"

	apperrors "github.com/sejta/nope/errors"
)

// TimeoutError пишет ответ при превышении deadline, если ответ ещё не начат.
func TimeoutError(write func(w http.ResponseWriter, r *http.Request)) func(http.Handler) http.Handler {
	if write == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tw := &timeoutWriter{ResponseWriter: w}
			done := make(chan struct{})
			go func() {
				next.ServeHTTP(tw, r)
				close(done)
			}()

			select {
			case <-done:
				return
			case <-r.Context().Done():
				if r.Context().Err() != context.DeadlineExceeded {
					return
				}
				tw.writeTimeout(write, r)
				return
			}
		})
	}
}

// DefaultTimeoutError пишет стандартную ошибку таймаута по контракту.
func DefaultTimeoutError(w http.ResponseWriter, r *http.Request) {
	apperrors.WriteError(w, r, apperrors.Timeout())
}

type timeoutWriter struct {
	http.ResponseWriter
	mu      sync.Mutex
	started bool
	timed   bool
}

func (w *timeoutWriter) writeTimeout(write func(http.ResponseWriter, *http.Request), r *http.Request) {
	w.mu.Lock()
	if w.started {
		w.mu.Unlock()
		return
	}
	w.started = true
	w.timed = true
	w.mu.Unlock()
	write(w.ResponseWriter, r)
}

func (w *timeoutWriter) WriteHeader(status int) {
	w.mu.Lock()
	if w.timed {
		w.mu.Unlock()
		return
	}
	w.started = true
	w.mu.Unlock()
	w.ResponseWriter.WriteHeader(status)
}

func (w *timeoutWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	if w.timed {
		w.mu.Unlock()
		return len(p), nil
	}
	w.started = true
	w.mu.Unlock()
	return w.ResponseWriter.Write(p)
}

func (w *timeoutWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *timeoutWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return h.Hijack()
}

func (w *timeoutWriter) Push(target string, opts *http.PushOptions) error {
	p, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return p.Push(target, opts)
}

func (w *timeoutWriter) ReadFrom(r io.Reader) (int64, error) {
	w.mu.Lock()
	if w.timed {
		w.mu.Unlock()
		return 0, nil
	}
	w.started = true
	w.mu.Unlock()
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	return io.Copy(w.ResponseWriter, r)
}
