package router

import "strings"

func splitPath(path string) []string {
	if path == "/" {
		return nil
	}
	path = strings.TrimPrefix(path, "/")
	return strings.Split(path, "/")
}

func matchPath(root *node, path string) (*node, []RouteParam, bool) {
	segments := splitPath(path)
	if root == nil {
		return nil, nil, false
	}

	params := make([]RouteParam, 0)
	n, ok := matchFrom(root, segments, 0, &params)
	if !ok {
		return nil, nil, false
	}
	return n, params, true
}

func matchFrom(cur *node, segments []string, idx int, params *[]RouteParam) (*node, bool) {
	if cur == nil {
		return nil, false
	}
	if idx == len(segments) {
		if cur.wildcard != nil {
			*params = append(*params, RouteParam{Key: cur.wcName, Value: ""})
			return cur.wildcard, true
		}
		return cur, true
	}

	seg := segments[idx]
	if next, ok := cur.static[seg]; ok {
		if n, ok := matchFrom(next, segments, idx+1, params); ok {
			return n, true
		}
	}

	if cur.param != nil && seg != "" {
		*params = append(*params, RouteParam{Key: cur.param.paramName, Value: seg})
		if n, ok := matchFrom(cur.param, segments, idx+1, params); ok {
			return n, true
		}
		*params = (*params)[:len(*params)-1]
	}

	if cur.wildcard != nil {
		tail := strings.Join(segments[idx:], "/")
		*params = append(*params, RouteParam{Key: cur.wcName, Value: tail})
		return cur.wildcard, true
	}

	return nil, false
}
