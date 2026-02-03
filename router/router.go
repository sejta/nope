package router

import (
	"net/http"
	"sort"
	"strings"
)

// Router — минимальный HTTP-роутер nope.
type Router struct {
	root   *node
	mounts []mount
}

// New создаёт новый Router.
func New() *Router {
	return &Router{
		root: newNode(),
	}
}

// Handle регистрирует handler на метод и паттерн.
func (r *Router) Handle(method, pattern string, h http.Handler) {
	if h == nil {
		panic("router: handler is nil")
	}
	method = strings.ToUpper(method)
	if !isSupportedMethod(method) {
		panic("router: unsupported method")
	}
	if pattern == "" || !strings.HasPrefix(pattern, "/") {
		panic("router: pattern must start with /")
	}

	cur := r.root
	segments := splitPath(pattern)
	for i, seg := range segments {
		switch {
		case strings.HasPrefix(seg, "*"):
			if len(seg) == 1 {
				panic("router: empty wildcard name")
			}
			if i != len(segments)-1 {
				panic("router: wildcard must be the last segment")
			}
			if cur.wildcard != nil && cur.wcName != seg[1:] {
				panic("router: wildcard conflict")
			}
			if cur.wildcard == nil {
				cur.wildcard = newNode()
				cur.wcName = seg[1:]
			}
			cur = cur.wildcard
			goto done
		case strings.HasPrefix(seg, ":"):
			if len(seg) == 1 {
				panic("router: empty param name")
			}
			if cur.param == nil {
				cur.param = newNode()
				cur.param.paramName = seg[1:]
			}
			cur = cur.param
		default:
			next := cur.static[seg]
			if next == nil {
				next = newNode()
				cur.static[seg] = next
			}
			cur = next
		}
	}
done:

	cur.handlers[method] = h
}

// HandleFunc регистрирует handler func на метод и паттерн.
func (r *Router) HandleFunc(method, pattern string, fn http.HandlerFunc) {
	r.Handle(method, pattern, fn)
}

// GET регистрирует handler на GET.
func (r *Router) GET(pattern string, h http.Handler) {
	r.Handle(http.MethodGet, pattern, h)
}

// POST регистрирует handler на POST.
func (r *Router) POST(pattern string, h http.Handler) {
	r.Handle(http.MethodPost, pattern, h)
}

// PUT регистрирует handler на PUT.
func (r *Router) PUT(pattern string, h http.Handler) {
	r.Handle(http.MethodPut, pattern, h)
}

// PATCH регистрирует handler на PATCH.
func (r *Router) PATCH(pattern string, h http.Handler) {
	r.Handle(http.MethodPatch, pattern, h)
}

// DELETE регистрирует handler на DELETE.
func (r *Router) DELETE(pattern string, h http.Handler) {
	r.Handle(http.MethodDelete, pattern, h)
}

// Mount монтирует под‑хендлер на prefix.
func (r *Router) Mount(prefix string, h http.Handler) {
	if h == nil {
		panic("router: handler is nil")
	}
	if prefix == "" || !strings.HasPrefix(prefix, "/") {
		panic("router: prefix must start with /")
	}
	if prefix != "/" && strings.HasSuffix(prefix, "/") {
		panic("router: prefix must not end with /")
	}
	r.mounts = append(r.mounts, mount{prefix: prefix, handler: h})
}

// ServeHTTP реализует net/http.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if m, rest, ok := matchMount(r.mounts, req.URL.Path); ok {
		r.dispatchMount(w, req, m, rest)
		return
	}

	n, params, ok := matchPath(r.root, req.URL.Path)
	if !ok || n == nil || len(n.handlers) == 0 {
		if !isSupportedMethod(req.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if h, ok := n.handlers[req.Method]; ok {
		req = req.WithContext(withParams(req.Context(), params))
		h.ServeHTTP(w, req)
		return
	}

	allow := allowHeader(n.handlers)
	if allow != "" {
		w.Header().Set("Allow", allow)
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (r *Router) dispatchMount(w http.ResponseWriter, req *http.Request, m mount, rest string) {
	req2 := req.Clone(req.Context())
	req2.URL.Path = rest
	m.handler.ServeHTTP(w, req2)
}

func allowHeader(handlers map[string]http.Handler) string {
	if len(handlers) == 0 {
		return ""
	}
	methods := make([]string, 0, len(handlers))
	for method := range handlers {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	return strings.Join(methods, ", ")
}

func isSupportedMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
