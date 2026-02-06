package errors

import "net/http"

// Timeout возвращает стандартную ошибку таймаута запроса.
func Timeout() *AppError {
	return E(http.StatusGatewayTimeout, CodeTimeout, MsgTimeout)
}
