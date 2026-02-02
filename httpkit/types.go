package httpkit

import (
	"context"
	"net/http"
)

// Handler описывает бизнес-хендлер, не зависящий от HTTP-деталей.
type Handler func(ctx context.Context, r *http.Request) (any, error)

type result interface {
	status() int
	payload() any
}
