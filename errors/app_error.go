package errors

import stderrors "errors"

// AppError описывает продуктовую ошибку для клиента.
type AppError struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	Cause   error
}

// Error возвращает безопасное текстовое описание ошибки.
func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" {
		return e.Code
	}
	return MsgInternal
}

// E создаёт AppError с заданными параметрами.
func E(status int, code, message string) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

// Wrap создаёт AppError и сохраняет первопричину.
func Wrap(cause error, status int, code, message string) *AppError {
	if cause == nil {
		return E(status, code, message)
	}
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// WithField добавляет одно поле к AppError.
func WithField(err error, key, val string) error {
	if err == nil {
		return nil
	}
	app := asAppError(err)
	out := cloneAppError(app)
	if out.Fields == nil {
		out.Fields = make(map[string]string, 1)
	}
	out.Fields[key] = val
	return out
}

// WithFields добавляет несколько полей к AppError.
func WithFields(err error, fields map[string]string) error {
	if err == nil {
		return nil
	}
	if len(fields) == 0 {
		return err
	}
	app := asAppError(err)
	out := cloneAppError(app)
	if out.Fields == nil {
		out.Fields = make(map[string]string, len(fields))
	}
	for k, v := range fields {
		out.Fields[k] = v
	}
	return out
}

func asAppError(err error) *AppError {
	var app *AppError
	if stderrors.As(err, &app) && app != nil {
		return app
	}
	return &AppError{
		Status:  500,
		Code:    CodeInternal,
		Message: MsgInternal,
		Cause:   err,
	}
}

func cloneAppError(in *AppError) *AppError {
	if in == nil {
		return &AppError{
			Status:  500,
			Code:    CodeInternal,
			Message: MsgInternal,
		}
	}
	out := *in
	if len(in.Fields) > 0 {
		out.Fields = make(map[string]string, len(in.Fields))
		for k, v := range in.Fields {
			out.Fields[k] = v
		}
	}
	return &out
}
