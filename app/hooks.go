package app

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"time"

	apperrors "github.com/sejta/nope/errors"
)

// Hooks описывает точки расширения для логирования и метрик.
type Hooks struct {
	OnRequestStart func(ctx context.Context, info RequestInfo) context.Context
	OnRequestEnd   func(ctx context.Context, info RequestInfo, res ResponseInfo)
	OnPanic        func(ctx context.Context, info RequestInfo, recovered any)
}

// RequestInfo содержит минимальные данные о запросе.
type RequestInfo struct {
	Method string
	Path   string // фактический URL.Path, не pattern
}

// ResponseInfo содержит итоговые данные об ответе.
type ResponseInfo struct {
	Status   int
	Duration time.Duration
	Err      error // nil если не было
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	err         error
	wroteHeader bool
	wroteBody   bool
}

func (w *statusRecorder) WriteHeader(status int) {
	w.wroteHeader = true
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(p []byte) (int, error) {
	w.wroteBody = true
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(p)
}

func (w *statusRecorder) SetErr(err error) {
	w.err = err
}

func (w *statusRecorder) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *statusRecorder) Err() error {
	return w.err
}

func (w *statusRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *statusRecorder) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return h.Hijack()
}

func (w *statusRecorder) Push(target string, opts *http.PushOptions) error {
	p, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return p.Push(target, opts)
}

func (w *statusRecorder) ReadFrom(r io.Reader) (int64, error) {
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		if w.status == 0 {
			w.status = http.StatusOK
		}
		return rf.ReadFrom(r)
	}
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return io.Copy(w.ResponseWriter, r)
}

func wrapHooks(h http.Handler, hooks Hooks) http.Handler {
	if hooks.OnRequestStart == nil && hooks.OnRequestEnd == nil && hooks.OnPanic == nil {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqInfo := RequestInfo{Method: r.Method, Path: r.URL.Path}
		ctx := r.Context()
		if hooks.OnRequestStart != nil {
			ctx = hooks.OnRequestStart(ctx, reqInfo)
			r = r.WithContext(ctx)
		}

		rec := &statusRecorder{ResponseWriter: w}
		var hookErr error

		defer func() {
			if recov := recover(); recov != nil {
				if hooks.OnPanic != nil {
					hooks.OnPanic(ctx, reqInfo, recov)
				}
				hookErr = apperrors.E(http.StatusInternalServerError, apperrors.CodeInternal, apperrors.MsgInternal)
				rec.SetErr(hookErr)
				if !rec.wroteHeader && !rec.wroteBody {
					apperrors.WriteError(rec, r, hookErr)
				}
			}

			if hooks.OnRequestEnd != nil {
				err := rec.Err()
				if hookErr != nil {
					err = hookErr
				}
				resInfo := ResponseInfo{
					Status:   rec.Status(),
					Duration: time.Since(start),
					Err:      err,
				}
				hooks.OnRequestEnd(ctx, reqInfo, resInfo)
			}
		}()

		h.ServeHTTP(rec, r)
	})
}
