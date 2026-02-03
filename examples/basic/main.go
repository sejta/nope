package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sejta/nope/httpkit"
	"github.com/sejta/nope/httpkit/middleware"
	"github.com/sejta/nope/router"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	startedAt := time.Now()

	r := router.New()
	r.GET("/healthz", httpkit.Adapt(func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]string{"status": "ok"}, nil
	}))
	r.GET("/ping", httpkit.Adapt(func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"pong": true}, nil
	}))
	r.Mount("/api", apiRouter())
	r.Mount("/admin", adminRouter(startedAt))

	h := http.Handler(r)
	h = middleware.Recover(h)
	h = middleware.RequestID(h)
	h = middleware.Timeout(2 * time.Second)(h)
	h = middleware.AccessLog(logger)(h)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: h,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Printf("server error: %v", err)
			return
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("shutdown error: %v", err)
	}
}
