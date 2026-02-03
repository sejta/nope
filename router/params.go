package router

import (
	"context"
	"net/http"
)

// RouteParam описывает параметр маршрута.
type RouteParam struct {
	Key   string
	Value string
}

type paramsKey struct{}

// Param возвращает значение параметра по ключу.
func Param(r *http.Request, key string) string {
	for _, p := range Params(r) {
		if p.Key == key {
			return p.Value
		}
	}
	return ""
}

// Params возвращает все параметры маршрута.
func Params(r *http.Request) []RouteParam {
	val := r.Context().Value(paramsKey{})
	if val == nil {
		return nil
	}
	params, ok := val.([]RouteParam)
	if !ok {
		return nil
	}
	return params
}

func withParams(ctx context.Context, params []RouteParam) context.Context {
	if len(params) == 0 {
		return ctx
	}
	return context.WithValue(ctx, paramsKey{}, params)
}
