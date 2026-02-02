package httpkit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/sejta/nope/errors"
)

type errorPayload struct {
	Error struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Fields  map[string]string `json:"fields,omitempty"`
	} `json:"error"`
}

func TestAdaptPayloadOK(t *testing.T) {
	h := func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]string{"ok": "1"}, nil
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	Adapt(h)(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusOK, w.Code)
	}
	assertJSONContentType(t, w)

	var body map[string]string
	decodeJSON(t, w, &body)
	if body["ok"] != "1" {
		t.Fatalf("ожидали ok=1, получили %q", body["ok"])
	}
}

func TestAdaptCreated(t *testing.T) {
	h := func(ctx context.Context, r *http.Request) (any, error) {
		return Created(map[string]string{"ok": "1"}), nil
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	Adapt(h)(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusCreated, w.Code)
	}
	assertJSONContentType(t, w)

	var body map[string]string
	decodeJSON(t, w, &body)
	if body["ok"] != "1" {
		t.Fatalf("ожидали ok=1, получили %q", body["ok"])
	}
}

func TestAdaptNoContent(t *testing.T) {
	h := func(ctx context.Context, r *http.Request) (any, error) {
		return NoContent(), nil
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/", nil)
	Adapt(h)(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusNoContent, w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("ожидали пустое тело, получили %q", w.Body.String())
	}
	if got := w.Header().Get("Content-Type"); got != "" {
		t.Fatalf("не ожидали Content-Type, получили %q", got)
	}
}

func TestAdaptAppError(t *testing.T) {
	h := func(ctx context.Context, r *http.Request) (any, error) {
		return nil, apperrors.E(http.StatusBadRequest, "bad_request", "bad")
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	Adapt(h)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusBadRequest, w.Code)
	}
	assertJSONContentType(t, w)

	var body errorPayload
	decodeJSON(t, w, &body)
	if body.Error.Code != "bad_request" {
		t.Fatalf("ожидали code=bad_request, получили %q", body.Error.Code)
	}
	if body.Error.Message != "bad" {
		t.Fatalf("ожидали message=bad, получили %q", body.Error.Message)
	}
}

func TestAdaptNilPayload(t *testing.T) {
	h := func(ctx context.Context, r *http.Request) (any, error) {
		return nil, nil
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	Adapt(h)(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusOK, w.Code)
	}
	assertJSONContentType(t, w)

	var body any
	decodeJSON(t, w, &body)
	if body != nil {
		t.Fatalf("ожидали null, получили %v", body)
	}
}

func assertJSONContentType(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if got := w.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("ожидали Content-Type %q, получили %q", "application/json; charset=utf-8", got)
	}
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(dst); err != nil {
		t.Fatalf("не удалось декодировать JSON: %v", err)
	}
}
