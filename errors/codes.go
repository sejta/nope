package errors

const (
	// CodeInternal — стабильный код для внутренних ошибок.
	CodeInternal = "internal_error"
	// MsgInternal — безопасное сообщение для внутренних ошибок.
	MsgInternal = "internal error"
	// CodeTimeout — стабильный код для ошибки таймаута.
	CodeTimeout = "timeout"
	// MsgTimeout — безопасное сообщение для ошибки таймаута.
	MsgTimeout = "request timed out"
)
