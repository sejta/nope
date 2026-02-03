package app

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Run запускает HTTP-сервер и блокируется до остановки.
// Сервер останавливается по отмене ctx или по SIGINT/SIGTERM.
func Run(ctx context.Context, cfg Config, h http.Handler) error {
	if h == nil {
		panic("app: handler is nil")
	}
	cfg = withDefaults(cfg)

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}

	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	return runWithListener(sigCtx, cfg, h, ln)
}

func runWithListener(ctx context.Context, cfg Config, h http.Handler, ln net.Listener) error {
	cfg = withDefaults(cfg)

	srv := &http.Server{
		Handler:           h,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		if shutdownCtx.Err() == context.DeadlineExceeded {
			_ = srv.Close()
		}
	}

	err := <-errCh
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
