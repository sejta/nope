package clientkit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultTimeout    = 10 * time.Second
	defaultMaxBody    = int64(1 << 20) // 1MB
	defaultMaxErrBody = int64(8 << 10) // 8KB
)

// ResponseMeta содержит метаданные HTTP-ответа без тела.
type ResponseMeta struct {
	Status  int
	Headers http.Header
}

// JSONOptions управляет заголовками и лимитами JSON-запроса.
type JSONOptions struct {
	Headers               map[string]string
	MaxBody               int64
	MaxErrBody            int64
	DisallowUnknownFields bool
}

// Options управляет заголовками и лимитом чтения тела для обычных запросов.
type Options struct {
	Headers    map[string]string
	MaxErrBody int64
}

// HTTPError описывает non-2xx ответ.
type HTTPError struct {
	Status     int
	StatusText string
	URL        string
	Method     string
	Body       []byte
}

// Error возвращает текст ошибки для HTTPError.
func (e *HTTPError) Error() string {
	return "clientkit: http error " + strconv.Itoa(e.Status) + " " + e.StatusText + " " + e.Method + " " + e.URL
}

// IsHTTPError проверяет, что ошибка является HTTPError.
func IsHTTPError(err error) (*HTTPError, bool) {
	var he *HTTPError
	if errors.As(err, &he) {
		return he, true
	}
	return nil, false
}

// DefaultClient возвращает http.Client с разумным таймаутом.
func DefaultClient() *http.Client {
	return &http.Client{Timeout: defaultTimeout}
}

// DoJSON выполняет JSON-запрос и декодирует ответ в out.
func DoJSON[Req any, Resp any](
	ctx context.Context,
	c *http.Client,
	method string,
	url string,
	req *Req,
	out *Resp,
	opt *JSONOptions,
) (ResponseMeta, error) {
	var meta ResponseMeta
	if c == nil {
		return meta, errors.New("clientkit: nil client")
	}
	if ctx == nil {
		return meta, errors.New("clientkit: nil context")
	}
	opts := normalizeJSONOptions(opt)

	var body io.Reader
	if req != nil {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		if err := enc.Encode(req); err != nil {
			return meta, err
		}
		body = &buf
	}

	hreq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return meta, err
	}
	if req != nil {
		hreq.Header.Set("Content-Type", "application/json")
	}
	applyHeaders(hreq.Header, opts.Headers)

	resp, err := c.Do(hreq)
	if err != nil {
		return meta, err
	}
	defer resp.Body.Close()

	meta = ResponseMeta{Status: resp.StatusCode, Headers: resp.Header.Clone()}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		bodyBytes := readErrBody(resp.Body, opts.MaxErrBody)
		return meta, &HTTPError{
			Status:     resp.StatusCode,
			StatusText: resp.Status,
			URL:        hreq.URL.String(),
			Method:     hreq.Method,
			Body:       bodyBytes,
		}
	}

	if out == nil {
		if err := discardBody(resp.Body, opts.MaxBody); err != nil {
			return meta, err
		}
		return meta, nil
	}

	payload, err := readBodyLimit(resp.Body, opts.MaxBody)
	if err != nil {
		return meta, err
	}
	dec := json.NewDecoder(bytes.NewReader(payload))
	if opts.DisallowUnknownFields {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(out); err != nil {
		return meta, err
	}
	if err := ensureJSONEOF(dec); err != nil {
		return meta, err
	}
	return meta, nil
}

// GetJSON выполняет GET и декодирует JSON-ответ.
func GetJSON[Resp any](ctx context.Context, c *http.Client, url string, out *Resp, opt *JSONOptions) (ResponseMeta, error) {
	return DoJSON[struct{}, Resp](ctx, c, http.MethodGet, url, nil, out, opt)
}

// PostJSON выполняет POST с JSON-телом и декодирует JSON-ответ.
func PostJSON[Req any, Resp any](ctx context.Context, c *http.Client, url string, req *Req, out *Resp, opt *JSONOptions) (ResponseMeta, error) {
	return DoJSON[Req, Resp](ctx, c, http.MethodPost, url, req, out, opt)
}

// Do выполняет обычный запрос и возвращает тело ответа.
func Do(ctx context.Context, c *http.Client, req *http.Request, opt *Options) (ResponseMeta, []byte, error) {
	var meta ResponseMeta
	if c == nil {
		return meta, nil, errors.New("clientkit: nil client")
	}
	if ctx == nil {
		return meta, nil, errors.New("clientkit: nil context")
	}
	if req == nil {
		return meta, nil, errors.New("clientkit: nil request")
	}
	opts := normalizeOptions(opt)

	req = req.WithContext(ctx)
	applyHeaders(req.Header, opts.Headers)

	resp, err := c.Do(req)
	if err != nil {
		return meta, nil, err
	}
	defer resp.Body.Close()

	meta = ResponseMeta{Status: resp.StatusCode, Headers: resp.Header.Clone()}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		bodyBytes := readErrBody(resp.Body, opts.MaxErrBody)
		return meta, nil, &HTTPError{
			Status:     resp.StatusCode,
			StatusText: resp.Status,
			URL:        req.URL.String(),
			Method:     req.Method,
			Body:       bodyBytes,
		}
	}

	payload, err := readBodyLimit(resp.Body, opts.MaxErrBody)
	if err != nil {
		return meta, nil, err
	}
	return meta, payload, nil
}

func normalizeJSONOptions(opt *JSONOptions) JSONOptions {
	if opt == nil {
		return JSONOptions{MaxBody: defaultMaxBody, MaxErrBody: defaultMaxErrBody}
	}

	out := *opt
	if out.MaxBody <= 0 {
		out.MaxBody = defaultMaxBody
	}
	if out.MaxErrBody <= 0 {
		out.MaxErrBody = defaultMaxErrBody
	}
	return out
}

func normalizeOptions(opt *Options) Options {
	if opt == nil {
		return Options{MaxErrBody: defaultMaxErrBody}
	}

	out := *opt
	if out.MaxErrBody <= 0 {
		out.MaxErrBody = defaultMaxErrBody
	}
	return out
}

func applyHeaders(h http.Header, headers map[string]string) {
	for k, v := range headers {
		h.Set(k, v)
	}
}

func readBodyLimit(r io.Reader, limit int64) ([]byte, error) {
	if limit <= 0 {
		return nil, errors.New("clientkit: invalid body limit")
	}
	buf, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(buf)) > limit {
		return nil, errors.New("clientkit: response body too large")
	}
	return buf, nil
}

func readErrBody(r io.Reader, limit int64) []byte {
	if limit <= 0 {
		return nil
	}
	buf, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil
	}
	if int64(len(buf)) > limit {
		return buf[:limit]
	}
	return buf
}

func discardBody(r io.Reader, limit int64) error {
	if limit <= 0 {
		return nil
	}
	_, err := io.Copy(io.Discard, io.LimitReader(r, limit))
	return err
}

func ensureJSONEOF(dec *json.Decoder) error {
	var tail any
	if err := dec.Decode(&tail); err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}
	return errors.New("clientkit: invalid json (extra data)")
}
