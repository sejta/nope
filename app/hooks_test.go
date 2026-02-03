package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	apperrors "github.com/sejta/nope/errors"
	"github.com/sejta/nope/httpkit"
)

type ctxKey string

func TestHooksStartEndCalled(t *testing.T) {
	var startCount int32
	var endCount int32
	var gotReq RequestInfo
	var gotRes ResponseInfo

	hooks := Hooks{
		OnRequestStart: func(ctx context.Context, info RequestInfo) context.Context {
			atomic.AddInt32(&startCount, 1)
			gotReq = info
			return ctx
		},
		OnRequestEnd: func(ctx context.Context, info RequestInfo, res ResponseInfo) {
			atomic.AddInt32(&endCount, 1)
			gotRes = res
		},
	}

	h := wrapHooks(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), hooks)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if atomic.LoadInt32(&startCount) != 1 {
		t.Fatalf("ожидали startCount=1, получили %d", startCount)
	}
	if atomic.LoadInt32(&endCount) != 1 {
		t.Fatalf("ожидали endCount=1, получили %d", endCount)
	}
	if gotReq.Method != http.MethodGet || gotReq.Path != "/healthz" {
		t.Fatalf("ожидали method=%q path=%q, получили %q %q", http.MethodGet, "/healthz", gotReq.Method, gotReq.Path)
	}
	if gotRes.Status != http.StatusOK {
		t.Fatalf("ожидали status %d, получили %d", http.StatusOK, gotRes.Status)
	}
	if gotRes.Duration < 0 {
		t.Fatalf("ожидали неотрицательный duration, получили %v", gotRes.Duration)
	}
	if gotRes.Err != nil {
		t.Fatalf("не ожидали Err, получили %v", gotRes.Err)
	}
}

func TestHooksContextPropagation(t *testing.T) {
	hooks := Hooks{
		OnRequestStart: func(ctx context.Context, info RequestInfo) context.Context {
			return context.WithValue(ctx, ctxKey("k"), "v")
		},
	}

	h := wrapHooks(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val, _ := r.Context().Value(ctxKey("k")).(string)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(val))
	}), hooks)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Body.String() != "v" {
		t.Fatalf("ожидали body %q, получили %q", "v", rec.Body.String())
	}
}

func TestHooksStatusCapture(t *testing.T) {
	var gotStatus int
	hooks := Hooks{
		OnRequestEnd: func(ctx context.Context, info RequestInfo, res ResponseInfo) {
			gotStatus = res.Status
		},
	}

	h := wrapHooks(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), hooks)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if gotStatus != http.StatusNoContent {
		t.Fatalf("ожидали status %d, получили %d", http.StatusNoContent, gotStatus)
	}
}

func TestHooksErrorCase(t *testing.T) {
	var gotStatus int
	var gotErr error
	hooks := Hooks{
		OnRequestEnd: func(ctx context.Context, info RequestInfo, res ResponseInfo) {
			gotStatus = res.Status
			gotErr = res.Err
		},
	}

	h := wrapHooks(httpkit.Adapt(func(ctx context.Context, r *http.Request) (any, error) {
		return nil, apperrors.E(http.StatusBadRequest, "bad_request", "bad")
	}), hooks)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if gotStatus != http.StatusBadRequest {
		t.Fatalf("ожидали status %d, получили %d", http.StatusBadRequest, gotStatus)
	}
	if gotErr == nil {
		t.Fatalf("ожидали Err, получили nil")
	}
}

func TestHooksPanicCase(t *testing.T) {
	var panicCount int32
	var endStatus int
	hooks := Hooks{
		OnPanic: func(ctx context.Context, info RequestInfo, recovered any) {
			atomic.AddInt32(&panicCount, 1)
		},
		OnRequestEnd: func(ctx context.Context, info RequestInfo, res ResponseInfo) {
			endStatus = res.Status
		},
	}

	h := wrapHooks(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}), hooks)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if atomic.LoadInt32(&panicCount) != 1 {
		t.Fatalf("ожидали OnPanic=1, получили %d", panicCount)
	}
	if endStatus != http.StatusInternalServerError {
		t.Fatalf("ожидали status %d, получили %d", http.StatusInternalServerError, endStatus)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("ошибка разбора JSON: %v", err)
	}
	payload, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("ожидали объект error в ответе")
	}
	if payload["code"] != apperrors.CodeInternal {
		t.Fatalf("ожидали code %q, получили %v", apperrors.CodeInternal, payload["code"])
	}
}

func TestHooksNoop(t *testing.T) {
	h := wrapHooks(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), Hooks{})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusOK, rec.Code)
	}
}
