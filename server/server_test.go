package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sejta/nope/httpkit/middleware"
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

func TestInvalidWildcardPatternDoesNotPanic(t *testing.T) {
	s := New(":0")
	s.GET("/files/*", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"ok": true}, nil
	})

	_, err := s.Handler()
	if err == nil {
		t.Fatalf("expected build error for invalid wildcard pattern")
	}
	if !strings.Contains(err.Error(), "router: empty wildcard name") {
		t.Fatalf("expected router panic reason in error, got %q", err.Error())
	}
}

func TestWildcardConflictDoesNotPanic(t *testing.T) {
	s := New(":0")
	h := func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"ok": true}, nil
	}

	s.GET("/files/*path", h)
	s.GET("/files/*rest", h)

	_, err := s.Handler()
	if err == nil {
		t.Fatalf("expected build error for wildcard conflict")
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

func TestEnableCORSAddsHeadersForAllowedOrigin(t *testing.T) {
	s := New(":0")
	s.EnableCORS(middleware.CORSOptions{
		AllowedOrigins: []string{"https://app.example"},
	})
	s.GET("/ping", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"pong": true}, nil
	})

	h, err := s.Handler()
	if err != nil {
		t.Fatalf("handler build failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://app.example")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "https://app.example" {
		t.Fatalf("missing allow origin header")
	}
}

func TestEnableCORSInvalidOptionsSetBuildError(t *testing.T) {
	s := New(":0")
	s.EnableCORS(middleware.CORSOptions{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	err := s.Validate()
	if err == nil {
		t.Fatalf("expected build error")
	}
	if !strings.Contains(err.Error(), "server: invalid cors options") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateOK(t *testing.T) {
	s := New(":0")
	s.GET("/ping", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"pong": true}, nil
	})

	if err := s.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidateReturnsBuildError(t *testing.T) {
	s := New(":0")
	s.GET("/files/*", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"ok": true}, nil
	})

	err := s.Validate()
	if err == nil {
		t.Fatalf("expected build error")
	}
	if !strings.Contains(err.Error(), "router: empty wildcard name") {
		t.Fatalf("expected router panic reason in error, got %q", err.Error())
	}
}
