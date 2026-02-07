package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCORSNoOrigin(t *testing.T) {
	called := false
	mw := CORS(CORSOptions{AllowedOrigins: []string{"https://app.example"}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if !called {
		t.Fatalf("ожидали вызов next")
	}
	if w.Header().Get(corsHeaderAllowOrigin) != "" {
		t.Fatalf("ожидали отсутствие CORS заголовков")
	}
}

func TestCORSAllowedOriginSimpleRequest(t *testing.T) {
	mw := CORS(CORSOptions{AllowedOrigins: []string{"https://app.example"}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(corsHeaderOrigin, "https://app.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Header().Get(corsHeaderAllowOrigin) != "https://app.example" {
		t.Fatalf("ожидали Allow-Origin, получили %q", w.Header().Get(corsHeaderAllowOrigin))
	}
	if !strings.Contains(w.Header().Get(corsHeaderVary), corsHeaderOrigin) {
		t.Fatalf("ожидали Vary: Origin")
	}
}

func TestCORSDisallowedOrigin(t *testing.T) {
	called := false
	mw := CORS(CORSOptions{AllowedOrigins: []string{"https://app.example"}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(corsHeaderOrigin, "https://evil.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if !called {
		t.Fatalf("ожидали вызов next")
	}
	if w.Header().Get(corsHeaderAllowOrigin) != "" {
		t.Fatalf("ожидали отсутствие CORS заголовков")
	}
}

func TestCORSPreflightAllowed(t *testing.T) {
	mw := CORS(CORSOptions{AllowedOrigins: []string{"https://app.example"}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set(corsHeaderOrigin, "https://app.example")
	req.Header.Set(corsHeaderRequestMethod, http.MethodPost)
	req.Header.Set(corsHeaderRequestHeaders, "X-Test, Content-Type")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("ожидали 204, получили %v", w.Code)
	}
	if w.Header().Get(corsHeaderAllowOrigin) != "https://app.example" {
		t.Fatalf("ожидали Allow-Origin, получили %q", w.Header().Get(corsHeaderAllowOrigin))
	}
	if w.Header().Get(corsHeaderAllowMethods) == "" {
		t.Fatalf("ожидали Allow-Methods")
	}
	if w.Header().Get(corsHeaderAllowHeaders) == "" {
		t.Fatalf("ожидали Allow-Headers")
	}
	if !strings.Contains(w.Header().Get(corsHeaderVary), corsHeaderOrigin) {
		t.Fatalf("ожидали Vary: Origin")
	}
	if !strings.Contains(w.Header().Get(corsHeaderVary), corsHeaderRequestMethod) {
		t.Fatalf("ожидали Vary: Access-Control-Request-Method")
	}
	if !strings.Contains(w.Header().Get(corsHeaderVary), corsHeaderRequestHeaders) {
		t.Fatalf("ожидали Vary: Access-Control-Request-Headers")
	}
}

func TestCORSPreflightDisallowed(t *testing.T) {
	called := false
	mw := CORS(CORSOptions{AllowedOrigins: []string{"https://app.example"}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set(corsHeaderOrigin, "https://evil.example")
	req.Header.Set(corsHeaderRequestMethod, http.MethodPost)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if called {
		t.Fatalf("не ожидали вызов next")
	}
	if w.Code != http.StatusNoContent {
		t.Fatalf("ожидали 204, получили %v", w.Code)
	}
	if w.Header().Get(corsHeaderAllowOrigin) != "" {
		t.Fatalf("ожидали отсутствие CORS заголовков")
	}
}

func TestCORSAllowCredentials(t *testing.T) {
	mw := CORS(CORSOptions{AllowedOrigins: []string{"https://app.example"}, AllowCredentials: true})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(corsHeaderOrigin, "https://app.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Header().Get(corsHeaderAllowCredentials) != "true" {
		t.Fatalf("ожидали Allow-Credentials=true")
	}
	if w.Header().Get(corsHeaderAllowOrigin) != "https://app.example" {
		t.Fatalf("ожидали Allow-Origin, получили %q", w.Header().Get(corsHeaderAllowOrigin))
	}
}

func TestCORSConfigInvalidWildcardWithCredentials(t *testing.T) {
	defer func() {
		if rec := recover(); rec == nil {
			t.Fatalf("ожидали panic")
		}
	}()
	_ = CORS(CORSOptions{AllowedOrigins: []string{"*"}, AllowCredentials: true})
}

func TestCORSExposeHeaders(t *testing.T) {
	mw := CORS(CORSOptions{AllowedOrigins: []string{"https://app.example"}, ExposedHeaders: []string{"X-Trace"}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(corsHeaderOrigin, "https://app.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Header().Get(corsHeaderExposeHeaders) != "X-Trace" {
		t.Fatalf("ожидали Expose-Headers, получили %q", w.Header().Get(corsHeaderExposeHeaders))
	}
}

func TestCORSMaxAge(t *testing.T) {
	mw := CORS(CORSOptions{AllowedOrigins: []string{"https://app.example"}, MaxAge: 10 * time.Second})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set(corsHeaderOrigin, "https://app.example")
	req.Header.Set(corsHeaderRequestMethod, http.MethodPost)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Header().Get(corsHeaderMaxAge) != "10" {
		t.Fatalf("ожидали Max-Age=10, получили %q", w.Header().Get(corsHeaderMaxAge))
	}
}
