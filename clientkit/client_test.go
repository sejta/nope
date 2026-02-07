package clientkit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type userResp struct {
	ID int `json:"id"`
}

func TestDoJSONSuccessDecode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": 7}`))
	}))
	t.Cleanup(srv.Close)

	var out userResp
	meta, err := DoJSON[struct{}, userResp](context.Background(), DefaultClient(), http.MethodGet, srv.URL, nil, &out, nil)
	if err != nil {
		t.Fatalf("ожидали nil error, получили %v", err)
	}
	if meta.Status != http.StatusOK {
		t.Fatalf("ожидали 200, получили %v", meta.Status)
	}
	if out.ID != 7 {
		t.Fatalf("ожидали id=7, получили %v", out.ID)
	}
}

func TestDoJSONNon2xxReturnsHTTPErrorWithBodyTruncated(t *testing.T) {
	body := strings.Repeat("x", int(defaultMaxErrBody)+1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	meta, err := DoJSON[struct{}, userResp](context.Background(), DefaultClient(), http.MethodGet, srv.URL, nil, nil, nil)
	if err == nil {
		t.Fatalf("ожидали ошибку, получили nil")
	}
	if meta.Status != http.StatusBadRequest {
		t.Fatalf("ожидали 400, получили %v", meta.Status)
	}
	he, ok := IsHTTPError(err)
	if !ok {
		t.Fatalf("ожидали HTTPError, получили %T", err)
	}
	if int64(len(he.Body)) > defaultMaxErrBody {
		t.Fatalf("ожидали body <= %d, получили %d", defaultMaxErrBody, len(he.Body))
	}
}

func TestDoJSONNoBodyOutNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	meta, err := DoJSON[struct{}, userResp](context.Background(), DefaultClient(), http.MethodGet, srv.URL, nil, nil, nil)
	if err != nil {
		t.Fatalf("ожидали nil error, получили %v", err)
	}
	if meta.Status != http.StatusOK {
		t.Fatalf("ожидали 200, получили %v", meta.Status)
	}
}

func TestDoJSONDisallowUnknownFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": 1, "extra": "x"}`))
	}))
	t.Cleanup(srv.Close)

	var out userResp
	_, err := DoJSON[struct{}, userResp](context.Background(), DefaultClient(), http.MethodGet, srv.URL, nil, &out, &JSONOptions{DisallowUnknownFields: true})
	if err == nil {
		t.Fatalf("ожидали ошибку, получили nil")
	}
}

func TestDoJSONMaxBodyLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := `{"data":"` + strings.Repeat("a", 64) + `"}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}))
	t.Cleanup(srv.Close)

	type dataResp struct {
		Data string `json:"data"`
	}
	var out dataResp
	_, err := DoJSON[struct{}, dataResp](context.Background(), DefaultClient(), http.MethodGet, srv.URL, nil, &out, &JSONOptions{MaxBody: 16})
	if err == nil {
		t.Fatalf("ожидали ошибку лимита, получили nil")
	}
}

func TestDoJSONContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": 1}`))
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out userResp
	_, err := DoJSON[struct{}, userResp](ctx, DefaultClient(), http.MethodGet, srv.URL, nil, &out, nil)
	if err == nil {
		t.Fatalf("ожидали ошибку отменённого контекста, получили nil")
	}
}
