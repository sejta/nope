package obs

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// NewTextLogger создаёт текстовый логгер для hooks.
func NewTextLogger(w io.Writer) Logger {
	if w == nil {
		return &noopLogger{}
	}
	return &textLogger{w: w}
}

type textLogger struct {
	w  io.Writer
	mu sync.Mutex
}

func (l *textLogger) LogRequestEnd(ctx context.Context, e RequestEndEvent) {
	_ = ctx
	line := fmt.Sprintf(
		"method=%s path=%s status=%d dur=%s bytes=%d req_id=%s\n",
		e.Method,
		e.Path,
		e.Status,
		e.Dur,
		e.Bytes,
		e.ReqID,
	)
	l.mu.Lock()
	_, _ = io.WriteString(l.w, line)
	l.mu.Unlock()
}

func (l *textLogger) LogPanic(ctx context.Context, e PanicEvent) {
	_ = ctx
	line := fmt.Sprintf(
		"panic method=%s path=%s req_id=%s value=%v\n",
		e.Method,
		e.Path,
		e.ReqID,
		e.Value,
	)
	l.mu.Lock()
	_, _ = io.WriteString(l.w, line)
	l.mu.Unlock()
}

type noopLogger struct{}

func (n *noopLogger) LogRequestEnd(ctx context.Context, e RequestEndEvent) {
	_ = ctx
	_ = e
}

func (n *noopLogger) LogPanic(ctx context.Context, e PanicEvent) {
	_ = ctx
	_ = e
}
