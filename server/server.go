package server

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sejta/nope/app"
	"github.com/sejta/nope/httpkit"
	"github.com/sejta/nope/httpkit/middleware"
	"github.com/sejta/nope/router"
)

var (
	errNilHandler     = errors.New("server: handler is nil")
	errInvalidPath    = errors.New("server: path must start with /")
	errInvalidPrefix  = errors.New("server: prefix must start with /")
	errTrailingPrefix = errors.New("server: prefix must not end with /")
	errInvalidRoute   = errors.New("server: invalid route registration")
)

// Middleware описывает HTTP middleware в формате net/http.
type Middleware func(http.Handler) http.Handler

// Preset определяет готовый набор настроек фасада.
type Preset int

const (
	// PresetMinimal включает только базовый runtime без middleware.
	PresetMinimal Preset = iota
	// PresetDefault включает минимально полезный набор middleware.
	PresetDefault
)

// Server предоставляет упрощённый API для регистрации роутов и запуска сервера.
type Server struct {
	addr              string
	r                 *router.Router
	cfg               app.Config
	globalMiddleware  []Middleware
	enableHealthRoute bool
	enablePprofRoute  bool
	buildErr          error
}

// Group объединяет роуты с общим prefix и локальными middleware.
type Group struct {
	s      *Server
	prefix string
	mws    []Middleware
}

// New создаёт новый Server с минимальным preset.
func New(addr string) *Server {
	return NewWithPreset(addr, PresetMinimal)
}

// NewWithPreset создаёт новый Server с выбранным preset.
func NewWithPreset(addr string, preset Preset) *Server {
	cfg := app.DefaultConfig()
	if addr != "" {
		cfg.Addr = addr
	}

	s := &Server{
		addr: cfg.Addr,
		r:    router.New(),
		cfg:  cfg,
	}

	s.applyPreset(preset)
	return s
}

// Config возвращает текущую runtime-конфигурацию.
func (s *Server) Config() app.Config {
	return s.cfg
}

// SetConfig заменяет runtime-конфигурацию целиком.
func (s *Server) SetConfig(cfg app.Config) {
	s.cfg = cfg
	if s.cfg.Addr == "" {
		s.cfg.Addr = s.addr
	}
}

// Use добавляет глобальные middleware для всех роутов.
func (s *Server) Use(mw ...Middleware) {
	for _, one := range mw {
		if one == nil {
			continue
		}
		s.globalMiddleware = append(s.globalMiddleware, one)
	}
}

// EnableHealth включает стандартный маршрут GET /healthz.
func (s *Server) EnableHealth() {
	s.enableHealthRoute = true
}

// EnablePprof включает маршруты /debug/pprof/*.
func (s *Server) EnablePprof() {
	s.enablePprofRoute = true
}

// Group создаёт группу роутов с общим prefix.
func (s *Server) Group(prefix string) *Group {
	if err := validatePrefix(prefix); err != nil {
		s.setBuildErr(err)
	}
	return &Group{s: s, prefix: prefix}
}

// GET регистрирует GET-хендлер по контракту httpkit.Handler.
func (s *Server) GET(routePath string, h httpkit.Handler) {
	s.handle(http.MethodGet, routePath, h, nil)
}

// POST регистрирует POST-хендлер по контракту httpkit.Handler.
func (s *Server) POST(routePath string, h httpkit.Handler) {
	s.handle(http.MethodPost, routePath, h, nil)
}

// PUT регистрирует PUT-хендлер по контракту httpkit.Handler.
func (s *Server) PUT(routePath string, h httpkit.Handler) {
	s.handle(http.MethodPut, routePath, h, nil)
}

// PATCH регистрирует PATCH-хендлер по контракту httpkit.Handler.
func (s *Server) PATCH(routePath string, h httpkit.Handler) {
	s.handle(http.MethodPatch, routePath, h, nil)
}

// DELETE регистрирует DELETE-хендлер по контракту httpkit.Handler.
func (s *Server) DELETE(routePath string, h httpkit.Handler) {
	s.handle(http.MethodDelete, routePath, h, nil)
}

// Run запускает HTTP-сервер с context.Background().
func (s *Server) Run() error {
	return s.RunContext(context.Background())
}

// RunContext запускает HTTP-сервер с указанным контекстом.
func (s *Server) RunContext(ctx context.Context) error {
	h, err := s.Handler()
	if err != nil {
		return err
	}
	cfg := s.cfg
	if cfg.Addr == "" {
		cfg.Addr = s.addr
	}
	return app.Run(ctx, cfg, h)
}

// Handler собирает итоговый http.Handler с учётом middleware и app-обёрток.
func (s *Server) Handler() (http.Handler, error) {
	if s.buildErr != nil {
		return nil, s.buildErr
	}

	var h http.Handler = s.r
	h = applyMiddleware(h, s.globalMiddleware)
	if s.enableHealthRoute {
		h = app.WithHealth(h)
	}
	if s.enablePprofRoute {
		h = app.WithPprof(h)
	}
	return h, nil
}

// Use добавляет middleware только для текущей группы.
func (g *Group) Use(mw ...Middleware) {
	for _, one := range mw {
		if one == nil {
			continue
		}
		g.mws = append(g.mws, one)
	}
}

// GET регистрирует GET-хендлер в группе.
func (g *Group) GET(routePath string, h httpkit.Handler) {
	g.handle(http.MethodGet, routePath, h)
}

// POST регистрирует POST-хендлер в группе.
func (g *Group) POST(routePath string, h httpkit.Handler) {
	g.handle(http.MethodPost, routePath, h)
}

// PUT регистрирует PUT-хендлер в группе.
func (g *Group) PUT(routePath string, h httpkit.Handler) {
	g.handle(http.MethodPut, routePath, h)
}

// PATCH регистрирует PATCH-хендлер в группе.
func (g *Group) PATCH(routePath string, h httpkit.Handler) {
	g.handle(http.MethodPatch, routePath, h)
}

// DELETE регистрирует DELETE-хендлер в группе.
func (g *Group) DELETE(routePath string, h httpkit.Handler) {
	g.handle(http.MethodDelete, routePath, h)
}

func (g *Group) handle(method, routePath string, h httpkit.Handler) {
	fullPath, err := joinPaths(g.prefix, routePath)
	if err != nil {
		g.s.setBuildErr(err)
		return
	}
	g.s.handle(method, fullPath, h, g.mws)
}

func (s *Server) handle(method, routePath string, h httpkit.Handler, local []Middleware) {
	if h == nil {
		s.setBuildErr(errNilHandler)
		return
	}
	if err := validatePath(routePath); err != nil {
		s.setBuildErr(err)
		return
	}

	httpHandler := http.Handler(httpkit.Adapt(h))
	if len(local) > 0 {
		httpHandler = applyMiddleware(httpHandler, local)
	}
	if err := s.safeHandle(method, routePath, httpHandler); err != nil {
		s.setBuildErr(err)
	}
}

func (s *Server) safeHandle(method, routePath string, h http.Handler) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Join(errInvalidRoute, panicCauseErr(rec))
		}
	}()
	s.r.Handle(method, routePath, h)
	return nil
}

func panicCauseErr(rec any) error {
	switch v := rec.(type) {
	case error:
		return v
	case string:
		return errors.New(v)
	default:
		return errors.New("panic")
	}
}

func (s *Server) setBuildErr(err error) {
	if err == nil || s.buildErr != nil {
		return
	}
	s.buildErr = err
}

func (s *Server) applyPreset(preset Preset) {
	switch preset {
	case PresetDefault:
		s.Use(middleware.Recover, middleware.RequestID, middleware.Timeout(5*time.Second))
	case PresetMinimal:
		return
	default:
		return
	}
}

func applyMiddleware(h http.Handler, mws []Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func validatePrefix(prefix string) error {
	if prefix == "" || !strings.HasPrefix(prefix, "/") {
		return errInvalidPrefix
	}
	if prefix != "/" && strings.HasSuffix(prefix, "/") {
		return errTrailingPrefix
	}
	return nil
}

func validatePath(routePath string) error {
	if routePath == "" || !strings.HasPrefix(routePath, "/") {
		return errInvalidPath
	}
	return nil
}

func joinPaths(prefix, routePath string) (string, error) {
	if err := validatePrefix(prefix); err != nil {
		return "", err
	}
	if err := validatePath(routePath); err != nil {
		return "", err
	}
	if prefix == "/" {
		return routePath, nil
	}
	if routePath == "/" {
		return prefix + "/", nil
	}
	return prefix + routePath, nil
}
