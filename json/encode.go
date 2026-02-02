package json

import (
	"bytes"
	stdjson "encoding/json"
	"net/http"

	apperrors "github.com/sejta/nope/errors"
)

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type errorPayload struct {
	Error errorBody `json:"error"`
}

// WriteJSON пишет JSON-ответ с заданным статусом.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	var buf bytes.Buffer
	enc := stdjson.NewEncoder(&buf)
	if err := enc.Encode(payload); err != nil {
		writeInternal(w)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func writeInternal(w http.ResponseWriter) {
	payload := errorPayload{
		Error: errorBody{
			Code:    apperrors.CodeInternal,
			Message: apperrors.MsgInternal,
		},
	}
	var buf bytes.Buffer
	_ = stdjson.NewEncoder(&buf).Encode(payload)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write(buf.Bytes())
}
