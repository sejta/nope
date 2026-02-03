package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	apperrors "github.com/sejta/nope/errors"
	"github.com/sejta/nope/httpkit"
	"github.com/sejta/nope/json"
	"github.com/sejta/nope/router"
)

const (
	codeValidationFailed = "validation_failed"
	msgValidationFailed  = "validation failed"
	codeRequestTimeout   = "request_timeout"
	msgRequestTimeout    = "request timeout"
)

func apiRouter() http.Handler {
	r := router.New()
	r.POST("/posts", httpkit.Adapt(handleCreatePost))
	r.GET("/posts/:id", httpkit.Adapt(handleGetPost))
	r.GET("/slow", httpkit.Adapt(handleSlow))
	return r
}

func handleCreatePost(ctx context.Context, r *http.Request) (any, error) {
	var req CreatePostRequest
	if err := json.DecodeJSON(r, &req); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Title) == "" {
		app := apperrors.E(http.StatusBadRequest, codeValidationFailed, msgValidationFailed)
		return nil, apperrors.WithField(app, "title", "required")
	}

	resp := PostResponse{ID: "1", Title: req.Title}
	return httpkit.Created(resp), nil
}

func handleGetPost(ctx context.Context, r *http.Request) (any, error) {
	id := router.Param(r, "id")
	resp := PostResponse{ID: id, Title: "demo"}
	return resp, nil
}

func handleSlow(ctx context.Context, r *http.Request) (any, error) {
	select {
	case <-time.After(3 * time.Second):
		return map[string]bool{"ok": true}, nil
	case <-ctx.Done():
		return nil, apperrors.E(http.StatusRequestTimeout, codeRequestTimeout, msgRequestTimeout)
	}
}
