package httpkit

import "net/http"

type createdResult struct {
	payloadValue any
}

func (r createdResult) status() int {
	return http.StatusCreated
}

func (r createdResult) payload() any {
	return r.payloadValue
}

type noContentResult struct{}

func (r noContentResult) status() int {
	return http.StatusNoContent
}

func (r noContentResult) payload() any {
	return nil
}

// Created возвращает результат 201 с JSON-полезной нагрузкой.
func Created(payload any) any {
	return createdResult{payloadValue: payload}
}

// NoContent возвращает результат 204 без тела для использования с Adapt.
func NoContent() any {
	return noContentResult{}
}
