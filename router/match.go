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
	cur := root
	if cur == nil {
		return nil, nil, false
	}

	params := make([]RouteParam, 0)
	for _, seg := range segments {
		if next, ok := cur.static[seg]; ok {
			cur = next
			continue
		}
		if cur.param != nil && seg != "" {
			params = append(params, RouteParam{Key: cur.param.paramName, Value: seg})
			cur = cur.param
			continue
		}
		return nil, nil, false
	}
	return cur, params, true
}
