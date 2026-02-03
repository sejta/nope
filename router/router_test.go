package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouterStaticMatch(t *testing.T) {
	r := New()
	called := false
	r.GET("/posts", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if !called {
		t.Fatal("handler was not called")
	}
}

func TestRouterParamMatch(t *testing.T) {
	r := New()
	var got string
	r.GET("/posts/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		got = Param(req, "id")
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/123", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if got != "123" {
		t.Fatalf("unexpected param: %q", got)
	}
}

func TestRouterStaticPriority(t *testing.T) {
	r := New()
	var got string
	r.GET("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		got = "param"
		w.WriteHeader(http.StatusNoContent)
	}))
	r.GET("/users/list", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		got = "static"
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/list", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if got != "static" {
		t.Fatalf("unexpected handler: %q", got)
	}
}

func TestRouterNotFound(t *testing.T) {
	r := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestRouterMethodNotAllowed(t *testing.T) {
	r := New()
	r.GET("/x", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	allow := rec.Header().Get("Allow")
	if !strings.Contains(allow, "GET") {
		t.Fatalf("allow header missing GET: %q", allow)
	}
}

func TestRouterMount(t *testing.T) {
	root := New()
	admin := New()
	var got string
	admin.GET("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		got = Param(req, "id")
		w.WriteHeader(http.StatusNoContent)
	}))
	root.Mount("/admin", admin)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/users/42", nil)
	root.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if got != "42" {
		t.Fatalf("unexpected param: %q", got)
	}
}

func TestRouterMountBoundary(t *testing.T) {
	root := New()
	admin := New()
	admin.GET("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	root.Mount("/admin", admin)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/adminX/users/42", nil)
	root.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestRouterTrailingSlash(t *testing.T) {
	r := New()
	r.GET("/users", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}
