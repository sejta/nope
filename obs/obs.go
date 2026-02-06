package obs

import (
	"context"
	stderrors "errors"
	"net/http"
	"time"

	"github.com/sejta/nope/app"
	apperrors "github.com/sejta/nope/errors"
	"github.com/sejta/nope/httpkit/middleware"
)

// Logger описывает минимальный контракт логирования событий.
type Logger interface {
	LogRequestEnd(ctx context.Context, e RequestEndEvent)
	LogPanic(ctx context.Context, e PanicEvent)
}

// Metrics описывает минимальный контракт метрик запросов.
type Metrics interface {
	ObserveRequest(method, path string, status int, dur time.Duration)
}

// RequestEndEvent содержит данные о завершении запроса.
type RequestEndEvent struct {
	Method  string
	Path    string
	Status  int
	Bytes   int
	Dur     time.Duration
	ReqID   string
	ErrKind string
}

// PanicEvent содержит данные о panic.
type PanicEvent struct {
	Method string
	Path   string
	ReqID  string
	Value  any
}

// NewHooks создаёт набор hooks для логирования и метрик.
func NewHooks(logger Logger, metrics Metrics) app.Hooks {
	return app.Hooks{
		OnRequestStart: func(ctx context.Context, info app.RequestInfo) context.Context {
			meta := &reqMeta{
				start: time.Now(),
				reqID: middleware.GetRequestID(ctx),
				bytes: &byteCounter{},
			}
			return context.WithValue(ctx, reqMetaKey{}, meta)
		},
		OnRequestEnd: func(ctx context.Context, info app.RequestInfo, res app.ResponseInfo) {
			meta := getReqMeta(ctx)
			event := RequestEndEvent{
				Method:  info.Method,
				Path:    info.Path,
				Status:  res.Status,
				Bytes:   meta.bytesValue(),
				Dur:     duration(meta, res.Duration),
				ReqID:   meta.reqIDValue(),
				ErrKind: errKind(res.Err),
			}
			if logger != nil {
				logger.LogRequestEnd(ctx, event)
			}
			if metrics != nil {
				metrics.ObserveRequest(info.Method, info.Path, res.Status, event.Dur)
			}
		},
		OnPanic: func(ctx context.Context, info app.RequestInfo, recovered any) {
			if logger == nil {
				return
			}
			meta := getReqMeta(ctx)
			logger.LogPanic(ctx, PanicEvent{
				Method: info.Method,
				Path:   info.Path,
				ReqID:  meta.reqIDValue(),
				Value:  recovered,
			})
		},
	}
}

// NewMetricsHook создаёт hooks только для метрик.
func NewMetricsHook(m Metrics) app.Hooks {
	if m == nil {
		return app.Hooks{}
	}
	return app.Hooks{
		OnRequestEnd: func(ctx context.Context, info app.RequestInfo, res app.ResponseInfo) {
			meta := getReqMeta(ctx)
			dur := duration(meta, res.Duration)
			m.ObserveRequest(info.Method, info.Path, res.Status, dur)
		},
	}
}

func errKind(err error) string {
	if err == nil {
		return ""
	}
	var appErr *apperrors.AppError
	if stderrors.As(err, &appErr) && appErr != nil {
		if appErr.Code == apperrors.CodeTimeout || appErr.Status == http.StatusGatewayTimeout {
			return "timeout"
		}
		if appErr.Status >= http.StatusInternalServerError {
			return "internal"
		}
		return ""
	}
	return "internal"
}

func duration(meta *reqMeta, dur time.Duration) time.Duration {
	if dur > 0 {
		return dur
	}
	if meta != nil && !meta.start.IsZero() {
		out := time.Since(meta.start)
		if out > 0 {
			return out
		}
	}
	return time.Nanosecond
}
