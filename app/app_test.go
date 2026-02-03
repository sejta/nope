package app

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWithHealthOK(t *testing.T) {
	called := false
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	h := WithHealth(base)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	h.ServeHTTP(rec, req)

	if called {
		t.Fatalf("ожидали, что базовый handler не будет вызван")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusOK, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("ожидали Content-Type application/json; charset=utf-8, получили %q", ct)
	}
	if body := rec.Body.String(); body != `{"status":"ok"}` {
		t.Fatalf("ожидали body %q, получили %q", `{"status":"ok"}`, body)
	}
}

func TestWithHealthPassThrough(t *testing.T) {
	called := false
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	h := WithHealth(base)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)

	h.ServeHTTP(rec, req)

	if !called {
		t.Fatalf("ожидали вызов базового handler")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusNoContent, rec.Code)
	}
}

func TestWithPprofHandled(t *testing.T) {
	called := false
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNotFound)
	})

	h := WithPprof(base)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)

	h.ServeHTTP(rec, req)

	if called {
		t.Fatalf("ожидали, что базовый handler не будет вызван")
	}
	if rec.Code == http.StatusNotFound {
		t.Fatalf("ожидали, что /debug/pprof/ не вернёт 404")
	}
}

func TestWithPprofPassThrough(t *testing.T) {
	called := false
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	h := WithPprof(base)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)

	h.ServeHTTP(rec, req)

	if !called {
		t.Fatalf("ожидали вызов базового handler")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusNoContent, rec.Code)
	}
}

func TestRunWithListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("не ожидали ошибку listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	h := WithHealth(base)

	cfg := DefaultConfig()
	cfg.Addr = ln.Addr().String()
	cfg.ShutdownTimeout = 200 * time.Millisecond

	done := make(chan error, 1)
	go func() {
		done <- runWithListener(ctx, cfg, h, ln)
	}()

	if err := waitForHealth(cfg.Addr, 2*time.Second); err != nil {
		t.Fatalf("не дождались /healthz: %v", err)
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("не ожидали ошибку Run: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("таймаут ожидания завершения Run")
	}
}

func waitForHealth(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := "http://" + addr + "/healthz"
	client := &http.Client{Timeout: 200 * time.Millisecond}

	for time.Now().Before(deadline) {
		r, err := client.Get(url)
		if err == nil {
			_ = r.Body.Close()
			if r.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return context.DeadlineExceeded
}
