package errors

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
)

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type errorPayload struct {
	Error errorBody `json:"error"`
}

// WriteError пишет единый JSON-ответ для ошибки.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	_ = r

	status := 500
	body := errorBody{
		Code:    CodeInternal,
		Message: MsgInternal,
	}

	var app *AppError
	if stderrors.As(err, &app) && app != nil {
		if app.Status != 0 {
			status = app.Status
		}
		if app.Code != "" {
			body.Code = app.Code
		}
		if app.Message != "" {
			body.Message = app.Message
		}
		if len(app.Fields) > 0 {
			body.Fields = make(map[string]string, len(app.Fields))
			for k, v := range app.Fields {
				body.Fields[k] = v
			}
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorPayload{Error: body})
}
