package router_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "github.com/sejta/nope/errors"
	"github.com/sejta/nope/httpkit"
	"github.com/sejta/nope/httpkit/middleware"
	"github.com/sejta/nope/router"
)

type response struct {
	status  int
	headers http.Header
	body    []byte
}

func doRequest(t *testing.T, h http.Handler, method, path string, headers map[string]string, body io.Reader) response {
	t.Helper()
	if body == nil {
		body = http.NoBody
	}
	req := httptest.NewRequest(method, path, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return response{
		status:  rec.Code,
		headers: rec.Header(),
		body:    rec.Body.Bytes(),
	}
}

func buildRouter() *router.Router {
	r := router.New()
	text := func(s string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(s))
		})
	}

	r.GET("/", text("root"))
	r.GET("/healthz", text("ok"))
	r.GET("/a", text("a"))
	r.GET("/a/b", text("ab"))
	r.GET("/assets/logo.png", text("logo"))
	r.GET("/assets/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := router.Param(r, "id")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("id:" + id))
	}))
	r.GET("/assets/*path", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := router.Param(r, "path")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(path))
	}))
	r.GET("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := router.Param(r, "id")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(id))
	}))
	r.GET("/json", httpkit.Adapt(func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"ok": true}, nil
	}))
	r.GET("/err", httpkit.Adapt(func(ctx context.Context, r *http.Request) (any, error) {
		return nil, apperrors.E(http.StatusBadRequest, "bad_request", "bad")
	}))
	r.GET("/nocontent", httpkit.Adapt(func(ctx context.Context, r *http.Request) (any, error) {
		return httpkit.NoContent(), nil
	}))
	r.GET("/panic", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	r.GET("/ctx", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := middleware.ReqID(r.Context())
		if id == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(id))
	}))
	r.GET("/order", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	api := router.New()
	api.GET("/", text("api-root"))
	api.GET("/ping", text("api-ping"))
	r.Mount("/api", api)

	return r
}

func TestRouterIntegrationBasics(t *testing.T) {
	h := http.Handler(buildRouter())

	cases := []struct {
		name              string
		method            string
		path              string
		wantStatus        int
		wantBody          string
		wantBodyContains  string
		wantHeader        map[string]string
		wantAllowContains string
		check             func(t *testing.T, resp response)
	}{
		{
			name:       "health ok",
			method:     http.MethodGet,
			path:       "/healthz",
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
		{
			name:       "health not found",
			method:     http.MethodGet,
			path:       "/health",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "root",
			method:     http.MethodGet,
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody:   "root",
		},
		{
			name:       "exact match",
			method:     http.MethodGet,
			path:       "/a/b",
			wantStatus: http.StatusOK,
			wantBody:   "ab",
		},
		{
			name:       "trailing slash policy",
			method:     http.MethodGet,
			path:       "/a/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:              "method not allowed",
			method:            http.MethodPost,
			path:              "/healthz",
			wantStatus:        http.StatusMethodNotAllowed,
			wantAllowContains: http.MethodGet,
		},
		{
			name:              "unsupported method",
			method:            http.MethodOptions,
			path:              "/healthz",
			wantStatus:        http.StatusMethodNotAllowed,
			wantAllowContains: http.MethodGet,
		},
		{
			name:       "mount root",
			method:     http.MethodGet,
			path:       "/api",
			wantStatus: http.StatusOK,
			wantBody:   "api-root",
		},
		{
			name:       "mount nested",
			method:     http.MethodGet,
			path:       "/api/ping",
			wantStatus: http.StatusOK,
			wantBody:   "api-ping",
		},
		{
			name:       "mount boundary",
			method:     http.MethodGet,
			path:       "/apiX",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "route params",
			method:     http.MethodGet,
			path:       "/users/42",
			wantStatus: http.StatusOK,
			wantBody:   "42",
		},
		{
			name:       "wildcard basic",
			method:     http.MethodGet,
			path:       "/assets/a/b.png",
			wantStatus: http.StatusOK,
			wantBody:   "a/b.png",
		},
		{
			name:       "wildcard empty slash",
			method:     http.MethodGet,
			path:       "/assets/",
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
		{
			name:       "wildcard empty no slash",
			method:     http.MethodGet,
			path:       "/assets",
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
		{
			name:       "wildcard static priority",
			method:     http.MethodGet,
			path:       "/assets/logo.png",
			wantStatus: http.StatusOK,
			wantBody:   "logo",
		},
		{
			name:       "wildcard param priority",
			method:     http.MethodGet,
			path:       "/assets/123",
			wantStatus: http.StatusOK,
			wantBody:   "id:123",
		},
		{
			name:       "wildcard boundary",
			method:     http.MethodGet,
			path:       "/assetsX/a",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "wildcard encoded path",
			method:     http.MethodGet,
			path:       "/assets/a%20b/c",
			wantStatus: http.StatusOK,
			wantBody:   "a b/c",
		},
		{
			name:              "wildcard method not allowed",
			method:            http.MethodPost,
			path:              "/assets/a",
			wantStatus:        http.StatusMethodNotAllowed,
			wantAllowContains: http.MethodGet,
		},
		{
			name:       "not found",
			method:     http.MethodGet,
			path:       "/missing",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "json response",
			method:     http.MethodGet,
			path:       "/json",
			wantStatus: http.StatusOK,
			check: func(t *testing.T, resp response) {
				if ct := resp.headers.Get("Content-Type"); ct != "application/json; charset=utf-8" {
					t.Fatalf("ожидали Content-Type %q, получили %q", "application/json; charset=utf-8", ct)
				}
				var body map[string]any
				if err := json.Unmarshal(resp.body, &body); err != nil {
					t.Fatalf("ошибка разбора JSON: %v", err)
				}
				if body["ok"] != true {
					t.Fatalf("ожидали ok=true, получили %v", body["ok"])
				}
			},
		},
		{
			name:       "error response",
			method:     http.MethodGet,
			path:       "/err",
			wantStatus: http.StatusBadRequest,
			check: func(t *testing.T, resp response) {
				var body map[string]any
				if err := json.Unmarshal(resp.body, &body); err != nil {
					t.Fatalf("ошибка разбора JSON: %v", err)
				}
				err, ok := body["error"].(map[string]any)
				if !ok {
					t.Fatalf("ожидали объект error в ответе")
				}
				if err["code"] != "bad_request" {
					t.Fatalf("ожидали code %q, получили %v", "bad_request", err["code"])
				}
				if err["message"] != "bad" {
					t.Fatalf("ожидали message %q, получили %v", "bad", err["message"])
				}
			},
		},
		{
			name:       "no content",
			method:     http.MethodGet,
			path:       "/nocontent",
			wantStatus: http.StatusNoContent,
			check: func(t *testing.T, resp response) {
				if len(resp.body) != 0 {
					t.Fatalf("ожидали пустое тело, получили %q", string(resp.body))
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doRequest(t, h, tc.method, tc.path, nil, nil)
			if resp.status != tc.wantStatus {
				t.Fatalf("ожидали статус %d, получили %d", tc.wantStatus, resp.status)
			}
			if tc.wantBody != "" && string(resp.body) != tc.wantBody {
				t.Fatalf("ожидали body %q, получили %q", tc.wantBody, string(resp.body))
			}
			if tc.wantBodyContains != "" && !strings.Contains(string(resp.body), tc.wantBodyContains) {
				t.Fatalf("ожидали body содержит %q, получили %q", tc.wantBodyContains, string(resp.body))
			}
			if tc.wantAllowContains != "" {
				allow := resp.headers.Get("Allow")
				if !strings.Contains(allow, tc.wantAllowContains) {
					t.Fatalf("ожидали Allow содержит %q, получили %q", tc.wantAllowContains, allow)
				}
			}
			for k, v := range tc.wantHeader {
				if got := resp.headers.Get(k); got != v {
					t.Fatalf("ожидали заголовок %s=%q, получили %q", k, v, got)
				}
			}
			if tc.check != nil {
				tc.check(t, resp)
			}
		})
	}
}

func TestRouterIntegrationWildcardConflicts(t *testing.T) {
	t.Run("wildcard not last segment", func(t *testing.T) {
		r := router.New()
		defer func() {
			if rec := recover(); rec == nil {
				t.Fatalf("ожидали panic")
			}
		}()
		r.GET("/assets/*path/x", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	})

	t.Run("wildcard conflict", func(t *testing.T) {
		r := router.New()
		r.GET("/assets/*a", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer func() {
			if rec := recover(); rec == nil {
				t.Fatalf("ожидали panic")
			}
		}()
		r.GET("/assets/*b", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	})
}

func TestRouterIntegrationMiddlewareOrder(t *testing.T) {
	order := make([]string, 0, 4)
	mw := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-before")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	h := http.Handler(buildRouter())
	h = mw("mw1")(mw("mw2")(h))

	resp := doRequest(t, h, http.MethodGet, "/order", nil, nil)
	if resp.status != http.StatusNoContent {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusNoContent, resp.status)
	}

	want := []string{"mw1-before", "mw2-before", "mw2-after", "mw1-after"}
	if len(order) != len(want) {
		t.Fatalf("ожидали %d шагов, получили %d", len(want), len(order))
	}
	for i, v := range want {
		if order[i] != v {
			t.Fatalf("ожидали порядок %v, получили %v", want, order)
		}
	}
}

func TestRouterIntegrationContextValue(t *testing.T) {
	h := http.Handler(buildRouter())
	h = middleware.RequestID(h)

	resp := doRequest(t, h, http.MethodGet, "/ctx", nil, nil)
	if resp.status != http.StatusOK {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusOK, resp.status)
	}
	if len(resp.body) == 0 {
		t.Fatalf("ожидали непустой request id")
	}
}

func TestRouterIntegrationRecover(t *testing.T) {
	h := http.Handler(buildRouter())
	h = middleware.Recover(h)

	resp := doRequest(t, h, http.MethodGet, "/panic", nil, nil)
	if resp.status != http.StatusInternalServerError {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusInternalServerError, resp.status)
	}
	if ct := resp.headers.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("ожидали Content-Type %q, получили %q", "application/json; charset=utf-8", ct)
	}
	var body map[string]any
	if err := json.Unmarshal(resp.body, &body); err != nil {
		t.Fatalf("ошибка разбора JSON: %v", err)
	}
	err, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("ожидали объект error в ответе")
	}
	if err["code"] != apperrors.CodeInternal {
		t.Fatalf("ожидали code %q, получили %v", apperrors.CodeInternal, err["code"])
	}
}

func TestRouterIntegrationHeaderFromMiddleware(t *testing.T) {
	setHeader := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "ok")
			next.ServeHTTP(w, r)
		})
	}

	h := http.Handler(buildRouter())
	h = setHeader(h)

	resp := doRequest(t, h, http.MethodGet, "/healthz", nil, nil)
	if resp.status != http.StatusOK {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusOK, resp.status)
	}
	if got := resp.headers.Get("X-Test"); got != "ok" {
		t.Fatalf("ожидали X-Test=ok, получили %q", got)
	}
}
