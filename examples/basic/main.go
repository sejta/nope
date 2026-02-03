package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/sejta/nope/app"
	"github.com/sejta/nope/httpkit"
	"github.com/sejta/nope/httpkit/middleware"
	"github.com/sejta/nope/router"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	startedAt := time.Now()

	r := router.New()
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
	h = app.WithHealth(h)

	cfg := app.DefaultConfig()
	cfg.Addr = ":8080"

	if err := app.Run(context.Background(), cfg, h); err != nil {
		logger.Printf("server error: %v", err)
	}
}
