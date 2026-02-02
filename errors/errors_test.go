package errors

import (
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"
)

func TestWriteErrorNil(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("не ожидали panic: %v", rec)
		}
	}()

	WriteError(w, r, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusOK, w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("ожидали пустое тело, получили %q", w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "" {
		t.Fatalf("ожидали пустой Content-Type, получили %q", ct)
	}
}

func TestWriteErrorAppError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	WriteError(w, r, E(http.StatusBadRequest, "bad_request", "bad"))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusBadRequest, w.Code)
	}
	assertContentType(t, w)
	body := decodeBody(t, w)

	err := getErrorMap(t, body)
	if err["code"] != "bad_request" {
		t.Fatalf("ожидали code %q, получили %v", "bad_request", err["code"])
	}
	if err["message"] != "bad" {
		t.Fatalf("ожидали message %q, получили %v", "bad", err["message"])
	}
	if _, ok := err["fields"]; ok {
		t.Fatalf("не ожидали fields в ответе")
	}
}

func TestWriteErrorWithField(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	err := WithField(E(http.StatusBadRequest, "bad_request", "bad"), "a", "b")
	WriteError(w, r, err)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusBadRequest, w.Code)
	}
	assertContentType(t, w)
	body := decodeBody(t, w)
	payload := getErrorMap(t, body)

	fields, ok := payload["fields"].(map[string]any)
	if !ok {
		t.Fatalf("ожидали fields в ответе")
	}
	if fields["a"] != "b" {
		t.Fatalf("ожидали fields[a]=%q, получили %v", "b", fields["a"])
	}
}

func TestWriteErrorNonAppError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	WriteError(w, r, stderrors.New("x"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusInternalServerError, w.Code)
	}
	assertContentType(t, w)
	body := decodeBody(t, w)

	err := getErrorMap(t, body)
	if err["code"] != CodeInternal {
		t.Fatalf("ожидали code %q, получили %v", CodeInternal, err["code"])
	}
	if err["message"] != MsgInternal {
		t.Fatalf("ожидали message %q, получили %v", MsgInternal, err["message"])
	}
	if _, ok := err["fields"]; ok {
		t.Fatalf("не ожидали fields в ответе")
	}
}

func TestWriteErrorCauseNotInJSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	cause := stderrors.New("boom")
	WriteError(w, r, Wrap(cause, http.StatusBadRequest, "bad_request", "bad"))

	assertContentType(t, w)
	body := decodeBody(t, w)
	payload := getErrorMap(t, body)
	if _, ok := payload["cause"]; ok {
		t.Fatalf("не ожидали cause в ответе")
	}
}

func TestWriteErrorZeroStatus(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	WriteError(w, r, &AppError{Status: 0, Code: "bad_request", Message: "bad"})

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusInternalServerError, w.Code)
	}
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("ошибка разбора JSON: %v", err)
	}
	return body
}

func getErrorMap(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	val, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("ожидали поле error в ответе")
	}
	return val
}

func assertContentType(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("ожидали Content-Type %q, получили %q", "application/json; charset=utf-8", ct)
	}
}
