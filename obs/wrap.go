package obs

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/sejta/nope/httpkit/middleware"
)

// Wrap оборачивает handler и считает записанные байты для hooks.
func Wrap(next http.Handler) http.Handler {
	if next == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		meta := getReqMeta(r.Context())
		if meta == nil {
			meta = &reqMeta{bytes: &byteCounter{}}
			r = r.WithContext(context.WithValue(r.Context(), reqMetaKey{}, meta))
		}
		if meta.reqID == "" {
			meta.reqID = middleware.GetRequestID(r.Context())
		}
		cw := &countingWriter{ResponseWriter: w, counter: meta.bytes}
		next.ServeHTTP(cw, r)
		if meta.reqID == "" {
			meta.reqID = r.Header.Get("X-Request-Id")
			if meta.reqID == "" {
				meta.reqID = w.Header().Get("X-Request-Id")
			}
		}
	})
}

type reqMetaKey struct{}

type reqMeta struct {
	start time.Time
	reqID string
	bytes *byteCounter
}

func getReqMeta(ctx context.Context) *reqMeta {
	if ctx == nil {
		return nil
	}
	meta, ok := ctx.Value(reqMetaKey{}).(*reqMeta)
	if !ok || meta == nil {
		return nil
	}
	return meta
}

func (m *reqMeta) bytesValue() int {
	if m == nil || m.bytes == nil {
		return 0
	}
	return m.bytes.Value()
}

func (m *reqMeta) reqIDValue() string {
	if m == nil {
		return ""
	}
	return m.reqID
}

type byteCounter struct {
	n int64
}

func (c *byteCounter) Add(n int64) {
	if c == nil || n <= 0 {
		return
	}
	atomic.AddInt64(&c.n, n)
}

func (c *byteCounter) Value() int {
	if c == nil {
		return 0
	}
	n := atomic.LoadInt64(&c.n)
	maxInt := int64(^uint(0) >> 1)
	if n > maxInt {
		return int(maxInt)
	}
	return int(n)
}

type countingWriter struct {
	http.ResponseWriter
	counter *byteCounter
}

func (w *countingWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	if w.counter != nil {
		w.counter.Add(int64(n))
	}
	return n, err
}

func (w *countingWriter) Unwrap() http.ResponseWriter {
	if uw, ok := w.ResponseWriter.(interface{ Unwrap() http.ResponseWriter }); ok {
		return uw.Unwrap()
	}
	return w.ResponseWriter
}

func (w *countingWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *countingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return h.Hijack()
}

func (w *countingWriter) Push(target string, opts *http.PushOptions) error {
	p, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return p.Push(target, opts)
}

func (w *countingWriter) ReadFrom(r io.Reader) (int64, error) {
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		n, err := rf.ReadFrom(r)
		if w.counter != nil {
			w.counter.Add(n)
		}
		return n, err
	}
	n, err := io.Copy(w.ResponseWriter, r)
	if w.counter != nil {
		w.counter.Add(n)
	}
	return n, err
}
