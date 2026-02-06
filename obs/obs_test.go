package obs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/sejta/nope/app"
	"github.com/sejta/nope/httpkit/middleware"
)

type testLogger struct {
	mu         sync.Mutex
	endCount   int
	panicCount int
	lastEnd    RequestEndEvent
}

func (l *testLogger) LogRequestEnd(ctx context.Context, e RequestEndEvent) {
	_ = ctx
	l.mu.Lock()
	l.endCount++
	l.lastEnd = e
	l.mu.Unlock()
}

func (l *testLogger) LogPanic(ctx context.Context, e PanicEvent) {
	_ = ctx
	_ = e
	l.mu.Lock()
	l.panicCount++
	l.mu.Unlock()
}

func TestObsHooks_LogsEndOnce(t *testing.T) {
	logger := &testLogger{}
	hooks := NewHooks(logger, nil)
	h := Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	serveWithHooks(h, hooks, rec, req)

	if logger.endCount != 1 {
		t.Fatalf("ожидали LogRequestEnd=1, получили %d", logger.endCount)
	}
}

func TestObsHooks_StatusAndDurationNonZero(t *testing.T) {
	logger := &testLogger{}
	hooks := NewHooks(logger, nil)
	h := Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spinUntilTick()
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("nope"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	serveWithHooks(h, hooks, rec, req)

	if logger.lastEnd.Status != http.StatusNotFound {
		t.Fatalf("ожидали status %d, получили %d", http.StatusNotFound, logger.lastEnd.Status)
	}
	if logger.lastEnd.Dur <= 0 {
		t.Fatalf("ожидали dur > 0, получили %v", logger.lastEnd.Dur)
	}
	if logger.lastEnd.Bytes == 0 {
		t.Fatalf("ожидали Bytes > 0, получили %d", logger.lastEnd.Bytes)
	}
}

func TestObsHooks_ReqIDBestEffort(t *testing.T) {
	logger := &testLogger{}
	hooks := NewHooks(logger, nil)
	h := Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	h = middleware.RequestID(h)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	serveWithHooks(h, hooks, rec, req)

	if logger.lastEnd.ReqID == "" {
		t.Fatalf("ожидали непустой req_id")
	}
}

type testRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *testRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *testRecorder) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

func (w *testRecorder) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func serveWithHooks(h http.Handler, hooks app.Hooks, w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	reqInfo := app.RequestInfo{Method: r.Method, Path: r.URL.Path}
	ctx := r.Context()
	if hooks.OnRequestStart != nil {
		ctx = hooks.OnRequestStart(ctx, reqInfo)
		r = r.WithContext(ctx)
	}
	rec := &testRecorder{ResponseWriter: w}
	h.ServeHTTP(rec, r)
	if hooks.OnRequestEnd != nil {
		res := app.ResponseInfo{
			Status:   rec.Status(),
			Duration: time.Since(start),
			Err:      nil,
		}
		hooks.OnRequestEnd(ctx, reqInfo, res)
	}
}

func spinUntilTick() {
	start := time.Now()
	for time.Since(start) == 0 {
	}
}
