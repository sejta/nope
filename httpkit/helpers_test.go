package httpkit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPKitJSON(t *testing.T) {
	w := httptest.NewRecorder()

	JSON(w, http.StatusCreated, map[string]any{"ok": true})
	if w.Code != http.StatusCreated {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusCreated, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("ожидали Content-Type %q, получили %q", "application/json; charset=utf-8", ct)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("ошибка разбора JSON: %v", err)
	}
	if body["ok"] != true {
		t.Fatalf("ожидали ok=true, получили %v", body["ok"])
	}
}

func TestHTTPKitDecodeJSONValid(t *testing.T) {
	type payload struct {
		A string `json:"a"`
	}
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":"ok"}`))
	var dst payload

	if err := DecodeJSON(r, &dst); err != nil {
		t.Fatalf("не ожидали ошибку: %v", err)
	}
	if dst.A != "ok" {
		t.Fatalf("ожидали A=ok, получили %q", dst.A)
	}
}

func TestHTTPKitDecodeJSONInvalid(t *testing.T) {
	type payload struct {
		A string `json:"a"`
	}
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{"))
	var dst payload

	if err := DecodeJSON(r, &dst); err == nil {
		t.Fatalf("ожидали ошибку")
	}
}

func TestHTTPKitDecodeJSONUnknownField(t *testing.T) {
	type payload struct {
		A string `json:"a"`
	}
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"b":"x"}`))
	var dst payload

	if err := DecodeJSON(r, &dst); err == nil {
		t.Fatalf("ожидали ошибку")
	}
}

func TestHTTPKitWriteNoContent(t *testing.T) {
	w := httptest.NewRecorder()

	WriteNoContent(w)

	if w.Code != http.StatusNoContent {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusNoContent, w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("ожидали пустое тело, получили %q", w.Body.String())
	}
}
