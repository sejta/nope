package json

import (
	stdjson "encoding/json"
	stderrors "errors"
	"io"
	"net/http"
	"strings"

	apperrors "github.com/sejta/nope/errors"
)

// DecodeJSON читает и валидирует JSON-тело запроса.
func DecodeJSON(r *http.Request, dst any, opts ...Option) error {
	if r == nil || r.Body == nil {
		return apperrors.E(http.StatusBadRequest, CodeInvalidJSON, MsgInvalidJSON)
	}

	cfg := decodeOptions{
		maxBodyBytes: DefaultMaxBodyBytes,
		strict:       true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	dec := stdjson.NewDecoder(http.MaxBytesReader(nil, r.Body, cfg.maxBodyBytes))
	if cfg.strict {
		dec.DisallowUnknownFields()
	}

	if err := dec.Decode(dst); err != nil {
		return mapDecodeError(err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return apperrors.E(http.StatusBadRequest, CodeInvalidJSON, MsgInvalidJSON)
	}
	return nil
}

func mapDecodeError(err error) error {
	if err == io.EOF {
		return apperrors.E(http.StatusBadRequest, CodeInvalidJSON, MsgInvalidJSON)
	}
	var maxErr *http.MaxBytesError
	if stderrors.As(err, &maxErr) {
		return apperrors.E(http.StatusRequestEntityTooLarge, CodeBodyTooLarge, MsgBodyTooLarge)
	}
	if name, ok := unknownFieldName(err); ok {
		app := apperrors.E(http.StatusBadRequest, CodeUnexpectedField, MsgUnexpectedField)
		return apperrors.WithField(app, name, MsgUnexpectedField)
	}
	return apperrors.E(http.StatusBadRequest, CodeInvalidJSON, MsgInvalidJSON)
}

func unknownFieldName(err error) (string, bool) {
	const prefix = "json: unknown field "
	msg := err.Error()
	if !strings.HasPrefix(msg, prefix) {
		return "", false
	}
	name := strings.TrimPrefix(msg, prefix)
	name = strings.Trim(name, "\"")
	if name == "" {
		return "", false
	}
	return name, true
}
