package main

import (
	"context"
	"net/http"
	"time"

	"github.com/sejta/nope/server"
)

func main() {
	srv := server.NewWithPreset(":8080", server.PresetDefault)
	srv.EnableHealth()

	srv.GET("/ping", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]bool{"pong": true}, nil
	})

	api := srv.Group("/api")
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-API", "1")
			next.ServeHTTP(w, r)
		})
	})
	api.GET("/time", func(ctx context.Context, r *http.Request) (any, error) {
		return map[string]string{"now": time.Now().UTC().Format(time.RFC3339)}, nil
	})

	_ = srv.Run()
}
