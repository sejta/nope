package router

import "net/http"

type node struct {
	static    map[string]*node
	param     *node
	paramName string
	wildcard  *node
	wcName    string
	handlers  map[string]http.Handler
}

func newNode() *node {
	return &node{
		static:   map[string]*node{},
		handlers: map[string]http.Handler{},
	}
}
