package json

import (
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "github.com/sejta/nope/errors"
)

type samplePayload struct {
	A string `json:"a"`
}

func TestDecodeJSONInvalid(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{"))
	var dst samplePayload

	err := DecodeJSON(r, &dst)
	app := assertAppError(t, err)

	assertAppErrorFields(t, app, http.StatusBadRequest, CodeInvalidJSON, MsgInvalidJSON)
}

func TestDecodeJSONEmptyBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	var dst samplePayload

	err := DecodeJSON(r, &dst)
	app := assertAppError(t, err)

	assertAppErrorFields(t, app, http.StatusBadRequest, CodeInvalidJSON, MsgInvalidJSON)
}

func TestDecodeJSONUnexpectedField(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"b":"x"}`))
	var dst samplePayload

	err := DecodeJSON(r, &dst)
	app := assertAppError(t, err)

	assertAppErrorFields(t, app, http.StatusBadRequest, CodeUnexpectedField, MsgUnexpectedField)
	if app.Fields == nil {
		t.Fatalf("ожидали fields в ошибке")
	}
	if app.Fields["b"] != MsgUnexpectedField {
		t.Fatalf("ожидали fields[b]=%q, получили %q", MsgUnexpectedField, app.Fields["b"])
	}
}

func TestDecodeJSONBodyTooLarge(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":"b"}`))
	var dst samplePayload

	err := DecodeJSON(r, &dst, WithMaxBodyBytes(5))
	app := assertAppError(t, err)

	assertAppErrorFields(t, app, http.StatusRequestEntityTooLarge, CodeBodyTooLarge, MsgBodyTooLarge)
}

func TestDecodeJSONValid(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":"ok"}`))
	var dst samplePayload

	err := DecodeJSON(r, &dst)
	if err != nil {
		t.Fatalf("не ожидали ошибку: %v", err)
	}
	if dst.A != "ok" {
		t.Fatalf("ожидали A=ok, получили %q", dst.A)
	}
}

func TestDecodeJSONTrailingGarbage(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":"ok"} {`))
	var dst samplePayload

	err := DecodeJSON(r, &dst)
	app := assertAppError(t, err)

	assertAppErrorFields(t, app, http.StatusBadRequest, CodeInvalidJSON, MsgInvalidJSON)
}

func assertAppError(t *testing.T, err error) *apperrors.AppError {
	t.Helper()
	if err == nil {
		t.Fatalf("ожидали ошибку, получили nil")
	}
	var app *apperrors.AppError
	if !stderrors.As(err, &app) || app == nil {
		t.Fatalf("ожидали AppError, получили %T", err)
	}
	return app
}

func assertAppErrorFields(t *testing.T, app *apperrors.AppError, status int, code, message string) {
	t.Helper()
	if app.Status != status {
		t.Fatalf("ожидали статус %d, получили %d", status, app.Status)
	}
	if app.Code != code {
		t.Fatalf("ожидали code %q, получили %q", code, app.Code)
	}
	if app.Message != message {
		t.Fatalf("ожидали message %q, получили %q", message, app.Message)
	}
}
