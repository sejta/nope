package router

import (
	"net/http"
	"strings"
)

type mount struct {
	prefix  string
	handler http.Handler
}

func matchMount(mounts []mount, path string) (mount, string, bool) {
	var best mount
	bestLen := -1
	var rest string
	for _, m := range mounts {
		ok, next := matchPrefix(path, m.prefix)
		if !ok {
			continue
		}
		if len(m.prefix) > bestLen {
			bestLen = len(m.prefix)
			best = m
			rest = next
		}
	}
	if bestLen == -1 {
		return mount{}, "", false
	}
	return best, rest, true
}

func matchPrefix(path, prefix string) (bool, string) {
	if prefix == "/" {
		if strings.HasPrefix(path, "/") {
			return true, path
		}
		return false, ""
	}
	if !strings.HasPrefix(path, prefix) {
		return false, ""
	}
	if len(path) == len(prefix) {
		return true, "/"
	}
	if path[len(prefix)] != '/' {
		return false, ""
	}
	return true, path[len(prefix):]
}
