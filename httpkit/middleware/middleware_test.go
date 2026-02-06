package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	apperrors "github.com/sejta/nope/errors"
)

type testLogger struct {
	lines []string
}

func (l *testLogger) Printf(format string, args ...any) {
	l.lines = append(l.lines, fmt.Sprintf(format, args...))
}

type errorPayload struct {
	Error struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Fields  map[string]string `json:"fields,omitempty"`
	} `json:"error"`
}

func TestRequestIDUsesHeader(t *testing.T) {
	var got string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = ReqID(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestIDHeader, "abc")
	w := httptest.NewRecorder()

	RequestID(inner).ServeHTTP(w, req)

	if got != "abc" {
		t.Fatalf("ожидали request id %q, получили %q", "abc", got)
	}
	if header := w.Header().Get(requestIDHeader); header != "abc" {
		t.Fatalf("ожидали заголовок %q, получили %q", "abc", header)
	}
}

func TestRequestIDGeneratesWhenMissing(t *testing.T) {
	var got string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = ReqID(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	RequestID(inner).ServeHTTP(w, req)

	if got == "" {
		t.Fatalf("ожидали непустой request id")
	}
	if header := w.Header().Get(requestIDHeader); header == "" {
		t.Fatalf("ожидали непустой заголовок request id")
	}
	if got != w.Header().Get(requestIDHeader) {
		t.Fatalf("ожидали одинаковые значения в контексте и заголовке")
	}
}

func TestRecoverPanic(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Recover(inner).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusInternalServerError, w.Code)
	}
	var body errorPayload
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("не удалось декодировать JSON: %v", err)
	}
	if body.Error.Code != apperrors.CodeInternal {
		t.Fatalf("ожидали code %q, получили %q", apperrors.CodeInternal, body.Error.Code)
	}
}

func TestTimeoutSetsDeadline(t *testing.T) {
	var ok bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok = r.Context().Deadline()
		w.WriteHeader(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Timeout(10*time.Millisecond)(inner).ServeHTTP(w, req)

	if !ok {
		t.Fatalf("ожидали deadline в контексте")
	}
}

func TestAccessLogWritesLine(t *testing.T) {
	logger := &testLogger{}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(requestIDHeader, "abc")
	w := httptest.NewRecorder()

	h := RequestID(AccessLog(logger)(inner))
	h.ServeHTTP(w, req)

	if len(logger.lines) != 1 {
		t.Fatalf("ожидали одну строку лога, получили %d", len(logger.lines))
	}
	line := logger.lines[0]
	checks := []string{
		"method=GET",
		"path=/x",
		"status=201",
		"bytes=2",
		"req_id=abc",
	}
	for _, part := range checks {
		if !strings.Contains(line, part) {
			t.Fatalf("ожидали в логе %q, получили %q", part, line)
		}
	}
}

func TestAccessLogGetsReqIDFromHeader(t *testing.T) {
	logger := &testLogger{}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()

	h := AccessLog(logger)(RequestID(inner))
	h.ServeHTTP(w, req)

	if len(logger.lines) != 1 {
		t.Fatalf("ожидали одну строку лога, получили %d", len(logger.lines))
	}
	line := logger.lines[0]
	reqID := extractField(line, "req_id=")
	if reqID == "" {
		t.Fatalf("ожидали непустой req_id в логе, получили %q", line)
	}
}

func TestTimeoutErrorWritesOnDeadline(t *testing.T) {
	started := make(chan struct{})
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := newManualDeadlineContext(req.Context())
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		TimeoutError(DefaultTimeoutError)(inner).ServeHTTP(w, req)
		close(done)
	}()

	<-started
	ctx.trigger(context.DeadlineExceeded)
	<-done

	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusGatewayTimeout, w.Code)
	}
	var body errorPayload
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("не удалось декодировать JSON: %v", err)
	}
	if body.Error.Code != apperrors.CodeTimeout {
		t.Fatalf("ожидали code %q, получили %q", apperrors.CodeTimeout, body.Error.Code)
	}
	if body.Error.Message != apperrors.MsgTimeout {
		t.Fatalf("ожидали message %q, получили %q", apperrors.MsgTimeout, body.Error.Message)
	}
}

func TestTimeoutErrorDoesNotOverwriteResponse(t *testing.T) {
	wrote := make(chan struct{})
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		close(wrote)
		<-r.Context().Done()
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := newManualDeadlineContext(req.Context())
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		TimeoutError(DefaultTimeoutError)(inner).ServeHTTP(w, req)
		close(done)
	}()

	<-wrote
	ctx.trigger(context.DeadlineExceeded)
	<-done

	if w.Code != http.StatusOK {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusOK, w.Code)
	}
	if body := w.Body.String(); body != "ok" {
		t.Fatalf("ожидали тело %q, получили %q", "ok", body)
	}
}

func TestTimeoutErrorDoesNotOverrideErrorResponse(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := apperrors.E(http.StatusBadRequest, "bad_request", "bad request")
		apperrors.WriteError(w, r, err)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := newManualDeadlineContext(req.Context())
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	TimeoutError(DefaultTimeoutError)(inner).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusBadRequest, w.Code)
	}
	var body errorPayload
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("не удалось декодировать JSON: %v", err)
	}
	if body.Error.Code == apperrors.CodeTimeout {
		t.Fatalf("не ожидали code %q", apperrors.CodeTimeout)
	}
}

func TestTimeoutErrorIgnoresCanceledContext(t *testing.T) {
	started := make(chan struct{})
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := newManualDeadlineContext(req.Context())
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		TimeoutError(DefaultTimeoutError)(inner).ServeHTTP(w, req)
		close(done)
	}()

	<-started
	ctx.trigger(context.Canceled)
	<-done

	if w.Code != http.StatusOK {
		t.Fatalf("ожидали статус %d, получили %d", http.StatusOK, w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("ожидали пустое тело, получили %q", w.Body.String())
	}
}

func TestReqIDNilContext(t *testing.T) {
	if ReqID(context.TODO()) != "" {
		t.Fatalf("ожидали пустой request id для пустого контекста")
	}
}

func extractField(line, key string) string {
	idx := strings.Index(line, key)
	if idx == -1 {
		return ""
	}
	start := idx + len(key)
	end := start
	for end < len(line) && line[end] != ' ' {
		end++
	}
	return line[start:end]
}

type manualDeadlineContext struct {
	parent   context.Context
	deadline time.Time
	done     chan struct{}
	once     sync.Once
	mu       sync.Mutex
	err      error
}

func newManualDeadlineContext(parent context.Context) *manualDeadlineContext {
	return &manualDeadlineContext{
		parent:   parent,
		deadline: time.Now().Add(time.Hour),
		done:     make(chan struct{}),
	}
}

func (c *manualDeadlineContext) Deadline() (time.Time, bool) {
	return c.deadline, true
}

func (c *manualDeadlineContext) Done() <-chan struct{} {
	return c.done
}

func (c *manualDeadlineContext) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *manualDeadlineContext) Value(key any) any {
	return c.parent.Value(key)
}

func (c *manualDeadlineContext) trigger(err error) {
	c.mu.Lock()
	c.err = err
	c.mu.Unlock()
	c.once.Do(func() {
		close(c.done)
	})
}
