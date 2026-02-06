package middleware

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"net/http"
	"time"
)

const requestIDHeader = "X-Request-Id"

type reqIDKey struct{}

// RequestID добавляет request id в контекст и в заголовок ответа.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = newReqID()
		}
		ctx := context.WithValue(r.Context(), reqIDKey{}, id)
		r = r.WithContext(ctx)
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r)
	})
}

// ReqID возвращает request id из контекста.
func ReqID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, ok := ctx.Value(reqIDKey{}).(string)
	if !ok {
		return ""
	}
	return id
}

// GetRequestID возвращает request id из контекста, если он есть.
func GetRequestID(ctx context.Context) string {
	return ReqID(ctx)
}

func newReqID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		now := time.Now().UnixNano()
		binary.LittleEndian.PutUint64(b[:8], uint64(now))
		binary.LittleEndian.PutUint64(b[8:], uint64(now)^uint64(0x9e3779b97f4a7c15))
	}
	return hex.EncodeToString(b[:])
}
