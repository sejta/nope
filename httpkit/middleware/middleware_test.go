package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestReqIDNilContext(t *testing.T) {
	if ReqID(nil) != "" {
		t.Fatalf("ожидали пустой request id для nil контекста")
	}
}
