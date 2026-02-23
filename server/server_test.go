package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerGETAndJSON(t *testing.T) {
	s := New(":0")
	s.GET("/ping", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]string{"status": "ok"}, nil
	})

	h, err := s.Handler()
	if err != nil {
		t.Fatalf("handler build failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}

	var payload map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("status=%q want=%q", payload["status"], "ok")
	}
}

func TestGroupPrefix(t *testing.T) {
	s := New(":0")
	api := s.Group("/api")
	api.GET("/users", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"ok": true}, nil
	})

	h, err := s.Handler()
	if err != nil {
		t.Fatalf("handler build failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}
}

func TestGroupRootPath(t *testing.T) {
	s := New(":0")
	api := s.Group("/api")
	api.GET("/", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"ok": true}, nil
	})

	h, err := s.Handler()
	if err != nil {
		t.Fatalf("handler build failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}
}

func TestGroupMiddlewareAppliedOnlyToGroup(t *testing.T) {
	s := New(":0")
	api := s.Group("/api")
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Group", "1")
			next.ServeHTTP(w, r)
		})
	})
	api.GET("/users", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"ok": true}, nil
	})
	s.GET("/ping", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"pong": true}, nil
	})

	h, err := s.Handler()
	if err != nil {
		t.Fatalf("handler build failed: %v", err)
	}

	rrGroup := httptest.NewRecorder()
	h.ServeHTTP(rrGroup, httptest.NewRequest(http.MethodGet, "/api/users", nil))
	if rrGroup.Header().Get("X-Group") != "1" {
		t.Fatalf("group header is missing")
	}

	rrRoot := httptest.NewRecorder()
	h.ServeHTTP(rrRoot, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rrRoot.Header().Get("X-Group") != "" {
		t.Fatalf("group middleware leaked to root route")
	}
}

func TestGlobalMiddlewareOrder(t *testing.T) {
	s := New(":0")
	s.Use(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Order", "A")
				next.ServeHTTP(w, r)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Order", "B")
				next.ServeHTTP(w, r)
			})
		},
	)
	s.GET("/ping", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"pong": true}, nil
	})

	h, err := s.Handler()
	if err != nil {
		t.Fatalf("handler build failed: %v", err)
	}

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ping", nil))
	got := rr.Header().Values("X-Order")
	if len(got) != 2 || got[0] != "A" || got[1] != "B" {
		t.Fatalf("order=%v want=[A B]", got)
	}
}

func TestHandlerReturnsBuildError(t *testing.T) {
	s := New(":0")
	s.GET("ping", func(ctx context.Context, r *http.Request) (any, error) {
		return nil, nil
	})

	_, err := s.Handler()
	if err == nil {
		t.Fatalf("expected build error")
	}
}

func TestEnableHealth(t *testing.T) {
	s := New(":0")
	s.EnableHealth()

	h, err := s.Handler()
	if err != nil {
		t.Fatalf("handler build failed: %v", err)
	}

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}
}

func TestPresetDefaultAddsRequestID(t *testing.T) {
	s := NewWithPreset(":0", PresetDefault)
	s.GET("/ping", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"pong": true}, nil
	})

	h, err := s.Handler()
	if err != nil {
		t.Fatalf("handler build failed: %v", err)
	}

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rr.Header().Get("X-Request-Id") == "" {
		t.Fatalf("missing request id header")
	}
}
